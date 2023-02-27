package metric

import (
	"time"

	"template/pkg/metric/model"
)

const (
	YMDFormat = "2006-01-02"
)

// Collector 收集器
type Collector interface {
	Collect(metrics chan<- Writer)
}

// Writer 格式化到 Metrics
type Writer interface {
	Write(out *model.Metric) error
}

// Handler 中间件处理器
type Handler interface {
	Handle(metric *model.Metric) *model.Metric
}

type NopHandler struct{}

func (n NopHandler) Handle(metric *model.Metric) *model.Metric {
	return metric
}

func day(t time.Time) int64 {
	return t.UTC().Unix() / (60 * 60 * 24)
}

// Push put metrics to channel
func Push(c Collector, metrics chan<- Writer) {
	go c.Collect(metrics)
}
