package prometheus

import (
	"math"
	"runtime"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/mem"
)

var lastTimes []cpu.TimesStat

func init() {
	lastTimes, _ = cpu.Times(false)
}

func NewCpuMemoryMetricsHandler(serviceName string) *cpuMemoryMetricsHandler {
	labels := prometheus.Labels{"service_name": serviceName}
	return &cpuMemoryMetricsHandler{
		memTotal:       prometheus.NewDesc("dcs_mem_total", "mem total number of bytes", nil, labels),
		memUsed:        prometheus.NewDesc("dcs_mem_used", "mem userd number of bytes", nil, labels),
		memFree:        prometheus.NewDesc("dcs_mem_free", "mem free number of bytes", nil, labels),
		memUsedPercent: prometheus.NewDesc("dcs_mem_used_percent", "mem used percent", nil, labels),
		cpuUser:        prometheus.NewDesc("dcs_cpu_user", "cpu user", nil, labels),
		cpuSystem:      prometheus.NewDesc("dcs_cpu_system", "cpu system", nil, labels),
		cpuIdle:        prometheus.NewDesc("dcs_cpu_idle", "cpu idle", nil, labels),
		cpuNice:        prometheus.NewDesc("dcs_cpu_nice", "cpu nice", nil, labels),
		cpuIowait:      prometheus.NewDesc("dcs_cpu_iowait", "cpu iowait", nil, labels),
		cpuIrq:         prometheus.NewDesc("dcs_cpu_irq", "cpu irq", nil, labels),
		cpuSoftirq:     prometheus.NewDesc("dcs_cpu_softirq", "cpu softirq", nil, labels),
	}
}

type cpuMemoryMetricsHandler struct {
	memTotal       *prometheus.Desc
	memUsed        *prometheus.Desc
	memFree        *prometheus.Desc
	memUsedPercent *prometheus.Desc
	cpuUser        *prometheus.Desc
	cpuSystem      *prometheus.Desc
	cpuIdle        *prometheus.Desc
	cpuNice        *prometheus.Desc
	cpuIowait      *prometheus.Desc
	cpuIrq         *prometheus.Desc
	cpuSoftirq     *prometheus.Desc
}

// Describe sends the super-set of all possible descriptors of metrics
// collected by this Collector to the provided channel and returns once
// the last descriptor has been sent. The sent descriptors fulfill the
// consistency and uniqueness requirements described in the Desc
// documentation.
//
// It is valid if one and the same Collector sends duplicate
// descriptors. Those duplicates are simply ignored. However, two
// different Collectors must not send duplicate descriptors.
//
// Sending no descriptor at all marks the Collector as “unchecked”,
// i.e. no checks will be performed at registration time, and the
// Collector may yield any Metric it sees fit in its Collect method.
//
// This method idempotently sends the same descriptors throughout the
// lifetime of the Collector. It may be called concurrently and
// therefore must be implemented in a concurrency safe way.
//
// If a Collector encounters an error while executing this method, it
// must send an invalid descriptor (created with NewInvalidDesc) to
// signal the error to the registry.
func (cp *cpuMemoryMetricsHandler) Describe(ch chan<- *prometheus.Desc) {
	ch <- cp.memTotal
	ch <- cp.memUsed
	ch <- cp.memFree
	ch <- cp.memUsedPercent
	ch <- cp.cpuUser
	ch <- cp.cpuSystem
	ch <- cp.cpuIdle
	ch <- cp.cpuNice
	ch <- cp.cpuIowait
	ch <- cp.cpuIrq
	ch <- cp.cpuSoftirq
}

// Collect is called by the Prometheus registry when collecting
// metrics. The implementation sends each collected metric via the
// provided channel and returns once the last metric has been sent. The
// descriptor of each sent metric is one of those returned by Describe
// (unless the Collector is unchecked, see above). Returned metrics that
// share the same descriptor must differ in their variable label
// values.
//
// This method may be called concurrently and must therefore be
// implemented in a concurrency safe way. Blocking occurs at the expense
// of total performance of rendering all registered metrics. Ideally,
// Collector implementations support concurrent readers.
func (cp *cpuMemoryMetricsHandler) Collect(ch chan<- prometheus.Metric) {
	v, err := mem.VirtualMemory()
	if err == nil {
		ch <- prometheus.MustNewConstMetric(cp.memTotal, prometheus.GaugeValue, float64(v.Total))
		ch <- prometheus.MustNewConstMetric(cp.memUsed, prometheus.GaugeValue, float64(v.Used))
		ch <- prometheus.MustNewConstMetric(cp.memFree, prometheus.GaugeValue, float64(v.Free))
		ch <- prometheus.MustNewConstMetric(cp.memUsedPercent, prometheus.GaugeValue, v.UsedPercent)
	}
	c, err := cpu.Times(false)
	if err == nil {
		cp.calculateAndSend(lastTimes[0], c[0], ch)
	}
}

func getTotal(t cpu.TimesStat) float64 {
	tot := t.User + t.System + t.Idle + t.Nice + t.Iowait + t.Irq +
		t.Softirq + t.Steal + t.Guest + t.GuestNice

	if runtime.GOOS == "linux" {
		tot -= t.Guest     // Linux 2.6.24+
		tot -= t.GuestNice // Linux 3.2.0+
	}
	return tot
}

func (cp *cpuMemoryMetricsHandler) calculateAndSend(
	t1, t2 cpu.TimesStat,
	ch chan<- prometheus.Metric,
) {
	t1All := getTotal(t1)
	t2All := getTotal(t2)

	ch <- prometheus.MustNewConstMetric(cp.cpuUser, prometheus.GaugeValue, calculatePercent(t1.User, t2.User, t1All, t2All))
	ch <- prometheus.MustNewConstMetric(cp.cpuSystem, prometheus.GaugeValue, calculatePercent(t1.System, t2.System, t1All, t2All))
	ch <- prometheus.MustNewConstMetric(cp.cpuIdle, prometheus.GaugeValue, calculatePercent(t1.Idle, t2.Idle, t1All, t2All))
	ch <- prometheus.MustNewConstMetric(cp.cpuNice, prometheus.GaugeValue, calculatePercent(t1.Nice, t2.Nice, t1All, t2All))
	ch <- prometheus.MustNewConstMetric(cp.cpuIowait, prometheus.GaugeValue, calculatePercent(t1.Iowait, t2.Iowait, t1All, t2All))
	ch <- prometheus.MustNewConstMetric(cp.cpuIrq, prometheus.GaugeValue, calculatePercent(t1.Irq, t2.Irq, t1All, t2All))
	ch <- prometheus.MustNewConstMetric(cp.cpuSoftirq, prometheus.GaugeValue, calculatePercent(t1.Softirq, t2.Softirq, t1All, t2All))
}

func calculatePercent(t1, t2, t1Total, t2Total float64) float64 {
	if t2 <= t1 {
		return 0
	}
	if t2Total <= t1Total {
		return -100
	}

	return math.Min(100, math.Max(0, (t2-t1)/(t2Total-t1Total)*100))
}
