package metric

import (
	"sort"
	"sync/atomic"
	"time"

	"template/pkg/metric/model"
)

func NewMaxGauge(name string, labels []*model.Label) Gauge {
	sort.Sort(model.Labels(labels))
	return &maxGauge{
		name:   name,
		value:  0,
		labels: labels,
	}
}

type maxGauge struct {
	name   string
	value  uint64
	labels []*model.Label
}

func (m *maxGauge) Collect(metrics chan<- Writer) {
	metrics <- m
}

func (m *maxGauge) Set(u uint64) {
	atomic.StoreUint64(&m.value, u)
}

func (m *maxGauge) Write(out *model.Metric) error {
	out.Day = day(time.Now())
	out.Labels = make([]*model.Label, len(m.labels))
	copy(out.Labels, m.labels)
	out.MaxGauge = &model.MaxGauge{
		Value: atomic.LoadUint64(&m.value),
	}
	return nil
}

func (m *maxGauge) With(labels []*model.Label) Gauge {
	s := make([]*model.Label, len(m.labels))
	copy(s, m.labels)
	return NewMaxGauge(m.name, model.MergeLabels(s, labels))
}

func NewMinGauge(name string, labels []*model.Label) Gauge {
	sort.Sort(model.Labels(labels))
	return &minGauge{
		name:   name,
		value:  0,
		labels: labels,
	}
}

type minGauge struct {
	name   string
	value  uint64
	labels []*model.Label
}

func (m *minGauge) Collect(metrics chan<- Writer) {
	metrics <- m
}

func (m *minGauge) Set(u uint64) {
	atomic.StoreUint64(&m.value, u)
}

func (m *minGauge) Write(out *model.Metric) error {
	out.Day = day(time.Now())
	out.Labels = make([]*model.Label, len(m.labels))
	copy(out.Labels, m.labels)
	out.MinGauge = &model.MinGauge{
		Value: atomic.LoadUint64(&m.value),
	}
	return nil
}

func (m *minGauge) With(labels []*model.Label) Gauge {
	s := make([]*model.Label, len(m.labels))
	copy(s, m.labels)
	return NewMinGauge(m.name, model.MergeLabels(s, labels))
}
