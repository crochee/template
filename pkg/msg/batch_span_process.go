package msg

import (
	"context"
	"runtime"
	"sync"
	"sync/atomic"
	"time"

	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/trace"

	"template/pkg/logger/gormx"
)

// Defaults for BatchSpanProcessorOptions.
const (
	DefaultMaxQueueSize       = 2048
	DefaultScheduleDelay      = 5000
	DefaultExportTimeout      = 30000
	DefaultMaxExportBatchSize = 512
)

// BatchSpanProcessorOption configures a BatchSpanProcessor.
type BatchSpanProcessorOption func(o *BatchSpanProcessorOptions)

// BatchSpanProcessorOptions is configuration settings for a
// BatchSpanProcessor.
type BatchSpanProcessorOptions struct {
	// MaxQueueSize is the maximum queue size to buffer spans for delayed processing. If the
	// queue gets full it drops the spans. Use BlockOnQueueFull to change this behavior.
	// The default value of MaxQueueSize is 2048.
	MaxQueueSize int

	// BatchTimeout is the maximum duration for constructing a batch. Processor
	// forcefully sends available spans when timeout is reached.
	// The default value of BatchTimeout is 5000 msec.
	BatchTimeout time.Duration

	// MaxExportBatchSize is the maximum number of spans to process in a single batch.
	// If there are more than one batch worth of spans then it processes multiple batches
	// of spans one batch after the other without any delay.
	// The default value of MaxExportBatchSize is 512.
	MaxExportBatchSize int

	// from get logger handle
	from func(context.Context) gormx.Logger
}

// BatchSpanProcessor is a SpanProcessor that batches asynchronously-received
// spans and sends them to a trace.Exporter when complete.
type BatchSpanProcessor struct {
	e sdktrace.SpanExporter
	o BatchSpanProcessorOptions

	queue   chan sdktrace.ReadOnlySpan
	dropped uint32

	batch      []sdktrace.ReadOnlySpan
	batchMutex sync.Mutex
	timer      *time.Timer
	stopWait   sync.WaitGroup
	stopOnce   sync.Once
	stopCh     chan struct{}
}

// NewBatchSpanProcessor creates a new SpanProcessor that will send completed
// span batches to the exporter with the supplied options.
//
// If the exporter is nil, the span processor will preform no action.
func NewBatchSpanProcessor(exporter sdktrace.SpanExporter, options ...BatchSpanProcessorOption) *BatchSpanProcessor {
	o := BatchSpanProcessorOptions{
		MaxQueueSize:       4096,
		BatchTimeout:       5 * time.Second,
		MaxExportBatchSize: 512,
		from:               gormx.Nop,
	}
	for _, opt := range options {
		opt(&o)
	}
	bsp := &BatchSpanProcessor{
		e:      exporter,
		o:      o,
		batch:  make([]sdktrace.ReadOnlySpan, 0, o.MaxExportBatchSize),
		timer:  time.NewTimer(o.BatchTimeout),
		queue:  make(chan sdktrace.ReadOnlySpan, o.MaxQueueSize),
		stopCh: make(chan struct{}),
	}

	bsp.stopWait.Add(1)
	go func() {
		defer bsp.stopWait.Done()
		bsp.processQueue()
		bsp.drainQueue()
	}()

	return bsp
}

func (bsp *BatchSpanProcessor) GetOptions() BatchSpanProcessorOptions {
	return bsp.o
}

func (bsp *BatchSpanProcessor) Len() int {
	return len(bsp.queue)
}

// OnStart method does nothing.
func (bsp *BatchSpanProcessor) OnStart(parent context.Context, s sdktrace.ReadWriteSpan) {}

// OnEnd method enqueues a ReadOnlySpan for later processing.
func (bsp *BatchSpanProcessor) OnEnd(s sdktrace.ReadOnlySpan) {
	// Do not enqueue spans if we are just going to drop them.
	if bsp.e == nil {
		bsp.o.from(context.TODO()).Warnf("Do not enqueue spans if we are just going to drop them")
		return
	}
	bsp.enqueueDrop(context.TODO(), s)
}

