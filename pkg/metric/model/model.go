package model

import (
	"math"
)

type MetricFamily struct {
	Name    string    `json:"name"`
	Metrics []*Metric `json:"metric"`
}

func (m *MetricFamily) Clone() *MetricFamily {
	temp := &MetricFamily{
		Name:    m.Name,
		Metrics: make([]*Metric, 0, len(m.Metrics)),
	}
	for _, data := range m.Metrics {
		temp.Metrics = append(temp.Metrics, data.Clone())
	}
	return temp
}

type Metric struct {
	Labels   []*Label  `json:"label"`
	Counter  *Counter  `json:"counter"`
	Gauge    *Gauge    `json:"gauge"`
	MaxGauge *MaxGauge `json:"max_gauge"`
	MinGauge *MinGauge `json:"min_gauge"`
	Day      int64     `json:"day"`
}

type Label struct {
	Name  string `json:"name"`
	Value string `json:"key"`
}

type Gauge struct {
	Value uint64 `json:"value"`
}

type MaxGauge struct {
	Value uint64 `json:"value"`
}

type MinGauge struct {
	Value uint64 `json:"value"`
}

type Counter struct {
	Value uint64 `json:"value"`
}

func (m *Metric) Merge(metric *Metric) {
	if metric.Counter != nil {
		// u64位溢出就重置整个m的指标数据
		if math.MaxUint64-metric.Counter.Value < m.Counter.Value {
			m.Counter = nil
			m.Gauge = nil
			m.MaxGauge = nil
			m.MinGauge = nil
		}
		if m.Counter == nil {
			m.Counter = &Counter{}
		}
		m.Counter.Value += metric.Counter.Value
	}
	if metric.Gauge != nil {
		if m.Gauge == nil {
			m.Gauge = &Gauge{}
		}
		m.Gauge.Value = metric.Gauge.Value
	}
	if metric.MaxGauge != nil {
		if m.MaxGauge == nil {
			m.MaxGauge = &MaxGauge{}
		}
		if metric.MaxGauge.Value > m.MaxGauge.Value {
			m.MaxGauge.Value = metric.MaxGauge.Value
		}
	}
	if metric.MinGauge != nil {
		if m.MinGauge == nil {
			m.MinGauge = &MinGauge{Value: math.MaxUint64}
		}
		if metric.MinGauge.Value < m.MinGauge.Value {
			m.MinGauge.Value = metric.MinGauge.Value
		}
	}
	if m.Day == 0 {
		m.Day = metric.Day
	}
}

func (m *Metric) Clone() *Metric {
	temp := &Metric{
		Labels: make([]*Label, len(m.Labels)),
	}
	copy(temp.Labels, m.Labels)
	if m.Counter != nil {
		temp.Counter = &Counter{Value: m.Counter.Value}
	}
	if m.Gauge != nil {
		temp.Gauge = &Gauge{Value: m.Gauge.Value}
	}
	if m.MaxGauge != nil {
		temp.MaxGauge = &MaxGauge{Value: m.MaxGauge.Value}
	}
	if m.MinGauge != nil {
		temp.MinGauge = &MinGauge{Value: m.MinGauge.Value}
	}
	temp.Day = m.Day
	return temp
}

type Labels []*Label

func (l Labels) Len() int {
	return len(l)
}

func (l Labels) Less(i, j int) bool {
	if l[i].Name < l[j].Name {
		return true
	}
	if l[i].Name == l[j].Name {
		return l[i].Value < l[j].Value
	}
	return false
}

func (l Labels) Swap(i, j int) {
	l[i], l[j] = l[j], l[i]
}

// MergeLabels 合并逻辑：相同name取b的节点
func MergeLabels(a, b []*Label) []*Label {
	result := make([]*Label, 0, (len(a)+len(b))*3/4)
	temp := make(map[string]struct{}, (len(a)+len(b))*3/4)
	for _, tempB := range b {
		if _, ok := temp[tempB.Name]; ok {
			continue
		}
		result = append(result, &Label{
			Name:  tempB.Name,
			Value: tempB.Value,
		})
		temp[tempB.Name] = struct{}{}
	}

	for _, tempA := range a {
		if _, ok := temp[tempA.Name]; ok {
			continue
		}
		result = append(result, &Label{
			Name:  tempA.Name,
			Value: tempA.Value,
		})
		temp[tempA.Name] = struct{}{}
	}
	return result
}

// ContainLabel a has b
func ContainLabel(a []*Label, b *Label) bool {
	for _, label := range a {
		if label.Name == "*" || label.Name == b.Name || b.Name == "*" {
			if label.Value == "*" || label.Value == b.Value || b.Value == "*" {
				return true
			}
		}
	}
	return false
}

// More a>=b
func More(a []*Label, b []*Label) bool {
	var notExist int
	for _, bValue := range b {
		if !ContainLabel(a, bValue) {
			notExist += 1
		}
	}
	// 筛选条件的label均不存在于已收集label中
	return notExist != len(b)
}
