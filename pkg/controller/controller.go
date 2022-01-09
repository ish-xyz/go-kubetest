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
func NewController(
	ldr loader.Loader,
	prv provisioner.Provisioner,
	mc *metrics.MetricsController,
	a *assert.Assert,
) *Controller {

	return &Controller{
		Loader:            ldr,
		Provisioner:       prv,
		MetricsController: mc,
		Assert:            a,
	}
}

// Start controller for periodically tests executions
func (ctrl *Controller) Run(
	ctx context.Context,
	namespace string,
	selectors map[string]interface{},
	wait time.Duration,
	once bool,
) error {

	testsList, err := ctrl.Loader.LoadTests(namespace, selectors)
	if err != nil {
		logrus.Fatal(err)
		return err
	}

	if !once {
		logrus.Infof("Starting metrics server at :%d", ctrl.MetricsController.Port)
		go ctrl.MetricsController.Run(namespace)
	}

	logrus.Info("Starting controller")
	for {

		for _, test := range testsList {
			logrus.Infof("Running test: '%s'", test.Name)

			// Create resources and wait for creation
			errors := ctrl.Setup(ctx, test.ObjectsList)
			if !ctrl.WaitForCreation(ctx, test.Setup.WaitFor) {
				logrus.Errorf("Error while waiting for resource/s to be created, skipping test '%s'", test.Name)
				ctrl.CreateTestResult(ctx, test.Name, false, nil)
				continue
			}

			// Run the actual tests and store results
			result, asrtRes := ctrl.Assert.Run(test, errors)
			err = ctrl.CreateTestResult(ctx, test.Name, result, asrtRes)
			if err != nil {
				logrus.Warningf("error creating test results %v", err)
			}

			// Delete resources and wait for deletion
			ctrl.Teardown(ctx, test.ObjectsList)
			if !ctrl.WaitForDeletion(ctx, test.Teardown.WaitFor) {
				logrus.Errorf("Error while waiting for resource/s to be deleted, test: '%s'", test.Name)
				continue
			}
		}

		logrus.Debug("Push new metrics to server")

		if once {
			logrus.Infof("Tests finished, results have been created")
			return nil
		}
		logrus.Infof("Waiting for next execution (%s)", wait)
		time.Sleep(wait)
	}
}

// WaitForCreation wait until a set of resources has been created
func (ctrl *Controller) WaitForCreation(ctx context.Context, resources []loader.WaitFor) bool {

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

			obj, _ := ctrl.Provisioner.ListWithSelectors(
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
func (ctrl *Controller) WaitForDeletion(ctx context.Context, resources []loader.WaitFor) bool {

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

			obj, _ := ctrl.Provisioner.ListWithSelectors(
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
func (ctrl *Controller) Setup(ctx context.Context, objects []*unstructured.Unstructured) []string {

	var errors []string

	for _, obj := range objects {
		err := ctrl.Provisioner.CreateOrUpdate(ctx, obj)
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
func (ctrl *Controller) Teardown(ctx context.Context, objects []*unstructured.Unstructured) []string {

	var errors []string

	for index := range objects {
		// Teardown needs to delete the objects in the
		// manifest, from the last one to the first one
		obj := objects[len(objects)-1-index]
		err := ctrl.Provisioner.Delete(ctx, obj)
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

// CreateTestResult resource
func (ctrl *Controller) CreateTestResult(ctx context.Context, name string, result bool, asrtRes map[string]interface{}) error {

	obj := &unstructured.Unstructured{
		Object: map[string]interface{}{
			"apiVersion": "go-kubetest.io/v1",
			"kind":       "TestResult",
			"metadata": map[string]interface{}{
				"name": name,
			},
			"spec": map[string]interface{}{
				"result":     result,
				"assertions": asrtRes,
			},
		},
	}

	err := ctrl.Provisioner.CreateOrUpdate(ctx, obj)
	return err
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
