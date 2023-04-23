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
			Name:  ApiLabelName,
			Value: "request_count",
		},
	}
	LabelRequestTotalLatency = []*model.Label{
		{
			Name:  ApiLabelName,
			Value: "request_total_latency",
		},
	}
	LabelMaxLatency = []*model.Label{
		{
			Name:  ApiLabelName,
			Value: "request_max_latency",
		},
	}
	LabelMinLatency = []*model.Label{
		{
			Name:  ApiLabelName,
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

const (
	RequestAverageLatency = "request_average_latency"
	ApiLabelName          = "api"
	PathLabelName         = "path"
)

// Run runs then runtime of metric
func Run(ctx context.Context) error {
	Monitor = New(ctx)
	Monitor.Run(Chan)
	return nil
}
