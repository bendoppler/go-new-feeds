package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var LoginLatency = promauto.NewSummaryVec(
	prometheus.SummaryOpts{
		Namespace: "newsfeed",
		Subsystem: "webapp",
		Name:      "latency",
		Help:      "Latency in milliseconds for various endpoints",
		Objectives: map[float64]float64{
			0.5:  0.05,
			0.9:  0.01,
			0.99: 0.001,
		},
	},
	[]string{"endpoint", "status"},
)

// RecordLoginLatency tracks the latency for the login endpoint
func RecordLoginLatency(status string, duration float64) {
	LoginLatency.WithLabelValues("login", status).Observe(duration)
}
