package prometheus

import (
	"bufio"
	"net"
	"net/http"
	"os"
	"strconv"
	"strings"
	"sync/atomic"

	"github.com/prometheus/client_golang/prometheus"
)

var serviceReceiveConn int64

func NewConnectionMetricHandler(serviceName string) *ConnectionMetricHandler {
	labels := prometheus.Labels{"service_name": serviceName}
	return &ConnectionMetricHandler{
		ServiceConnTotal:    prometheus.NewDesc("service_conn_total", "service connection total", nil, labels),
		ServiceReceiveConn:  prometheus.NewDesc("service_receive_conn", "service receive connection count", nil, labels),
		ServiceReceiveRatio: prometheus.NewDesc("service_receive_ratio", "service receive connection ratio", nil, labels),
	}
}

type ConnectionMetricHandler struct {
	ServiceConnTotal    *prometheus.Desc
	ServiceReceiveConn  *prometheus.Desc
	ServiceReceiveRatio *prometheus.Desc
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
func (cp *ConnectionMetricHandler) Describe(ch chan<- *prometheus.Desc) {
	ch <- cp.ServiceConnTotal
	ch <- cp.ServiceReceiveConn
	ch <- cp.ServiceReceiveRatio
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
func (cp *ConnectionMetricHandler) Collect(ch chan<- prometheus.Metric) {
	file, err := os.Open("/proc/net/sockstat")
	if err != nil {
		return
	}
	defer file.Close()

	var socketsNum float64
	reader := bufio.NewReader(file)
	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			return
		}

		if !strings.HasPrefix(line, "sockets:") {
			continue
		}

		parts := strings.Fields(line)
		if len(parts) > 1 {
			socketsNum, err = strconv.ParseFloat(parts[len(parts)-1], 10)
			if err != nil {
				return
			}
			ch <- prometheus.MustNewConstMetric(cp.ServiceConnTotal, prometheus.GaugeValue, socketsNum)
		}
		break
	}

	if socketsNum == float64(0) {
		return
	}

	receiveConnCount := atomic.LoadInt64(&serviceReceiveConn)
	ch <- prometheus.MustNewConstMetric(cp.ServiceReceiveConn, prometheus.GaugeValue, float64(receiveConnCount))
	serviceReceiveRatio := float64(receiveConnCount * 100 / int64(socketsNum))
	ch <- prometheus.MustNewConstMetric(cp.ServiceReceiveRatio, prometheus.GaugeValue, serviceReceiveRatio)
}

func RegisterHttpConnCountMetric(srv *http.Server) {
	srv.ConnState = func(conn net.Conn, state http.ConnState) {
		switch state {
		case http.StateNew:
			atomic.AddInt64(&serviceReceiveConn, 1)
		case http.StateClosed:
			atomic.AddInt64(&serviceReceiveConn, -1)
		}
	}
}
