package metrics

import (
	"fmt"
	"net/http"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func NewMetricsValues() *MetricsValues {
	return &MetricsValues{
		TestStatus:       map[string]float64{},
		TotalTests:       0,
		TotalTestsPassed: 0,
		TotalTestsFailed: 0,
	}
}

func NewServer(address string, port int) *Server {

	return &Server{
		Port:    port,
		Address: address,
		Path:    "/metrics",
		Metrics: Metrics{
			TestStatus: promauto.NewGaugeVec(
				prometheus.GaugeOpts{
					Name: "kubetest_test_status",
					Help: "A 0/1 metrics to indicate if a given integration tests has passed or failed",
				},
				[]string{
					"name",
				},
			),
			TotalTests: promauto.NewGauge(
				prometheus.GaugeOpts{
					Name: "kubetest_total_tests",
					Help: "Total number of tests executed in the last run",
				},
			),
			TotalTestsPassed: promauto.NewGauge(
				prometheus.GaugeOpts{
					Name: "kubetest_total_tests_passed",
					Help: "Total number of passed tests in the last execution",
				},
			),
			TotalTestsFailed: promauto.NewGauge(
				prometheus.GaugeOpts{
					Name: "kubetest_total_tests_failed",
					Help: "Total number of failed tests in the last execution",
				},
			),
		},
	}
}

func (s *Server) Serve() {

	http.Handle("/metrics", promhttp.Handler())
	http.ListenAndServe(fmt.Sprintf("%s:%d", s.Address, s.Port), nil)

}
