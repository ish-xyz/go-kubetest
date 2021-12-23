package metrics

import "github.com/prometheus/client_golang/prometheus"

type Metrics struct {
	TestStatus       *prometheus.GaugeVec
	TotalTests       prometheus.Gauge
	TotalTestsPassed prometheus.Gauge
	TotalTestsFailed prometheus.Gauge
}

type MetricsValues struct {
	TestStatus       map[string]float64
	TotalTests       float64
	TotalTestsPassed float64
	TotalTestsFailed float64
}

type Server struct {
	Port    int
	Address string
	Path    string
	Metrics Metrics
}
