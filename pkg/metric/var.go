package metric

import (
	"context"

	"template/pkg/metric/model"
)

var (
	Monitor           *monitor
	Chan              = make(chan Writer, 1024*1024)
	LabelRequestCount = []*model.Label{
		{
			Name:  "api",
			Value: "request_count",
		},
	}
	LabelRequestTotalLatency = []*model.Label{
		{
			Name:  "api",
			Value: "request_average_latency",
		},
	}
	LabelMaxLatency = []*model.Label{
		{
			Name:  "api",
			Value: "request_max_latency",
		},
	}
	LabelMinLatency = []*model.Label{
		{
			Name:  "api",
			Value: "request_min_latency",
		},
	}
)

var (
	CounterRequestCount = NewCounter("counter", LabelRequestCount)
	CounterLatency      = NewCounter("counterLatency", LabelRequestTotalLatency)
	MaxGauge            = NewMaxGauge("maxGauge", LabelMaxLatency)
	MinGauge            = NewMinGauge("minGauge", LabelMinLatency)
)

const (
	GoroutineNums = 3
	BucketNums    = 7
)

const (
	notTargetDataIndex = -1
	allDayIndex        = -1
)

// Run runs then runtime of metric
func Run(ctx context.Context) error {
	Monitor = New(ctx)
	Monitor.Run(Chan)
	return nil
}
