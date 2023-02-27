package metric

import (
	"sort"
	"sync/atomic"
	"time"

	"template/pkg/metric/model"
)

type Counter interface {
	Collector
	Writer
	With(labels []*model.Label) Counter
	Set(uint64)
	Inc()
}

func NewCounter(name string, labels []*model.Label) Counter {
	sort.Sort(model.Labels(labels))
	return &counter{
		name:   name,
		value:  0,
		labels: labels,
	}
}

type counter struct {
	name   string
	value  uint64
	labels []*model.Label
}

func (c *counter) With(labels []*model.Label) Counter {
	s := make([]*model.Label, len(c.labels))
	copy(s, c.labels)
	return NewCounter(c.name, model.MergeLabels(s, labels))
}

func (c *counter) Collect(metrics chan<- Writer) {
	metrics <- c
}

func (c *counter) Write(out *model.Metric) error {
	out.Day = day(time.Now())
	out.Labels = make([]*model.Label, len(c.labels))
	copy(out.Labels, c.labels)
	out.Counter = &model.Counter{
		Value: atomic.LoadUint64(&c.value),
	}
	return nil
}

func (c *counter) Set(v uint64) {
	atomic.StoreUint64(&c.value, v)
}

func (c *counter) Inc() {
	atomic.AddUint64(&c.value, 1)
}
