// Package msg
package msg

import (
	"context"
	"errors"
	"fmt"
	"os"
	"runtime"
	"sync"
	"sync/atomic"
	"time"

	"github.com/ThreeDotsLabs/watermill/message"
	"github.com/json-iterator/go"
	"github.com/patrickmn/go-cache"
	"github.com/streadway/amqp"
	"go.uber.org/zap"

	"template/pkg/async"
	"template/pkg/logger"
	"template/pkg/validator"
)

func NewWriter(ctx context.Context, opts ...func(*WriterOption)) *Writer {
	w := &Writer{
		cache: cache.New(30*time.Minute, time.Minute),
		WriterOption: WriterOption{
			ServiceNameFunc: ServiceName,
			CallerFunc:      CallerFunc,
			MetadataPool:    NewMetadataPool(),
			NowFunc: func() time.Time {
				return time.Now().Local()
			},
			Queue:       NewQueue(1024 * 10),
			Cfg:         NewCfgHandler(),
			JSONHandler: jsoniter.ConfigCompatibleWithStandardLibrary,
			Plugins:     make([]Plugin, 0, 2),
			Validator:   validator.NewValidator(),
			Marshal:     async.DefaultMarshal{},
		},
		flushEnable: 1,
	}

	for _, o := range opts {
		o(&w.WriterOption)
	}
	ctx, w.cancel = context.WithCancel(ctx)
	w.wg.Add(1)
	go w.handle(ctx)

	runtime.SetFinalizer(w, func(w *Writer) {
		if err := w.Close(); err != nil {
			_, _ = os.Stderr.WriteString(err.Error())
		}
	})
	return w
}

type WriterOption struct {
	ServiceNameFunc ServiceNameHandler
	CallerFunc      CallerHandler
	MetadataPool    MetadataPool
	NowFunc         NowHandler
	Queue           Queue
	Cfg             CfgHandler
	JSONHandler     jsoniter.API
	Plugins         []Plugin
	Validator       validator.Validator
	Publisher       async.Producer
	Channel         async.Channel
	Marshal         async.MarshalAPI
	TraceID         func(context.Context) string
	RequestID       func(context.Context) string
	AccountID       func(context.Context) string
	UserID          func(context.Context) string
}

type Writer struct {
	WriterOption
	cache       *cache.Cache
	cancel      context.CancelFunc
	mux         sync.RWMutex
	wg          sync.WaitGroup
	flushEnable uint32
}

func (w *Writer) Error(ctx context.Context, skip int, err error, msg string) {
	if !w.Cfg.Enable() {
		return
	}
	// get metadata
	meta := w.MetadataPool.Get()
	meta.TraceID = w.TraceID(ctx)
	meta.ServiceName = w.ServiceNameFunc()
	meta.Locate = w.CallerFunc(skip)
	meta.RequestID = w.RequestID(ctx)
	meta.AccountID = w.AccountID(ctx)
	meta.UserID = w.UserID(ctx)
	if err != nil {
		meta.Summary = fmt.Sprintf("%+v", err)
	} else {
		meta.Summary = "null"
	}
	if msg != "" {
		//meta.Desc = wlog.PwdReplacerReplaceStr(msg)
		meta.Desc = msg
	}
	meta.ErrorTime = w.NowFunc()

	w.Write(ctx, meta)
}

// MergeInfo 根据trace_id为错误的消息增加一条数据
func (w *Writer) MergeInfo(ctx context.Context, skip int, msg string) {
	if !w.Cfg.Enable() {
		return
	}
	// get metadata
	meta := w.MetadataPool.Get()
	meta.TraceID = w.TraceID(ctx)
	meta.ServiceName = w.ServiceNameFunc()
	meta.Locate = w.CallerFunc(skip)
	meta.RequestID = w.RequestID(ctx)
	meta.AccountID = w.AccountID(ctx)
	meta.UserID = w.UserID(ctx)
	meta.Summary = "merge info"
	meta.Desc = msg
	meta.ErrorTime = w.NowFunc()

	if err := w.Validator.ValidateStruct(meta); err != nil {
		logger.From(ctx).Error(fmt.Sprintf("err:%v %v\n", err, meta))
		w.MetadataPool.Put(meta)
		return
	}
	w.cache.SetDefault(meta.TraceID, meta)
}

func (w *Writer) Write(ctx context.Context, meta *Metadata) {
	if !w.Cfg.Enable() {
		return
	}

	if err := w.Validator.ValidateStruct(meta); err != nil {
		logger.From(ctx).Error(fmt.Sprintf("err:%v %v\n", err, meta))
		w.MetadataPool.Put(meta)
		return
	}
	// push metadata
	w.wg.Add(1)
	go func() {
		_, found := w.cache.Get(meta.TraceID + meta.Locate)
		if found {
			w.wg.Done()
			return
		}
		w.Queue.Write(meta)
		// 设置错误记录已经在改行有记录了
		w.cache.SetDefault(meta.TraceID+meta.Locate, nil)
		// 根据trace_id获取缓存的curl记录，有的话追加到队列里，并删除记录
		var value interface{}
		if value, found = w.cache.Get(meta.TraceID); found {
			if curlMeta, ok := value.(*Metadata); ok {
				w.Queue.Write(curlMeta)
				w.cache.Delete(meta.TraceID)
			}
		}
		w.wg.Done()
	}()
}

