package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"k8s.io/client-go/dynamic"
)

type Metrics struct {
	TestStatus       *prometheus.GaugeVec
	TotalTests       prometheus.Gauge
	TotalTestsPassed prometheus.Gauge
	TotalTestsFailed prometheus.Gauge
	AssertionStatus  *prometheus.GaugeVec
}

type MetricsController struct {
	DynClient dynamic.Interface
	Port      int
	Address   string
	Path      string
	Metrics   Metrics
}
