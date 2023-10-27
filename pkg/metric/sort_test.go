package metric

import (
	"encoding/json"
	"reflect"
	"testing"

	"template/pkg/metric/model"
)

func Test_sortMaxLatency_Sort(t *testing.T) {
	tests := []struct {
		name string
		m    sortMaxLatency
		want sortMaxLatency
	}{
		{
			name: "MaxLatency-1",
			m: sortMaxLatency{
				{
					MaxLatency: 5,
				},
				{
					MaxLatency: 6,
				},
				{
					MaxLatency: 1,
				},
				{
					MaxLatency: 2,
				},
				{
					MaxLatency: 3,
				},
			},
			want: sortMaxLatency{
				{
					MaxLatency: 6,
				},
				{
					MaxLatency: 5,
				},
				{
					MaxLatency: 3,
				},
				{
					MaxLatency: 2,
				},
				{
					MaxLatency: 1,
				},
			},
		},
		{
			name: "MaxLatency-2",
			m: sortMaxLatency{
				{
					MaxLatency: 2,
				},
				{
					MaxLatency: 3,
				},
				{
					MaxLatency: 5,
				},
				{
					MaxLatency: 1,
				},
				{
					MaxLatency: 6,
				},
			},
			want: sortMaxLatency{
				{
					MaxLatency: 6,
				},
				{
					MaxLatency: 5,
				},
				{
					MaxLatency: 3,
				},
				{
					MaxLatency: 2,
				},
				{
					MaxLatency: 1,
				},
			},
		},
		{
			name: "MaxLatency-3",
			m: sortMaxLatency{
				{
					MaxLatency: 1,
				},
				{
					MaxLatency: 6,
				},
				{
					MaxLatency: 5,
				},
				{
					MaxLatency: 2,
				},
				{
					MaxLatency: 3,
				},
			},
			want: sortMaxLatency{
				{
					MaxLatency: 6,
				},
				{
					MaxLatency: 5,
				},
				{
					MaxLatency: 3,
				},
				{
					MaxLatency: 2,
				},
				{
					MaxLatency: 1,
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.m.Sort()
			if !reflect.DeepEqual(tt.m, tt.want) {
				result, _ := json.Marshal(tt.m)
				want, _ := json.Marshal(tt.m)
				t.Errorf("result: %v\nwant: %v\n", string(result), string(want))
			}
		})
	}
}

func Test_sortRequestCount_Sort(t *testing.T) {
	tests := []struct {
		name string
		r    sortRequestCount
		want sortRequestCount
	}{
		{
			name: "requestCount-1",
			r: sortRequestCount{
				{
					RequestCount: 100,
				},
				{
					RequestCount: 500,
				},
				{
					RequestCount: 300,
				},
				{
					RequestCount: 700,
				},
				{
					RequestCount: 200,
				},
			},
			want: sortRequestCount{
				{
					RequestCount: 700,
				},
				{
					RequestCount: 500,
				},
				{
					RequestCount: 300,
				},
				{
					RequestCount: 200,
				},
				{
					RequestCount: 100,
				},
			},
		},
		{
			name: "requestCount-2",
			r: sortRequestCount{
				{
					RequestCount: 100,
				},
				{
					RequestCount: 200,
				},
				{
					RequestCount: 300,
				},
				{
					RequestCount: 500,
				},
				{
					RequestCount: 700,
				},
			},
			want: sortRequestCount{
				{
					RequestCount: 700,
				},
				{
					RequestCount: 500,
				},
				{
					RequestCount: 300,
				},
				{
					RequestCount: 200,
				},
				{
					RequestCount: 100,
				},
			},
		},
		{
			name: "requestCount-3",
			r: sortRequestCount{
				{
					RequestCount: 300,
				},
				{
					RequestCount: 500,
				},
				{
					RequestCount: 100,
				},
				{
					RequestCount: 200,
				},
				{
					RequestCount: 700,
				},
			},
			want: sortRequestCount{
				{
					RequestCount: 700,
				},
				{
					RequestCount: 500,
				},
				{
					RequestCount: 300,
				},
				{
					RequestCount: 200,
				},
				{
					RequestCount: 100,
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.r.Sort()
			if !reflect.DeepEqual(tt.r, tt.want) {
				result, _ := json.Marshal(tt.r)
				want, _ := json.Marshal(tt.r)
				t.Errorf("result: %v\nwant: %v\n", string(result), string(want))
			}
		})
	}
}

func Test_sortAverageLatency_Sort(t *testing.T) {
	tests := []struct {
		name string
		a    sortAverageLatency
		want sortAverageLatency
	}{
		{
			name: "averageLatency-1",
			a: sortAverageLatency{
				{
					AverageLatency: 20,
				},
				{
					AverageLatency: 30,
				},
				{
					AverageLatency: 50,
				},
				{
					AverageLatency: 70,
				},
				{
					AverageLatency: 90,
				},
			},
			want: sortAverageLatency{
				{
					AverageLatency: 90,
				},
				{
					AverageLatency: 70,
				},
				{
					AverageLatency: 50,
				},
				{
					AverageLatency: 30,
				},
				{
					AverageLatency: 20,
				},
			},
		},
		{
			name: "averageLatency-2",
			a: sortAverageLatency{
				{
					AverageLatency: 20,
				},
				{
					AverageLatency: 90,
				},
				{
					AverageLatency: 70,
				},
				{
					AverageLatency: 50,
				},
				{
					AverageLatency: 30,
				},
			},
			want: sortAverageLatency{
				{
					AverageLatency: 90,
				},
				{
					AverageLatency: 70,
				},
				{
					AverageLatency: 50,
				},
				{
					AverageLatency: 30,
				},
				{
					AverageLatency: 20,
				},
			},
		},
		{
			name: "averageLatency-3",
			a: sortAverageLatency{
				{
					AverageLatency: 50,
				},
				{
					AverageLatency: 70,
				},
				{
					AverageLatency: 30,
				},
				{
					AverageLatency: 90,
				},
				{
					AverageLatency: 20,
				},
			},
			want: sortAverageLatency{
				{
					AverageLatency: 90,
				},
				{
					AverageLatency: 70,
				},
				{
					AverageLatency: 50,
				},
				{
					AverageLatency: 30,
				},
				{
					AverageLatency: 20,
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.a.Sort()
			if !reflect.DeepEqual(tt.a, tt.want) {
				result, _ := json.Marshal(tt.a)
				want, _ := json.Marshal(tt.a)
				t.Errorf("result: %v\nwant: %v\n", string(result), string(want))
			}
		})
	}
}

func Test_memoryStats_filterMetricsByLabels(t *testing.T) {

	labelDetail := "GET-/metric/details"
	detailsArg := []*model.Metric{
		{
			Labels: []*model.Label{
				{
					Name:  "api",
					Value: "request_max_latency",
				},
				{
					Name:  "path",
					Value: labelDetail,
				},
			},
			MaxGauge: &model.MaxGauge{Value: 10},
			Day:      19388,
		},
		{
			Labels: []*model.Label{
				{
					Name:  "api",
					Value: "request_min_latency",
				},
				{
					Name:  "path",
					Value: labelDetail,
				},
			},
			MinGauge: &model.MinGauge{Value: 0},
			Day:      19388,
		},
		{
			Labels: []*model.Label{
				{
					Name:  "api",
					Value: "request_count",
				},
				{
					Name:  "path",
					Value: labelDetail,
				},
			},
			Counter: &model.Counter{Value: 1},
			Day:     19388,
		},
		{
			Labels: []*model.Label{
				{
					Name:  "api",
					Value: "request_average_latency",
				},
				{
					Name:  "path",
					Value: labelDetail,
				},
			},
			Day: 19388,
		},
	}

	labelQuery := "GET-/metric/query"
	queryArg := []*model.Metric{
		{
			Labels: []*model.Label{
				{
					Name:  "api",
					Value: "request_max_latency",
				},
				{
					Name:  "path",
					Value: labelQuery,
				},
			},
			MaxGauge: &model.MaxGauge{Value: 20},
			Day:      19388,
		},
		{
			Labels: []*model.Label{
				{
					Name:  "api",
					Value: "request_min_latency",
				},
				{
					Name:  "path",
					Value: labelQuery,
				},
			},
			MinGauge: &model.MinGauge{Value: 0},
			Day:      19388,
		},
		{
			Labels: []*model.Label{
				{
					Name:  "api",
					Value: "request_count",
				},
				{
					Name:  "path",
					Value: labelQuery,
				},
			},
			Counter: &model.Counter{Value: 2},
			Day:     19388,
		},
		{
			Labels: []*model.Label{
				{
					Name:  "api",
					Value: "request_average_latency",
				},
				{
					Name:  "path",
					Value: labelQuery,
				},
			},
			Day: 19388,
		},
	}

	labelStat := "GET-/metric/stat"
	statArg := []*model.Metric{
		{
			Labels: []*model.Label{
				{
					Name:  "api",
					Value: "request_max_latency",
				},
				{
					Name:  "path",
					Value: labelStat,
				},
			},
			MaxGauge: &model.MaxGauge{Value: 30},
			Day:      19387,
		},
		{
			Labels: []*model.Label{
				{
					Name:  "api",
					Value: "request_min_latency",
				},
				{
					Name:  "path",
					Value: labelStat,
				},
			},
			MinGauge: &model.MinGauge{Value: 0},
			Day:      19387,
		},
		{
			Labels: []*model.Label{
				{
					Name:  "api",
					Value: "request_count",
				},
				{
					Name:  "path",
					Value: labelStat,
				},
			},
			Counter: &model.Counter{Value: 3},
			Day:     19387,
		},
		{
			Labels: []*model.Label{
				{
					Name:  "api",
					Value: "request_average_latency",
				},
				{
					Name:  "path",
					Value: labelStat,
				},
			},
			Day: 19387,
		},
	}

	var metrics []*model.Metric
	metrics = append(metrics, detailsArg...)
	metrics = append(metrics, queryArg...)

	type fields struct {
		link *dataLink
		all  *data
	}
	type args struct {
		labels       model.Labels
		metricFamily *model.MetricFamily
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   int
	}{
		{
			name:   "success-yesterday",
			fields: fields{},
			args: args{
				labels: []*model.Label{
					{
						Name:  "path",
						Value: labelStat,
					},
					{
						Name:  "path",
						Value: "test-not-exist",
					},
				},
				metricFamily: &model.MetricFamily{
					Name:    "bucket0",
					Metrics: statArg,
				},
			},
			want: 4,
		},
		{
			name:   "success-label-1",
			fields: fields{},
			args: args{
				labels: []*model.Label{
					{
						Name:  "path",
						Value: labelDetail,
					},
				},
				metricFamily: &model.MetricFamily{
					Name:    "bucket1",
					Metrics: metrics,
				},
			},
			want: 4,
		},
		{
			name:   "success-label-2",
			fields: fields{},
			args: args{
				labels: []*model.Label{
					{
						Name:  "path",
						Value: labelDetail,
					},
					{
						Name:  "path",
						Value: labelQuery,
					},
				},
				metricFamily: &model.MetricFamily{
					Name:    "bucket1",
					Metrics: metrics,
				},
			},
			want: 8,
		},

		{
			name:   "success-label-notExist-2",
			fields: fields{},
			args: args{
				labels: []*model.Label{
					{
						Name:  "path",
						Value: "test-not-exist",
					},
					{
						Name:  "path",
						Value: "test-not-exist-2",
					},
				},
				metricFamily: &model.MetricFamily{
					Name:    "bucket1",
					Metrics: metrics,
				},
			},
			want: 0,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			findLabels := make(map[string]struct{})
			for _, fLabel := range tt.args.labels {
				findLabels[fLabel.Value] = struct{}{}
			}

			m := &memoryStats{
				link: tt.fields.link,
				all:  tt.fields.all,
			}
			m.filterMetricsByLabels([]model.Labels{tt.args.labels}, tt.args.metricFamily)

			// 开始校验结果
			if len(tt.args.metricFamily.Metrics) != tt.want {
				t.Fatalf("get result:%v, want:%v", len(tt.args.metricFamily.Metrics), tt.want)
				return
			}

			for _, metric := range tt.args.metricFamily.Metrics {
				for _, label := range metric.Labels {
					if label.Name != "path" {
						continue
					}
					if _, ok := findLabels[label.Value]; !ok {
						t.Fatalf("result-label: %v, find-label:%v",
							label.Value, findLabels)
					}
				}
			}
		})
	}
}