// Shutdown flushes the queue and waits until all spans are processed.
// It only executes once. Subsequent call does nothing.
func (bsp *BatchSpanProcessor) Shutdown(ctx context.Context) error {
	var err error
	bsp.stopOnce.Do(func() {
		wait := make(chan struct{})
		go func() {
			close(bsp.stopCh)
			bsp.stopWait.Wait()
			if bsp.e != nil {
				if err := bsp.e.Shutdown(ctx); err != nil {
					bsp.o.from(ctx).Errorf("failed to shutdown span exporter: %v", err)
				}
			}
			close(wait)
		}()
		// Wait until the wait group is done or the context is cancelled
		select {
		case <-wait:
		case <-ctx.Done():
			err = ctx.Err()
		}
	})
	return err
}

type forceFlushSpan struct {
	sdktrace.ReadOnlySpan
	flushed chan struct{}
}

func (f forceFlushSpan) SpanContext() trace.SpanContext {
	return trace.NewSpanContext(trace.SpanContextConfig{TraceFlags: trace.FlagsSampled})
}

// ForceFlush exports all ended spans that have not yet been exported.
func (bsp *BatchSpanProcessor) ForceFlush(ctx context.Context) error {
	var err error
	if bsp.e != nil {
		flushCh := make(chan struct{})
		if bsp.enqueueBlockOnQueueFull(ctx, forceFlushSpan{flushed: flushCh}) {
			select {
			case <-flushCh:
				// Processed any items in queue prior to ForceFlush being called
			case <-ctx.Done():
				return ctx.Err()
			}
		}

		wait := make(chan error)
		go func() {
			wait <- bsp.exportSpans(ctx)
			close(wait)
		}()
		// Wait until the export is finished or the context is cancelled/timed out
		select {
		case err = <-wait:
		case <-ctx.Done():
			err = ctx.Err()
		}
	}
	return err
}

// WithMaxQueueSize returns a BatchSpanProcessorOption that configures the
// maximum queue size allowed for a BatchSpanProcessor.
func WithMaxQueueSize(size int) BatchSpanProcessorOption {
	return func(o *BatchSpanProcessorOptions) {
		o.MaxQueueSize = size
	}
}

// WithMaxExportBatchSize returns a BatchSpanProcessorOption that configures
// the maximum export batch size allowed for a BatchSpanProcessor.
func WithMaxExportBatchSize(size int) BatchSpanProcessorOption {
	return func(o *BatchSpanProcessorOptions) {
		o.MaxExportBatchSize = size
	}
}

// WithBatchTimeout returns a BatchSpanProcessorOption that configures the
// maximum delay allowed for a BatchSpanProcessor before it will export any
// held span (whether the queue is full or not).
func WithBatchTimeout(delay time.Duration) BatchSpanProcessorOption {
	return func(o *BatchSpanProcessorOptions) {
		o.BatchTimeout = delay
	}
}

func WithLoggerFrom(from func(context.Context) gormx.Logger) BatchSpanProcessorOption {
	return func(o *BatchSpanProcessorOptions) {
		o.from = from
	}
}

// exportSpans is a subroutine of processing and draining the queue.
func (bsp *BatchSpanProcessor) exportSpans(ctx context.Context) error {
	bsp.timer.Reset(bsp.o.BatchTimeout)

	bsp.batchMutex.Lock()
	defer bsp.batchMutex.Unlock()

	if l := len(bsp.batch); l > 0 {
		bsp.o.from(ctx).Debugf("exporting spans count %d total_dropped %d", len(bsp.batch), atomic.LoadUint32(&bsp.dropped))
		err := bsp.e.ExportSpans(ctx, bsp.batch)

		// A new batch is always created after exporting, even if the batch failed to be exported.
		//
		// It is up to the exporter to implement any type of retry logic if a batch is failing
		// to be exported, since it is specific to the protocol and backend being sent to.
		bsp.batch = bsp.batch[:0]

		if err != nil {
			return err
		}
	}
	return nil
}

