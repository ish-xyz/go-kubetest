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
		AssertionStatus:  map[string]float64{},
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
			AssertionStatus: promauto.NewGaugeVec(
				prometheus.GaugeOpts{
					Name: "kubetest_assertion_status",
					Help: "A 0/1 metrics to indicate if a given assertion has passed or failed",
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

func (mv *MetricsValues) Store(testName string, result bool, assertionResults map[string]interface{}) {

	set := func(result bool) float64 {
		if result {
			return 1
		} else {
			return 0
		}
	}

	for key, val := range assertionResults {
		mv.AssertionStatus[key] = set(val.(bool))
	}

	mv.TotalTestsFailed += set(!result)
	mv.TotalTestsPassed += set(result)
	mv.TestStatus[testName] = set(result)
	mv.TotalTests += 1
}

func (mv *MetricsValues) Publish(ms *Server) {

	for key, value := range mv.TestStatus {
		ms.Metrics.TestStatus.WithLabelValues(key).Set(value)
	}

	for key, value := range mv.AssertionStatus {
		ms.Metrics.AssertionStatus.WithLabelValues(key).Set(value)

	}

	ms.Metrics.TotalTestsFailed.Set(mv.TotalTestsFailed)
	ms.Metrics.TotalTestsPassed.Set(mv.TotalTestsPassed)
	ms.Metrics.TotalTests.Set(mv.TotalTests)
}
