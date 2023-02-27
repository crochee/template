package metric

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"template/pkg/metric/model"
)

func Test_dataLink_handle(t *testing.T) {
	type args struct {
		metric *model.Metric
	}
	tests := []struct {
		name   string
		fields *dataLink
		args   args
		want   *dataLink
	}{
		{
			name: "zero",
			fields: &dataLink{
				buckets: []*model.MetricFamily{
					{
						Name:    "1",
						Metrics: nil,
					},
				},
				length: 1,
				cur:    0,
			},
			args: args{
				metric: &model.Metric{
					Labels: []*model.Label{
						{
							Name:  "method",
							Value: "get",
						},
						{
							Name:  "path",
							Value: "/v1/server",
						},
					},
					Counter: &model.Counter{Value: 1},
					Gauge:   nil,
					Day:     0,
				},
			},
			want: &dataLink{
				buckets: []*model.MetricFamily{
					{
						Name: "1",
						Metrics: []*model.Metric{
							{
								Labels: []*model.Label{
									{
										Name:  "method",
										Value: "get",
									},
									{
										Name:  "path",
										Value: "/v1/server",
									},
								},
								Counter: &model.Counter{Value: 1},
								Gauge:   nil,
								Day:     0,
							},
						},
					},
				},
				length: 1,
				cur:    0,
			},
		},
		{
			name: "one metric",
			fields: &dataLink{
				buckets: []*model.MetricFamily{
					{
						Name: "1",
						Metrics: []*model.Metric{
							{
								Labels: []*model.Label{
									{
										Name:  "method",
										Value: "get",
									},
									{
										Name:  "path",
										Value: "/v1/server",
									},
								},
								Counter: &model.Counter{Value: 4},
								Gauge:   nil,
								Day:     5,
							},
						},
					},
				},
				length: 1,
				cur:    0,
			},
			args: args{
				metric: &model.Metric{
					Labels: []*model.Label{
						{
							Name:  "method",
							Value: "get",
						},
						{
							Name:  "path",
							Value: "/v1/server",
						},
					},
					Counter: &model.Counter{Value: 1},
					Gauge:   nil,
					Day:     5,
				},
			},
			want: &dataLink{
				buckets: []*model.MetricFamily{
					{
						Name: "1",
						Metrics: []*model.Metric{
							{
								Labels: []*model.Label{
									{
										Name:  "method",
										Value: "get",
									},
									{
										Name:  "path",
										Value: "/v1/server",
									},
								},
								Counter: &model.Counter{Value: 5},
								Gauge:   nil,
								Day:     5,
							},
						},
					},
				},
				length: 1,
				cur:    0,
			},
		},
		{
			name: "day not same",
			fields: &dataLink{
				buckets: []*model.MetricFamily{
					{
						Name: "1",
						Metrics: []*model.Metric{
							{
								Labels: []*model.Label{
									{
										Name:  "method",
										Value: "get",
									},
									{
										Name:  "path",
										Value: "/v1/server",
									},
								},
								Counter: &model.Counter{Value: 4},
								Gauge:   nil,
								Day:     5,
							},
						},
					},
					{
						Name: "1",
						Metrics: []*model.Metric{
							{
								Labels: []*model.Label{
									{
										Name:  "method",
										Value: "get",
									},
									{
										Name:  "path",
										Value: "/v1/server",
									},
								},
								Counter: &model.Counter{Value: 4},
								Gauge:   nil,
								Day:     5,
							},
						},
					},
				},
				length: 2,
				cur:    0,
			},
			args: args{
				metric: &model.Metric{
					Labels: []*model.Label{
						{
							Name:  "method",
							Value: "get",
						},
						{
							Name:  "path",
							Value: "/v1/server",
						},
					},
					Counter: &model.Counter{Value: 1},
					Gauge:   nil,
					Day:     4,
				},
			},
			want: &dataLink{
				buckets: []*model.MetricFamily{
					{
						Name: "1",
						Metrics: []*model.Metric{
							{
								Labels: []*model.Label{
									{
										Name:  "method",
										Value: "get",
									},
									{
										Name:  "path",
										Value: "/v1/server",
									},
								},
								Counter: &model.Counter{Value: 4},
								Gauge:   nil,
								Day:     5,
							},
						},
					},
					{
						Name: "1",
						Metrics: []*model.Metric{
							{
								Labels: []*model.Label{
									{
										Name:  "method",
										Value: "get",
									},
									{
										Name:  "path",
										Value: "/v1/server",
									},
								},
								Counter: &model.Counter{Value: 1},
								Gauge:   nil,
								Day:     4,
							},
						},
					},
				},
				length: 2,
				cur:    1,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.fields.handle(tt.args.metric)
			assert.Equal(t, tt.want, tt.fields)
		})
	}
}

