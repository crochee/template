package prometheus

import (
	"github.com/prometheus/client_golang/prometheus"
)

var (
	PanicCounterVec = prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "cds_panic_num",
		Help: "panic total counter.",
	}, []string{"method", "path"})
)
