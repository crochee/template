package collect

import (
	"context"
	"regexp"
	"time"

	"github.com/cespare/xxhash/v2"
	"github.com/prometheus/client_golang/prometheus"
)

var nameHelpRegexp = regexp.MustCompile(`Desc{fqName: "(.*)", help: "(.*)", constLabels:`)

// BackupMetric 同步数据至存储介质中
func BackupMetric(ctx context.Context, store Store, metric prometheus.Metric) error {
	var m MetricWithUpdate
	if err := metric.Write(&m.Metric); err != nil {
		return err
	}
	desc := metric.Desc().String()
	results := nameHelpRegexp.FindStringSubmatch(desc)
	xxh := xxhash.New()
	xxh.WriteString(results[1])
	xxh.WriteString(results[2])
	for _, lp := range m.Metric.Label {
		xxh.WriteString(*lp.Name)
		xxh.WriteString(*lp.Value)
	}
	id := xxh.Sum64()

	m.Name = results[1]
	m.Help = results[2]
	m.UpdatedAt = time.Now().UnixMilli()
	return store.Put(ctx, id, &m)
}

func Filter(input []*MetricWithUpdate, filter func(*MetricWithUpdate) bool) []*MetricWithUpdate {
	result := make([]*MetricWithUpdate, 0, len(input))
	for _, v := range input {
		if filter(v) {
			result = append(result, v)
		}
	}
	return result
}

func FilterFunc(toleration int64, labels map[string]string) func(v *MetricWithUpdate) bool {
	return func(v *MetricWithUpdate) bool {
		for k, l := range labels {
			for _, v := range v.Metric.Label {
				// 过滤的标签
				if *v.Name == k && *v.Value != l {
					return false
				}
			}
		}
		return time.Now().UnixMilli()-toleration > v.UpdatedAt
	}
}