func Test_data_handle(t *testing.T) {
	type args struct {
		metric *model.Metric
	}
	tests := []struct {
		name   string
		fields *data
		args   args
		want   *data
	}{
		{
			name: "zero",
			fields: &data{
				value: &model.MetricFamily{
					Name:    "",
					Metrics: nil,
				},
			},
			args: args{
				metric: &model.Metric{
					Labels: []*model.Label{
						{
							Name:  "method",
							Value: "POST",
						},
					},
					Counter: &model.Counter{Value: 40},
					Gauge:   nil,
					Day:     9,
				},
			},
			want: &data{
				value: &model.MetricFamily{
					Name: "",
					Metrics: []*model.Metric{
						{
							Labels: []*model.Label{
								{
									Name:  "method",
									Value: "POST",
								},
							},
							Counter: &model.Counter{Value: 40},
							Gauge:   nil,
							Day:     9,
						},
					},
				},
			},
		},
		{
			name: "one same",
			fields: &data{
				value: &model.MetricFamily{
					Name: "",
					Metrics: []*model.Metric{
						{
							Labels: []*model.Label{
								{
									Name:  "method",
									Value: "POST",
								},
							},
							Counter: &model.Counter{Value: 40},
							Gauge:   nil,
							Day:     0,
						},
					},
				},
			},
			args: args{
				metric: &model.Metric{
					Labels: []*model.Label{
						{
							Name:  "method",
							Value: "POST",
						},
					},
					Counter: &model.Counter{Value: 40},
					Gauge:   nil,
					Day:     9,
				},
			},
			want: &data{
				value: &model.MetricFamily{
					Name: "",
					Metrics: []*model.Metric{
						{
							Labels: []*model.Label{
								{
									Name:  "method",
									Value: "POST",
								},
							},
							Counter: &model.Counter{Value: 80},
							Gauge:   nil,
							Day:     9,
						},
					},
				},
			},
		},
		{
			name: "one not same",
			fields: &data{
				value: &model.MetricFamily{
					Name: "",
					Metrics: []*model.Metric{
						{
							Labels: []*model.Label{
								{
									Name:  "method",
									Value: "POST",
								},
							},
							Counter: &model.Counter{Value: 40},
							Gauge:   nil,
							Day:     0,
						},
					},
				},
			},
			args: args{
				metric: &model.Metric{
					Labels: []*model.Label{
						{
							Name:  "method",
							Value: "PUT",
						},
					},
					Counter: &model.Counter{Value: 40},
					Gauge:   nil,
					Day:     9,
				},
			},
			want: &data{
				value: &model.MetricFamily{
					Name: "",
					Metrics: []*model.Metric{
						{
							Labels: []*model.Label{
								{
									Name:  "method",
									Value: "POST",
								},
							},
							Counter: &model.Counter{Value: 40},
							Gauge:   nil,
							Day:     0,
						},
						{
							Labels: []*model.Label{
								{
									Name:  "method",
									Value: "PUT",
								},
							},
							Counter: &model.Counter{Value: 40},
							Gauge:   nil,
							Day:     9,
						},
					},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.fields.handle(tt.args.metric)
			assert.Equal(t, tt.want, tt.fields)
		})
	}
}
