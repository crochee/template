package metric

import (
	"sort"
	"sync/atomic"
	"time"

	"template/pkg/metric/model"
)

type Gauge interface {
	Collector
	Writer
	With(labels []*model.Label) Gauge
	Set(uint64)
}

func NewGauge(name string, labels []*model.Label) Gauge {
	sort.Sort(model.Labels(labels))
	return &gauge{
		name:   name,
		value:  0,
		labels: labels,
	}
}

type gauge struct {
	name   string
	value  uint64
	labels []*model.Label
}

func (g *gauge) With(labels []*model.Label) Gauge {
	s := make([]*model.Label, len(g.labels))
	copy(s, g.labels)
	return NewGauge(g.name, model.MergeLabels(s, labels))
}

func (g *gauge) Collect(metrics chan<- Writer) {
	metrics <- g
}

func (g *gauge) Write(out *model.Metric) error {
	out.Day = day(time.Now())
	out.Labels = make([]*model.Label, len(g.labels))
	copy(out.Labels, g.labels)
	out.Gauge = &model.Gauge{
		Value: atomic.LoadUint64(&g.value),
	}
	return nil
}

func (g *gauge) Set(u uint64) {
	atomic.StoreUint64(&g.value, u)
}
