package controller

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/ish-xyz/go-kubetest/pkg/assert"
	"github.com/ish-xyz/go-kubetest/pkg/loader"
	"github.com/ish-xyz/go-kubetest/pkg/metrics"
	"github.com/ish-xyz/go-kubetest/pkg/provisioner"
	"github.com/sirupsen/logrus"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

const defaultMaxWait = "60s"

func NewController(prv provisioner.Provisioner, ms *metrics.Server, a *assert.Assert) *Controller {
	return &Controller{
		Provisioner:   prv,
		MetricsServer: ms,
		Assert:        a,
	}
}

// Start controller for periodically tests executions
func (c *Controller) Run(testsList []*loader.TestDefinition, wait time.Duration) {

	logrus.Infof("Starting metrics server at :%d", c.MetricsServer.Port)
	go c.MetricsServer.Serve()

	logrus.Info("Starting controller")
	for {
		metricsValues := metrics.NewMetricsValues()
		for _, test := range testsList {
			logrus.Infof("Running test: '%s'", test.Name)

			// Create resources and wait for creation
			errors := c.Setup(test.ObjectsList)
			c.WaitForCreation(test.Setup.WaitFor)

			// Run the actual tests
			result := c.Assert.Run(test, errors)
			metricsValues = updateMetricsValues(metricsValues, test.Name, result)

			// Delete resources and wait for deletion
			c.Teardown(test.ObjectsList)
			c.WaitForDeletion(test.Teardown.WaitFor)
		}
		logrus.Debug("Push new metrics to server")
		c.setMetrics(metricsValues)

		logrus.Infof("Waiting for next execution (%s)", wait)
		time.Sleep(wait)
	}

}

// WaitForCreation wait until a set of resources has been created
func (c *Controller) WaitForCreation(resources []loader.WaitFor) {

	condition := func(list []unstructured.Unstructured) bool {
		return len(list) != 0
	}

	waitFor(c.Provisioner, resources, condition)
}

// WaitForDeletion wait until a set of resources has been deleted
func (c *Controller) WaitForDeletion(resources []loader.WaitFor) {

	condition := func(list []unstructured.Unstructured) bool {
		return len(list) == 0
	}

	waitFor(c.Provisioner, resources, condition)
}

// Create resources defined on manifests
func (c *Controller) Setup(objects []*loader.LoadedObject) []string {

	var errors []string

	for _, obj := range objects {
		err := c.Provisioner.CreateOrUpdate(context.TODO(), obj)
		if err != nil {
			logrus.Debugf("Couldn't delete resource %s", obj.Object.GetName())
			logrus.Debugln(err)
			errors = append(errors, fmt.Sprintf("%v", err))
		} else {
			logrus.Debugf("Setup: resource created %s\n", obj.Object.GetName())
		}
	}

	return errors
}

// Delete resources defined on manifests
func (c *Controller) Teardown(objects []*loader.LoadedObject) {

	for index := range objects {
		// Teardown needs to delete the objects in the
		// manifest, from the last one to the first one
		obj := objects[len(objects)-1-index]
		err := c.Provisioner.Delete(context.TODO(), obj)
		if err != nil {
			logrus.Errorf("Teardown: Couldn't delete resource %s", obj.Object.GetName())
			logrus.Errorln(err)
		} else {
			logrus.Debugf("Teardown: Resource deleted %s\n", obj.Object.GetName())
		}
	}
}

// Set actual metrics values for promtheus package to export them
func (c *Controller) setMetrics(metricsValues *metrics.MetricsValues) {

	rMetrics := c.MetricsServer.Metrics

	for key, value := range metricsValues.TestStatus {
		rMetrics.TestStatus.WithLabelValues(key).Set(value)
	}

	rMetrics.TotalTestsFailed.Set(metricsValues.TotalTestsFailed)
	rMetrics.TotalTestsPassed.Set(metricsValues.TotalTestsPassed)
	rMetrics.TotalTests.Set(metricsValues.TotalTests)
}

// Update a temporary struct that will then used to push metrics in one go
func updateMetricsValues(metricsValues *metrics.MetricsValues, testName string, result bool) *metrics.MetricsValues {

	set := func(result bool) float64 {
		if result {
			return 1
		} else {
			return 0
		}
	}

	metricsValues.TotalTestsFailed += set(!result)
	metricsValues.TotalTestsPassed += set(result)
	metricsValues.TestStatus[testName] += set(result)
	metricsValues.TotalTests += 1

	return metricsValues
}

func getResourceDataFromPath(resourcePath string) (string, string, string, string) {

	path := strings.TrimSuffix(strings.TrimPrefix(resourcePath, "/"), "/")
	gvk := strings.Split(path, "/")
	if len(gvk) > 3 {
		return gvk[0], gvk[1], gvk[2], gvk[3]
	}

	return gvk[0], gvk[1], "", gvk[2]

}

func getMaxRetries(waitTime string) int {

	// Get max wait time and retries/interval
	maxWait, err := time.ParseDuration(waitTime)
	if err != nil {
		maxWait, _ = time.ParseDuration(defaultMaxWait)
	}
	return int(maxWait.Seconds()) / 5

}

func waitFor(prv provisioner.Provisioner, resources []loader.WaitFor, checkFunc func([]unstructured.Unstructured) bool) {

	for _, resource := range resources {

		limit := getMaxRetries(resource.Timeout)
		logrus.Debugf(
			"Waiting for resource %s, retrying every 5s for %d times",
			resource.Resource,
			limit,
		)

		version, kind, namespace, name := getResourceDataFromPath(resource.Resource)
		for counter := 0; counter < limit; counter++ {

			obj, _ := prv.ListWithSelectors(
				context.TODO(),
				version,
				kind,
				namespace,
				map[string]interface{}{
					"metadata.name": name,
				},
			)
			if checkFunc(obj.Items) {
				logrus.Debugf("waitFor: operation completed.")
				break
			}
			time.Sleep(5 * time.Second)
		}
	}
}
