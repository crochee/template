package metric

import (
	"fmt"
	"reflect"
	"sort"
	"strings"
	"sync"
	"time"

	"template/pkg/metric/model"
)

type dataLink struct {
	rw      sync.RWMutex
	buckets []*model.MetricFamily
	length  int
	cur     int
}

func newDataLink(length int) *dataLink {
	d := &dataLink{
		buckets: make([]*model.MetricFamily, length),
		length:  length,
		cur:     0,
	}
	for i := 0; i < length; i++ {
		d.buckets[i] = &model.MetricFamily{
			Name:    fmt.Sprintf("bucket%d", i),
			Metrics: nil,
		}
	}
	return d
}

func (d *dataLink) handle(metric *model.Metric) {
	d.rw.Lock()
	defer d.rw.Unlock()

	if len(d.buckets[d.cur].Metrics) == 0 {
		d.buckets[d.cur].Metrics = append(d.buckets[d.cur].Metrics, metric.Clone())
		return
	}
	// 匹配 labels
	var dayNotRight bool
	for _, m := range d.buckets[d.cur].Metrics {
		if m.Day == metric.Day {
			if reflect.DeepEqual(m.Labels, metric.Labels) {
				m.Merge(metric)
				return
			}
			continue
		}
		dayNotRight = true
		break
	}
	// 不在同一天，就指针向左移动
	if dayNotRight {
		// index -1
		if d.cur-1 < 0 {
			d.cur = d.length - 1
		} else {
			d.cur--
		}
		d.buckets[d.cur].Metrics = []*model.Metric{metric.Clone()}
		return
	}
	// label不同时追加
	d.buckets[d.cur].Metrics = append(d.buckets[d.cur].Metrics, metric.Clone())
}

func (d *dataLink) len() int {
	return d.length
}

// GetRealIndex 通过天数获取桶的真实索引
func (d *dataLink) getRealIndex(target int64) int {
	// 获取当前的天数
	curDay := day(time.Now())
	// 查询的天数
	searchDay := curDay - target + 1
	// 遍历查询
	d.rw.RLock()
	defer d.rw.RUnlock()

	for i, metricFamily := range d.buckets {
		if len(metricFamily.Metrics) == 0 {
			continue
		}
		if searchDay == metricFamily.Metrics[0].Day {
			return i
		}
	}
	return notTargetDataIndex
}

// getByIndex 通过索引获取数据
func (d *dataLink) getByIndex(target int64) *model.MetricFamily {
	index := d.getRealIndex(target)
	if index == notTargetDataIndex {
		return &model.MetricFamily{}
	}
	d.rw.RLock()
	tmp := d.buckets[index].Clone()
	d.rw.RUnlock()
	return tmp
}

type data struct {
	rw    sync.RWMutex
	value *model.MetricFamily
}

func newData() *data {
	return &data{
		value: &model.MetricFamily{
			Name:    "All",
			Metrics: nil,
		},
	}
}

func (d *data) handle(metric *model.Metric) {
	d.rw.Lock()
	defer d.rw.Unlock()
	for _, m := range d.value.Metrics {
		if reflect.DeepEqual(m.Labels, metric.Labels) {
			m.Merge(metric)
			return
		}
	}
	d.value.Metrics = append(d.value.Metrics, metric.Clone())
}

func (d *data) get() *model.MetricFamily {
	d.rw.RLock()
	tmp := d.value.Clone()
	d.rw.RUnlock()
	return tmp
}

func (d *data) getLabels(maxLatency, minLatency uint64) []*model.Label {
	// label集合
	labelSet := make(map[model.Label]struct{})
	d.rw.RLock()
	if maxLatency == 0 && minLatency == 0 {
		for _, metric := range d.value.Metrics {
			for _, label := range metric.Labels {
				labelSet[*label] = struct{}{}
			}
		}
	} else {
		for _, metric := range d.value.Metrics {
			var flag bool
			if maxLatency != 0 && metric.MaxGauge != nil {
				if metric.MaxGauge.Value > maxLatency {
					flag = true
				}
			}
			if minLatency != 0 && metric.MinGauge != nil {
				if metric.MinGauge.Value < minLatency {
					flag = true
				}
			}
			if flag {
				for _, label := range metric.Labels {
					if !strings.Contains(label.Name, "api") {
						labelSet[*label] = struct{}{}
					}
				}
			}
		}
	}
	d.rw.RUnlock()
	labels := make([]*model.Label, 0, len(labelSet))
	for k := range labelSet {
		labels = append(labels, &model.Label{
			Name:  k.Name,
			Value: k.Value,
		})
	}

	sort.Sort(model.Labels(labels))
	return labels
}

type memoryStats struct {
	link *dataLink
	all  *data
}

func NewStore(length int) *memoryStats {
	return &memoryStats{
		link: newDataLink(length),
		all:  newData(),
	}
}

// GetLabels 获取所有的标签信息，从ALL桶获取
func (m *memoryStats) GetLabels(maxLatency, minLatency uint64) []*model.Label {
	return m.all.getLabels(maxLatency, minLatency)
}

// Filter 通过天数以及标签信息，过滤监控数据
func (m *memoryStats) Filter(f Filter) []*model.MetricFamily {
	if len(f.Days) == 0 {
		return []*model.MetricFamily{}
	}
	// 通过天数过滤监控数据
	metricsData := m.getMetricsByDay(f.Days)
	// 如果标签为空，直接返回
	if len(f.Labels) == 0 {
		return metricsData
	}

	for _, v := range metricsData {
		m.filterMetricsByLabels(f.Labels, v)
	}
	// 根据label过滤监控数据
	return metricsData
}

// getMetricsByDay 通过天数获取监控数据
func (m *memoryStats) getMetricsByDay(dayIndexes []int) []*model.MetricFamily {
	res := make([]*model.MetricFamily, 0, len(dayIndexes))
	sort.Ints(dayIndexes)
	for _, dayIndex := range dayIndexes {
		if dayIndex != allDayIndex {
			res = append(res, m.link.getByIndex(int64(dayIndex)))
			continue
		}
		res = append(res, m.all.get())
	}
	return res
}

// getMetricsByLabels 通过标签过滤监控数据
func (m *memoryStats) filterMetricsByLabels(labels model.Labels, metricFamily *model.MetricFamily) {
	res := make([]*model.Metric, 0, len(metricFamily.Metrics))
	for _, metric := range metricFamily.Metrics {
		if !model.More(metric.Labels, labels) {
			continue
		}
		res = append(res, metric)
	}
	// 更新结果
	metricFamily.Metrics = res
}

func (m *memoryStats) handle(metric *model.Metric) {
	m.link.handle(metric)
	m.all.handle(metric)
}
