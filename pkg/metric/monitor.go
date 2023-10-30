package metric

import (
	"context"
	"strconv"
	"time"

	"github.com/spf13/viper"
	"go.uber.org/zap"

	"template/pkg/conc/pool"
	"template/pkg/logger"
	"template/pkg/metric/model"
)

type option struct {
	handler       Handler
	goroutineNums uint8
	bucketNums    uint8
}

type Option func(*option)

// WithHandler set a mid handler to handle metric
func WithHandler(handler Handler) Option {
	return func(o *option) {
		o.handler = handler
	}
}

// WithGoroutines set goroutine pool size
func WithGoroutines(nums uint8) Option {
	return func(o *option) {
		o.goroutineNums = nums
	}
}

// WithLinkLength set max load day
func WithLinkLength(l uint8) Option {
	return func(o *option) {
		o.bucketNums = l
	}
}

// New creates a metric
func New(ctx context.Context, opts ...Option) *monitor {
	o := &option{
		handler:       NopHandler{},
		goroutineNums: GoroutineNums,
		bucketNums:    BucketNums,
	}
	for _, f := range opts {
		f(o)
	}
	return &monitor{
		handler:         o.handler,
		pool:            pool.New().WithContext(ctx),
		goroutineLength: o.goroutineNums,
		value:           NewStore(int(o.bucketNums)),
	}
}

type monitor struct {
	handler         Handler
	pool            *pool.ContextPool
	goroutineLength uint8
	value           *memoryStats
}

func (m *monitor) Run(metricChannel <-chan Writer) {
	for i := 0; i < int(m.goroutineLength); i++ {
		m.pool.Go(func(ctx context.Context) error {
			for {
				select {
				case w, ok := <-metricChannel:
					if !ok {
						return nil
					}
					metric := &model.Metric{}
					if err := w.Write(metric); err != nil {
						logger.From(ctx).Error("", zap.Error(err))
					}
					m.work(m.handler.Handle(metric))
				case <-ctx.Done():
					return ctx.Err()
				}
			}
		})
	}
	_ = m.pool.Wait()
}

func (m *monitor) work(metric *model.Metric) {
	m.value.handle(metric)
}

// GetMaxDay 获取最大的天数，即Link的长度
func (m *monitor) GetMaxDay() int {
	return m.value.link.len()
}

// GetLabels 获取所有的标签
func (m *monitor) GetLabels(maxLatency, minLatency uint64) []*model.Label {
	return m.value.GetLabels(maxLatency, minLatency)
}

type Filter struct {
	Labels model.Labels
	Days   []int
	// 查询指定条数的数据
	Number int
	// 排序字段
	Type string `json:"type" binding:"omitempty,oneof=max average count"`
}

type Metrics struct {
	MaxLatency          uint64 `json:"max_latency"`
	MinLatency          uint64 `json:"min_latency"`
	AverageLatency      uint64 `json:"average_latency"`
	RequestCount        uint64 `json:"request_count"`
	RequestTotalLatency uint64 `json:"-"`
	Day                 string `json:"day"`
}

func (m *monitor) Metrics(f Filter) []*Metrics {
	metrics := m.value.Filter(f)
	res := make([]*Metrics, len(metrics))
	if len(metrics) == 0 {
		return res
	}
	for i, metric := range metrics {
		res[i] = &Metrics{}
		if metric.Name == "" {
			continue
		}
		// 先遍历获取总次数以及总时延
		for _, metricData := range metric.Metrics {
			res[i].Day = time.Unix(metricData.Day*(60*60*24), 0).Format(YMDFormat)
			if model.ContainLabel(metricData.Labels, LabelRequestCount[0]) {
				res[i].RequestCount += metricData.Counter.Value
				continue
			}
			if model.ContainLabel(metricData.Labels, LabelRequestTotalLatency[0]) {
				res[i].RequestTotalLatency += metricData.Counter.Value
				continue
			}
			if model.ContainLabel(metricData.Labels, LabelMaxLatency[0]) {
				if res[i].MaxLatency < metricData.MaxGauge.Value {
					res[i].MaxLatency = metricData.MaxGauge.Value
				}
				continue
			}
			if model.ContainLabel(metricData.Labels, LabelMinLatency[0]) {
				if res[i].MinLatency > metricData.MinGauge.Value {
					res[i].MinLatency = metricData.MinGauge.Value
				}
				continue
			}
		}
		if res[i].RequestCount != 0 {
			res[i].AverageLatency = res[i].RequestTotalLatency / res[i].RequestCount
		}
	}
	return res
}

