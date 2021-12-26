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

func NewController(prv provisioner.Provisioner, ms *metrics.Server, a *assert.Assert) *Controller {
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
			errors := c.Setup(test.ObjectsList)

			// Wait for resources to be provisioned
			c.WaitFor("create", test.Setup.WaitFor)

			// Run the actual tests
			result := c.Assert.Run(test, errors)
			metricsValues = updateMetricsValues(
				metricsValues,
				test.Name,
				result,
			)

			// Delete resources and wait for deletion
			c.Teardown(test.ObjectsList)

			// Wait for resources to be deleted
			c.WaitFor("delete", test.Teardown.WaitFor)
		}

		logrus.Debug("Push new metrics to server")
		c.setMetrics(metricsValues)

		logrus.Infof("Waiting for next execution (%s)", wait)
		time.Sleep(wait)
	}

}

// Wait for a particular resource to be either deleted or created
func (c *Controller) WaitFor(ops string, resources []loader.WaitFor) {

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

			if ops == "create" {
				if len(obj.Items) != 0 {
					logrus.Debugf("Success object retrieved: %s", resource.Resource)
					break
				}
				logrus.Debugf("Can't retrieve object, retrying in 5s. Object: %s", resource.Resource)
			}

			if ops == "delete" {
				if len(obj.Items) == 0 {
					logrus.Debugf("Object deleted: %s", resource.Resource)
					break
				}
				logrus.Warningf("Object %s still exists, retrying in 5s. Object", resource.Resource)
			}
			time.Sleep(5 * time.Second)
		}
	}
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
