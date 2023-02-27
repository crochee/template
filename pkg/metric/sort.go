package metric

import (
	"sort"
)

const (
	SortWithMaxLatency     = "max"
	SortWithMinLatency     = "min"
	SortWithAverageLatency = "average"
	SortWithRequestCount   = "count"
)

type sortMaxLatency []*MetricsSort

func (m sortMaxLatency) Len() int {
	return len(m)
}

func (m sortMaxLatency) Less(i, j int) bool {
	if m[i].MaxLatency > m[j].MaxLatency {
		return true
	}

	if m[i].MaxLatency < m[j].MaxLatency {
		return false
	}

	return false
}

func (m sortMaxLatency) Swap(i, j int) {
	m[i], m[j] = m[j], m[i]
}

func (m *sortMaxLatency) Sort() {
	sort.Sort(m)
}

type sortMinLatency []*MetricsSort

func (m sortMinLatency) Len() int {
	return len(m)
}

func (m sortMinLatency) Less(i, j int) bool {
	if m[i].MinLatency > m[j].MinLatency {
		return true
	}

	if m[i].MinLatency < m[j].MinLatency {
		return false
	}

	return false
}

func (m sortMinLatency) Swap(i, j int) {
	m[i], m[j] = m[j], m[i]
}

func (m *sortMinLatency) Sort() {
	sort.Sort(m)
}

type sortRequestCount []*MetricsSort

func (r sortRequestCount) Len() int {
	return len(r)
}

func (r sortRequestCount) Less(i, j int) bool {
	if r[i].RequestCount > r[j].RequestCount {
		return true
	}

	if r[i].RequestCount < r[j].RequestCount {
		return false
	}

	return false
}

func (r sortRequestCount) Swap(i, j int) {
	r[i], r[j] = r[j], r[i]
}

func (r *sortRequestCount) Sort() {
	sort.Sort(r)
}

type sortAverageLatency []*MetricsSort

func (a sortAverageLatency) Len() int {
	return len(a)
}

func (a sortAverageLatency) Less(i, j int) bool {
	if a[i].AverageLatency > a[j].AverageLatency {
		return true
	}

	if a[i].AverageLatency < a[j].AverageLatency {
		return false
	}

	return false
}

func (a sortAverageLatency) Swap(i, j int) {
	a[i], a[j] = a[j], a[i]
}

func (a *sortAverageLatency) Sort() {
	sort.Sort(a)
}