func (w *Writer) handle(ctx context.Context) {
	defer w.wg.Done()
	var (
		t            = time.NewTimer(5 * time.Minute)
		flag         Status
		metadataList = make([]*Metadata, 0, w.Cfg.Limit())
		enableChan   = make(chan struct{})
	)
	w.wg.Add(1)
	go w.check(ctx, enableChan)
	for {
		w.Cfg.OUtEnable().CheckChangeClose(func() {
			select {
			case <-enableChan:
			case <-ctx.Done():
			}
		})
		select {
		case data := <-w.Queue.Read():
			if flag.NotHasStatus(ResetTime) { // 没有进行重置时间
				t.Reset(w.Cfg.Timeout())
				flag.AddStatus(ResetTime) // 写入已经进行重置时间的标识
			}
			for _, plugin := range w.Plugins { // 插件使用
				plugin.Use(data)
			}
			metadataList = append(metadataList, data)
			if len(metadataList) >= int(w.Cfg.Limit()) {
				break
			}
			continue
		case <-ctx.Done():
			flag.AddStatus(Exit) // 写入希望退出的标识
			if len(metadataList) == 0 {
				t.Stop()
				return
			}
			metadataList = append(metadataList, w.Queue.ListAndClear()...)
		case <-t.C:
			if len(metadataList) == 0 {
				flag.DeleteStatus(ResetTime) // 删除已经进行重置时间的标识
				continue
			}
		}
		result := w.putAndTransfer(ctx, metadataList)
		metadataList = metadataList[0:0]

		atomic.StoreUint32(&w.flushEnable, 0)
		if err := w.GetPublisher().Publish(ctx, w.Channel, w.Cfg.Exchange(), w.Cfg.RoutingKey(), result); err != nil {
			logger.From(ctx).Error("", zap.Error(err))
		}
		atomic.StoreUint32(&w.flushEnable, 1)

		flag.DeleteStatus(ResetTime) // 删除已经进行重置时间的标识
		if flag.HasStatus(Exit) {    // 判断是否退出
			t.Stop()
			return
		}
	}
}

func (w *Writer) putAndTransfer(ctx context.Context, metadataList []*Metadata) []amqp.Publishing {
	result := make([]amqp.Publishing, 0, len(metadataList))
	for _, metadata := range metadataList {
		data, err := w.JSONHandler.Marshal(metadata)
		w.MetadataPool.Put(metadata)
		w.cache.Delete(metadata.TraceID)
		if err != nil {
			logger.From(ctx).Warn("", zap.Error(err))
			continue
		}
		var value amqp.Publishing
		if value, err = w.Marshal.Marshal(message.NewMessage(metadata.TraceID, data)); err != nil {
			logger.From(ctx).Warn("", zap.Error(err))
			continue
		}
		result = append(result, value)
	}
	return result
}

func (w *Writer) SetPublisher(publisher async.Producer) error {
	if atomic.LoadUint32(&w.flushEnable) == 1 {
		w.mux.Lock()
		w.Publisher = publisher
		w.mux.Unlock()
		return nil
	}
	var index int
	t := time.NewTicker(5 * time.Second)
	for range t.C {
		index++
		if atomic.LoadUint32(&w.flushEnable) == 1 {
			w.mux.Lock()
			w.Publisher = publisher
			w.mux.Unlock()
			t.Stop()
			return nil
		}
		if index > 6 {
			t.Stop()
			break
		}
	}
	return errors.New("set publisher time out")
}

func (w *Writer) SetChannel(channel async.Channel) error {
	if atomic.LoadUint32(&w.flushEnable) == 1 {
		w.mux.Lock()
		w.Channel = channel
		w.mux.Unlock()
		return nil
	}
	var index int
	t := time.NewTicker(5 * time.Second)
	for range t.C {
		index++
		if atomic.LoadUint32(&w.flushEnable) == 1 {
			w.mux.Lock()
			w.Channel = channel
			w.mux.Unlock()
			t.Stop()
			return nil
		}
		if index > 6 {
			t.Stop()
			break
		}
	}
	return errors.New("set publisher time out")
}

func (w *Writer) GetPublisher() async.Producer {
	w.mux.RLock()
	publisher := w.Publisher
	w.mux.RUnlock()
	return publisher
}

func (w *Writer) Close() error {
	w.cancel()
	w.wg.Wait()
	return w.GetPublisher().Close()
}

func (w *Writer) check(ctx context.Context, enableCh chan struct{}) {
	ticker := time.NewTicker(time.Second)
	for {
		select {
		case <-ctx.Done():
			ticker.Stop()
			w.wg.Done()
			return
		case <-ticker.C:
			w.Cfg.OUtEnable().CheckChangeOpen(func() {
				select {
				case enableCh <- struct{}{}:
				case <-ctx.Done():
				}
			})
		}
	}
}
