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
)

const defaultMaxWait = "60s"

func NewController(prv *provisioner.Provisioner, ms *metrics.Server, a *assert.Assert) *Controller {
	return &Controller{
		Provisioner:   prv,
		MetricsServer: ms,
		Assert:        a,
	}
}

func (c *Controller) Run(testsList []*loader.TestDefinition, wait time.Duration) {

	logrus.Infof("Starting metrics server at :%d", c.MetricsServer.Port)
	go c.MetricsServer.Serve()

	logrus.Info("Starting controller")
	for {

		metricsValues := metrics.NewMetricsValues()
		for _, test := range testsList {
			logrus.Infof("Running test: '%s'", test.Name)

			// Create resources and wait for creation
			errors := c.setup(test.ObjectsList)

			// Wait for resources to be provisioned
			c.waitForResources(test.Setup.WaitFor)

			// Run the actual tests
			result := c.Assert.Run(test, errors)
			metricsValues = updateMetricsValues(
				metricsValues,
				test.Name,
				result,
			)

			// Delete resources and wait for deletion
			c.teardown(test.ObjectsList)

			// Wait for resources to be deleted
			c.waitForResources(test.Teardown.WaitFor)
		}
		logrus.Debug("Push new metrics to server")
		c.serveMetrics(metricsValues)

		logrus.Infof("Waiting for next execution (%s)", wait)
		time.Sleep(wait)
	}

}

func (c *Controller) waitForResources(resources []loader.WaitFor) {

	for _, resource := range resources {

		// Parse wait time
		maxWait, err := time.ParseDuration(resource.Timeout)
		if err != nil {
			maxWait, _ = time.ParseDuration(defaultMaxWait)
		}
		limit := int(maxWait.Seconds()) / 5
		logrus.Debugf(
			"Waiting for resource %s, with max timeout %ds, and max retries %d",
			resource.Resource,
			int(maxWait.Seconds()),
			limit,
		)

		// Get Resource Path:
		// Let's first assume is a cluster-wide
		// resource and if the namespace is
		// defined in the resourcePath switch
		// to namespaced
		resourcePath := strings.TrimSuffix(strings.TrimPrefix(resource.Resource, "/"), "/")
		gvk := strings.Split(resourcePath, "/")
		version := gvk[0]
		kind := gvk[1]
		namespace := ""
		name := gvk[2]
		if len(gvk) > 3 {
			namespace = gvk[2]
			name = gvk[3]
		}

		for counter := 0; counter < limit; counter++ {
			obj, _ := c.Provisioner.ListWithSelectors(
				context.TODO(),
				version,
				kind,
				namespace,
				map[string]interface{}{
					"metadata.name": name,
				},
			)
			if obj != nil {
				logrus.Debugf("Success object retrieved: %s", resource.Resource)
				break
			} else {
				logrus.Warningf("Can't retrieve object, retrying in 5s. Object: %s", resource.Resource)
			}
			time.Sleep(5 * time.Second)
		}
	}
}

// Create resources defined on manifests
func (c *Controller) setup(objects []*loader.LoadedObject) []string {

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
func (c *Controller) teardown(objects []*loader.LoadedObject) {

	for _, obj := range objects {
		err := c.Provisioner.Delete(context.TODO(), obj)
		if err != nil {
			logrus.Errorf("Teardown: Couldn't delete resource %s", obj.Object.GetName())
			logrus.Errorln(err)
		} else {
			logrus.Debugf("Teardown: Resource deleted %s\n", obj.Object.GetName())
		}
	}
}

// Update a temporary struct that will then used to push metrics in one go
func updateMetricsValues(metricsValues *metrics.MetricsValues, testName string, result bool) *metrics.MetricsValues {

	if result {
		metricsValues.TotalTestsPassed += 1
		metricsValues.TestStatus[testName] = 1
	} else {
		metricsValues.TotalTestsFailed += 1
		metricsValues.TestStatus[testName] = 0
	}

	metricsValues.TotalTests += 1

	return metricsValues
}

func (c *Controller) serveMetrics(metricsValues *metrics.MetricsValues) {

	rMetrics := c.MetricsServer.Metrics

	for key, value := range metricsValues.TestStatus {
		rMetrics.TestStatus.WithLabelValues(key).Set(value)
	}

	rMetrics.TotalTestsFailed.Set(metricsValues.TotalTestsFailed)
	rMetrics.TotalTestsPassed.Set(metricsValues.TotalTestsPassed)
	rMetrics.TotalTests.Set(metricsValues.TotalTests)
}
