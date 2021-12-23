package controller

import (
	"context"
	"fmt"
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

func (c *Controller) Run(testsList []loader.TestDefinition, wait time.Duration) {

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
			c.waitForResources(test)

			// Run the actual tests
			result := c.Assert.Run(test, errors)
			metricsValues = updateMetricsValues(metricsValues, result)

			// Delete resources and wait for deletion
			c.teardown(test.ObjectsList)

			// Wait for resources to be deleted
			c.waitForResources(test)
		}
		logrus.Debug("Push new metrics to server")
		c.serveMetrics(metricsValues)

		logrus.Infof("Waiting for next execution (%s)", wait)
		time.Sleep(wait)
	}

}

func (c *Controller) waitForResources(test loader.TestDefinition) {
	timeout, err := time.ParseDuration(test.MaxWait) // TODO: change from "sleep for $MaxWait" to retry with timeout $MaxWait
	if err != nil {
		logrus.Warnf("Couldn't parse test timeout, defaulting to %s", defaultMaxWait)
		timeout, _ = time.ParseDuration(defaultMaxWait)
	}
	time.Sleep(timeout)
}

// Create resources defined on manifests
func (c *Controller) setup(objects []loader.LoadedObject) []string {

	var errors []string

	for _, obj := range objects {
		_, err := c.Provisioner.CreateOrUpdate(context.TODO(), obj)
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
func (c *Controller) teardown(objects []loader.LoadedObject) {

	for _, obj := range objects {
		_, err := c.Provisioner.Delete(context.TODO(), obj)
		if err != nil {
			logrus.Errorf("Teardown: Couldn't delete resource %s", obj.Object.GetName())
			logrus.Errorln(err)
		} else {
			logrus.Infof("Teardown: Resource deleted %s\n", obj.Object.GetName())
		}
	}
}

func updateMetricsValues(metricsValues *metrics.MetricsValues, result assert.TestResult) *metrics.MetricsValues {
	metricsValues.TotalTests += 1

	// TODO this is ugly: init map if nil (only for first execution)
	if metricsValues.TestStatus == nil {
		metricsValues.TestStatus = map[string]float64{}
	}

	if result.Passed {
		metricsValues.TotalTestsPassed += 1
		metricsValues.TestStatus[result.Name] = 1
	} else {
		metricsValues.TotalTestsFailed += 1
		metricsValues.TestStatus[result.Name] = 0
	}

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