type MetricsSort struct {
	MaxLatency          uint64 `json:"max_latency,omitempty" csv:"最大时延（ms）"`
	MinLatency          uint64 `json:"min_latency" csv:"最小时延（ms）"`
	AverageLatency      uint64 `json:"average_latency,omitempty" csv:"平均时延（ms）"`
	RequestCount        uint64 `json:"request_count,omitempty" csv:"请求总数（次）"`
	RequestTotalLatency uint64 `json:"-"`
	Day                 string `json:"day" csv:"日期"`
	Label               string `json:"label" csv:"资源标签"`
}

func (m *monitor) MetricsSort(f Filter) []*MetricsSort {
	metrics := m.value.Filter(f)
	var result = make(map[string]*MetricsSort)
	for _, metric := range metrics {
		if metric.Name == "" {
			continue
		}

		// 先遍历获取总次数以及总时延
		for _, metricData := range metric.Metrics {
			label := metricData.Labels[1].Value
			curDay := time.Unix(metricData.Day*(60*60*24), 0).Format(YMDFormat)
			tmp, ok := result[label+curDay]
			if !ok {
				tmp = &MetricsSort{}
			}
			if model.ContainLabel(metricData.Labels, LabelRequestCount[0]) {
				tmp.RequestCount += metricData.Counter.Value
			}
			if model.ContainLabel(metricData.Labels, LabelRequestTotalLatency[0]) {
				tmp.RequestTotalLatency += metricData.Counter.Value
			}
			if model.ContainLabel(metricData.Labels, LabelMaxLatency[0]) {
				if tmp.MaxLatency < metricData.MaxGauge.Value {
					tmp.MaxLatency = metricData.MaxGauge.Value
				}
			}
			if model.ContainLabel(metricData.Labels, LabelMinLatency[0]) {
				if tmp.MinLatency > metricData.MinGauge.Value {
					tmp.MinLatency = metricData.MinGauge.Value
				}
			}

			tmp.Day = curDay
			tmp.Label = label
			result[label+curDay] = tmp
		}
	}

	var res []*MetricsSort
	for _, v := range result {
		if v.RequestCount != 0 {
			v.AverageLatency = v.RequestTotalLatency / v.RequestCount
		}
		res = append(res, v)
	}

	analyseResult := metricsSort(f, res)
	if analyseResult == nil {
		analyseResult = make([]*MetricsSort, 0)
	}
	return analyseResult
}

type MetricsSortTable struct {
	List []*MetricsSort `json:"list"`
}

func GenerateTableAttr(f Filter) ([]string, string) {
	// 生成 sheet = service_max_top?
	name := viper.GetString("project.name")
	sheet := name + "_" + f.Type + "_top" + strconv.Itoa(f.Number)

	// 生成表头信息
	tableHeader := []string{"资源标签"}
	switch f.Type {
	case SortWithMaxLatency:
		tableHeader = append(tableHeader, "最大时延（ms）")
	case SortWithMinLatency:
		tableHeader = append(tableHeader, "最小时延（ms）")
	case SortWithAverageLatency:
		tableHeader = append(tableHeader, "平均时延（ms）")
	case SortWithRequestCount:
		tableHeader = append(tableHeader, "请求总数（次）")
	}
	tableHeader = append(tableHeader, "日期")

	return tableHeader, sheet
}

func metricsSort(f Filter, res []*MetricsSort) []*MetricsSort {
	var (
		max         sortMaxLatency
		min         sortMinLatency
		average     sortAverageLatency
		count       sortRequestCount
		sortMetrics []*MetricsSort
	)

	switch f.Type {
	case SortWithMaxLatency:
		max = res
		max.Sort()
		sortMetrics = max
	case SortWithMinLatency:
		min = res
		min.Sort()
		sortMetrics = min
	case SortWithAverageLatency:
		average = res
		average.Sort()
		sortMetrics = average
	case SortWithRequestCount:
		count = res
		count.Sort()
		sortMetrics = count
	}

	if f.Number != 0 && len(sortMetrics) > f.Number {
		sortMetrics = sortMetrics[:f.Number]
	}

	metricsResult(f, sortMetrics)
	return sortMetrics
}

func metricsResult(f Filter, metrics []*MetricsSort) {
	switch f.Type {
	case SortWithMaxLatency:
		for _, metric := range metrics {
			metric.MinLatency = 0
			metric.AverageLatency = 0
			metric.RequestCount = 0
		}
	case SortWithMinLatency:
		for _, metric := range metrics {
			metric.MaxLatency = 0
			metric.AverageLatency = 0
			metric.RequestCount = 0
		}
	case SortWithAverageLatency:
		for _, metric := range metrics {
			metric.MaxLatency = 0
			metric.RequestCount = 0
			metric.MinLatency = 0
		}
	case SortWithRequestCount:
		for _, metric := range metrics {
			metric.AverageLatency = 0
			metric.MaxLatency = 0
			metric.MinLatency = 0
		}
	}
}
