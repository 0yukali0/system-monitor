package metrics

import (
	"net/http"
	"time"

	"github.com/julienschmidt/httprouter"
	"github.com/prometheus/client_golang/prometheus"
)

var (
	httpRequestsTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "http_requests_total",
			Help: "Total number of HTTP requests",
		},
		[]string{"method", "path"},
	)

	httpRequestDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name: "http_request_duration_seconds",
			Help: "Histogram of response time for handler in seconds",
			Buckets: []float64{
				1e-5, // 10 µs
				1e-4, // 100 µs
				1e-3, // 1 ms (millisecond)
				1e-2, // 10 ms
				1e-1, // 100 ms
				1,    // 1 s (second)
				10,   // 10 s
			},
		},
		[]string{"method", "path"},
	)

	CPU_usage = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "cpu_resource_usage",
			Help: "cpu percentage",
		},
		[]string{"node"},
	)

	Mem_usage = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "memory_resource_usage",
			Help: "memory percentage",
		},
		[]string{"node"},
	)
	Net_usage = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "net_byte",
			Help: "net bytes",
		},
		[]string{"node", "method"},
	)
)

func init() {
	prometheus.MustRegister(httpRequestsTotal)
	prometheus.MustRegister(httpRequestDuration)
}

func MetricsMiddleware(next httprouter.Handle) httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
		start := time.Now()
		next(w, r, ps)
		duration := time.Since(start).Seconds()

		httpRequestsTotal.WithLabelValues(r.Method, r.URL.Path).Inc()
		httpRequestDuration.WithLabelValues(r.Method, r.URL.Path).Observe(duration)
	}
}
