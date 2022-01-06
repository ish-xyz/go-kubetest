package metrics

import (
	"fmt"
	"net/http"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/sirupsen/logrus"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/dynamic/dynamicinformer"
	"k8s.io/client-go/tools/cache"
)

var resource = schema.GroupVersionResource{
	Group:    "go-kubetest.io",
	Version:  "v1",
	Resource: "testresults",
}

func NewMetricsController(dc dynamic.Interface, address string, port int) *MetricsController {
	return &MetricsController{
		DynClient: dc,
		Port:      port,
		Address:   address,
		Path:      "/metrics",
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
					"assertion",
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

func (m *MetricsController) Run(namespace string) {

	logrus.Infoln("Metrics server is starting.")

	// Start metrics web server
	http.Handle("/metrics", promhttp.Handler())
	go http.ListenAndServe(fmt.Sprintf("%s:%d", m.Address, m.Port), nil)

	// Init informer and run it
	sharedInformerFactory := dynamicinformer.NewFilteredDynamicSharedInformerFactory(m.DynClient, 0, namespace, nil)
	genericInformer := sharedInformerFactory.ForResource(resource)
	stopCh := make(chan struct{})
	defer close(stopCh)

	sharedInformer := genericInformer.Informer()
	handlers := cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			m.handlerAddMetrics(obj.(*unstructured.Unstructured))
		},
		UpdateFunc: func(oldObj, obj interface{}) {
			m.handlerUpdateMetrics(obj.(*unstructured.Unstructured))
		},
		DeleteFunc: func(obj interface{}) {
			m.handlerDeleteMetrics(obj.(*unstructured.Unstructured))
		},
	}
	sharedInformer.AddEventHandler(handlers)
	go sharedInformer.Run(stopCh)

	// wait forever
	select {}
}

func (m *MetricsController) handlerAddMetrics(obj *unstructured.Unstructured) {

	delete := false
	spec := obj.Object["spec"].(map[string]interface{})

	m.setMetricTestStatus(delete, obj.GetName(), spec["result"].(bool))
	m.setMetricAssertionStatus(delete, obj.GetName(), spec["assertions"].(map[string]interface{}))
	m.setMetricTotalTests(delete)
	m.setMetricTotalTestsPassed(delete, spec["result"].(bool))
	m.setMetricTotalTestsFailed(delete, spec["result"].(bool))
}

func (m *MetricsController) handlerUpdateMetrics(obj *unstructured.Unstructured) {
	delete := false
	spec := obj.Object["spec"].(map[string]interface{})

	m.setMetricTestStatus(delete, obj.GetName(), spec["result"].(bool))
	m.setMetricAssertionStatus(delete, obj.GetName(), spec["assertions"].(map[string]interface{}))
}

func (m *MetricsController) handlerDeleteMetrics(obj *unstructured.Unstructured) {

	delete := true
	spec := obj.Object["spec"].(map[string]interface{})

	m.setMetricTestStatus(delete, obj.GetName(), spec["result"].(bool))
	m.setMetricAssertionStatus(delete, obj.GetName(), spec["assertions"].(map[string]interface{}))
	m.setMetricTotalTests(delete)
	m.setMetricTotalTestsPassed(delete, spec["result"].(bool))
	m.setMetricTotalTestsFailed(delete, spec["result"].(bool))

}

func (m *MetricsController) setMetricTestStatus(delete bool, key string, value bool) {
	if delete {
		m.Metrics.TestStatus.DeleteLabelValues(key)
		return
	}
	m.Metrics.TestStatus.WithLabelValues(key).Set(getPromVal(value))
}

func (m *MetricsController) setMetricAssertionStatus(delete bool, testName string, assertions map[string]interface{}) {

	if delete {
		for key, _ := range assertions {
			m.Metrics.AssertionStatus.DeleteLabelValues(testName, key)
		}
		return
	}

	for key, value := range assertions {
		m.Metrics.AssertionStatus.WithLabelValues(testName, key).Set(getPromVal(value.(bool)))
	}
}

func (m *MetricsController) setMetricTotalTests(delete bool) {
	if delete {
		m.Metrics.TotalTests.Dec()
		return
	}
	m.Metrics.TotalTests.Inc()
}

func (m *MetricsController) setMetricTotalTestsPassed(delete bool, result bool) {
	if delete {
		m.Metrics.TotalTestsPassed.Sub(getPromVal(result))
		return
	}
	m.Metrics.TotalTestsPassed.Add(getPromVal(result))
}

func (m *MetricsController) setMetricTotalTestsFailed(delete bool, result bool) {
	if delete {
		m.Metrics.TotalTestsFailed.Sub(getPromVal(!result))
		return
	}
	m.Metrics.TotalTestsFailed.Add(getPromVal(!result))
}

func getPromVal(result bool) float64 {
	if result {
		return 1
	}
	return 0
}