// processQueue removes spans from the `queue` channel until processor
// is shut down. It calls the exporter in batches of up to MaxExportBatchSize
// waiting up to BatchTimeout to form a batch.
func (bsp *BatchSpanProcessor) processQueue() {
	defer bsp.timer.Stop()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	for {
		select {
		case <-bsp.stopCh:
			bsp.o.from(ctx).Warnf("received stop ch")
			return
		case <-bsp.timer.C:
			if err := bsp.exportSpans(ctx); err != nil {
				bsp.o.from(ctx).Errorf("failed to export spans: %v", err)
			}
		case sd := <-bsp.queue:
			if ffs, ok := sd.(forceFlushSpan); ok {
				close(ffs.flushed)
				continue
			}
			bsp.batchMutex.Lock()
			bsp.batch = append(bsp.batch, sd)
			shouldExport := len(bsp.batch) >= bsp.o.MaxExportBatchSize
			bsp.batchMutex.Unlock()
			if shouldExport {
				if !bsp.timer.Stop() {
					<-bsp.timer.C
				}
				if err := bsp.exportSpans(ctx); err != nil {
					bsp.o.from(ctx).Errorf("failed to export spans: %v", err)
				}
			}
		}
	}
}

// drainQueue awaits the any caller that had added to bsp.stopWait
// to finish the enqueue, then exports the final batch.
func (bsp *BatchSpanProcessor) drainQueue() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	for {
		select {
		case sd := <-bsp.queue:
			if sd == nil {
				bsp.o.from(ctx).Infof("received nil span, stopping queue")
				if err := bsp.exportSpans(ctx); err != nil {
					bsp.o.from(ctx).Errorf("failed to export spans: %v", err)
				}
				return
			}

			bsp.batchMutex.Lock()
			bsp.batch = append(bsp.batch, sd)
			shouldExport := len(bsp.batch) == bsp.o.MaxExportBatchSize
			bsp.batchMutex.Unlock()

			if shouldExport {
				if err := bsp.exportSpans(ctx); err != nil {
					bsp.o.from(ctx).Errorf("failed to export spans: %v", err)
				}
			}
		default:
			close(bsp.queue)
		}
	}
}

func recoverSendOnClosedChan() {
	x := recover()
	switch err := x.(type) {
	case nil:
		return
	case runtime.Error:
		if err.Error() == "send on closed channel" {
			return
		}
	}
	panic(x)
}

func (bsp *BatchSpanProcessor) enqueueBlockOnQueueFull(ctx context.Context, sd sdktrace.ReadOnlySpan) bool {
	if !sd.SpanContext().IsSampled() {
		return false
	}

	// This ensures the bsp.queue<- below does not panic as the
	// processor shuts down.
	defer recoverSendOnClosedChan()

	select {
	case <-bsp.stopCh:
		return false
	default:
	}

	select {
	case bsp.queue <- sd:
		return true
	case <-ctx.Done():
		return false
	}
}

func (bsp *BatchSpanProcessor) enqueueDrop(ctx context.Context, sd sdktrace.ReadOnlySpan) bool {
	if !sd.SpanContext().IsSampled() {
		return false
	}

	// This ensures the bsp.queue<- below does not panic as the
	// processor shuts down.
	defer recoverSendOnClosedChan()

	select {
	case <-bsp.stopCh:
		bsp.o.from(ctx).Warnf("received stop ch, not enqueueing span %+v", sd)
		return false
	default:
	}

	select {
	case bsp.queue <- sd:
		return true
	default:
		bsp.o.from(ctx).Warnf("dropping span %+v", sd)
		atomic.AddUint32(&bsp.dropped, 1)
	}
	return false
}

// MarshalLog is the marshaling function used by the logging system to represent this exporter.
func (bsp *BatchSpanProcessor) MarshalLog() interface{} {
	return struct {
		Type         string
		SpanExporter sdktrace.SpanExporter
		Config       BatchSpanProcessorOptions
	}{
		Type:         "BatchSpanProcessor",
		SpanExporter: bsp.e,
		Config:       bsp.o,
	}
}
