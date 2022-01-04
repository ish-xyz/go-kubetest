package controller

import (
	"context"
	"errors"
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

// Return a new instance for controller
func NewController(ldr loader.Loader, prv provisioner.Provisioner, ms *metrics.Server, a *assert.Assert) *Controller {
	return &Controller{
		Loader:        ldr,
		Provisioner:   prv,
		MetricsServer: ms,
		Assert:        a,
	}
}

// Start controller for periodically tests executions
func (c *Controller) Run(
	ctx context.Context,
	namespace string,
	selectors map[string]interface{},
	wait time.Duration,
	once bool,
) error {

	testsList, err := c.Loader.LoadTests(namespace, selectors)
	if err != nil {
		logrus.Fatal(err)
		return err
	}

	logrus.Infof("Starting metrics server at :%d", c.MetricsServer.Port)
	go c.MetricsServer.Serve()

	logrus.Info("Starting controller")
	for {
		metricsValues := metrics.NewMetricsValues()
		for _, test := range testsList {
			logrus.Infof("Running test: '%s'", test.Name)

			// Create resources and wait for creation
			errors := c.Setup(ctx, test.ObjectsList)
			if !c.WaitForCreation(ctx, test.Setup.WaitFor) {
				logrus.Errorf("Error while waiting for resource/s to be created, skipping test '%s'", test.Name)
				metricsValues.Store(test.Name, false, nil)
				continue
			}

			// Run the actual tests
			result, asrtRes := c.Assert.Run(test, errors)
			metricsValues.Store(test.Name, result, asrtRes)

			// Delete resources and wait for deletion
			c.Teardown(ctx, test.ObjectsList)
			if !c.WaitForDeletion(ctx, test.Teardown.WaitFor) {
				logrus.Errorf("Error while waiting for resource/s to be deleted, test: '%s'", test.Name)
				continue
			}
		}

		logrus.Debug("Push new metrics to server")
		metricsValues.Publish(c.MetricsServer)

		logrus.Infof("Waiting for next execution (%s)", wait)
		if once {
			return nil
		}
		time.Sleep(wait)
	}
}

// WaitForCreation wait until a set of resources has been created
func (c *Controller) WaitForCreation(ctx context.Context, resources []loader.WaitFor) bool {

	for _, resource := range resources {

		gvkData, err := getResourceDataFromPath(resource.Resource)
		if err != nil {
			logrus.Debugf("%v", err)
			return false
		}
		created, interval := false, 2
		limit := getMaxRetries(resource.Timeout, interval)

		logrus.Debugf("Waiting for resource %s, retrying every %ds for %d times", resource.Resource, interval, limit)
		for counter := 0; counter < limit; counter++ {

			obj, _ := c.Provisioner.ListWithSelectors(
				ctx,
				gvkData,
				map[string]interface{}{
					"metadata.name": gvkData["name"],
				},
			)
			if len(obj.Items) != 0 {
				created = true
				logrus.Debugf("resource %s has been created.", resource.Resource)
				break
			}
			time.Sleep(time.Duration(interval) * time.Second)
		}
		if !created {
			return false
		}
	}
	return true
}

// WaitForDeletion wait until a set of resources has been deleted
func (c *Controller) WaitForDeletion(ctx context.Context, resources []loader.WaitFor) bool {

	for _, resource := range resources {

		gvkData, err := getResourceDataFromPath(resource.Resource)
		if err != nil {
			logrus.Debugf("%v", err)
			return false
		}
		deleted, interval := false, 2
		limit := getMaxRetries(resource.Timeout, interval)

		logrus.Debugf("Waiting for resource %s, retrying every %ds for %d times", resource.Resource, interval, limit)
		for counter := 0; counter < limit; counter++ {

			obj, _ := c.Provisioner.ListWithSelectors(
				ctx,
				gvkData,
				map[string]interface{}{
					"metadata.name": gvkData["name"],
				},
			)
			if len(obj.Items) == 0 {
				logrus.Debugf("resource %s has been deleted.", resource.Resource)
				deleted = true
				break
			}
			time.Sleep(time.Duration(interval) * time.Second)
		}
		if !deleted {
			return false
		}
	}

	return true
}

// Create resources defined on manifests
func (c *Controller) Setup(ctx context.Context, objects []*unstructured.Unstructured) []string {

	var errors []string

	for _, obj := range objects {
		err := c.Provisioner.CreateOrUpdate(ctx, obj)
		if err != nil {
			logrus.Debugf("Couldn't create resource %s", obj.GetName())
			logrus.Debugln(err)
			errors = append(errors, fmt.Sprintf("%v", err))
			continue
		}
		logrus.Debugf("Setup: resource created %s\n", obj.GetName())
	}

	return errors
}

// Delete resources defined on manifests
func (c *Controller) Teardown(ctx context.Context, objects []*unstructured.Unstructured) []string {

	var errors []string

	for index := range objects {
		// Teardown needs to delete the objects in the
		// manifest, from the last one to the first one
		obj := objects[len(objects)-1-index]
		err := c.Provisioner.Delete(ctx, obj)
		if err != nil {
			logrus.Debugf("Couldn't delete resource %s", obj.GetName())
			logrus.Debugln(err)
			errors = append(errors, fmt.Sprintf("%v", err))
			continue
		}
		logrus.Debugf("Teardown: Resource deleted %s\n", obj.GetName())
	}
	return errors
}

func getResourceDataFromPath(resourcePath string) (map[string]string, error) {

	path := strings.TrimSuffix(strings.TrimPrefix(resourcePath, ":"), ":")
	gvk := strings.Split(path, ":")

	if len(gvk) < 3 {
		err := errors.New("can't retrieve gvk from resourcePath")
		return map[string]string{}, err
	}

	if len(gvk) > 3 {
		return map[string]string{
			"apiVersion": gvk[0],
			"kind":       gvk[1],
			"namespace":  gvk[2],
			"name":       gvk[3],
		}, nil
	}

	return map[string]string{
		"apiVersion": gvk[0],
		"kind":       gvk[1],
		"namespace":  "",
		"name":       gvk[2],
	}, nil

}

func getMaxRetries(waitTime string, interval int) int {

	// Get max wait time and retries/interval
	maxWait, err := time.ParseDuration(waitTime)
	if err != nil {
		maxWait, _ = time.ParseDuration(defaultMaxWait)
	}
	return int(maxWait.Seconds()) / interval

}
