package controller

import (
	"context"
	"fmt"
	"time"

	"github.com/ish-xyz/go-kubetest/pkg/loader"
	"github.com/ish-xyz/go-kubetest/pkg/provisioner"
	"github.com/sirupsen/logrus"
)

type Controller struct {
	Provisioner *provisioner.Provisioner
}

func NewController(prv *provisioner.Provisioner) *Controller {
	return &Controller{
		Provisioner: prv,
	}
}

func (c *Controller) Run(testsList []loader.TestDefinition, wait time.Duration) {

	logrus.Info("Starting controller")
	for {
		for _, test := range testsList {
			logrus.Infof("Running test: '%s'", test.Name)

			// Create resources and wait for creation
			errors := c.create(test.ObjectsList)
			time.Sleep(test.Timeout * time.Second)

			// Run the actual tests
			fmt.Println(c.execTest(test, errors))

			// Delete resources and wait for deletion
			c.delete(test.ObjectsList)
			time.Sleep(test.Timeout * time.Second)
		}
		logrus.Infof("Waiting for next execution", wait)
		time.Sleep(wait)
	}
}

func (c *Controller) execTest(test loader.TestDefinition, errors []error) []bool {

	var result []bool
	assertions := test.Assert
	objects, err := c.Provisioner.ListWithSelectors(
		context.TODO(),
		assertions[0].ApiVersion,
		assertions[0].Kind,
		assertions[0].Namespace,
		assertions[0].Selectors,
	)
	if err != nil {
		logrus.Debugln(err)
		return []bool{false}
	}

	if len(objects.Items) != assertions[0].Count {
		result = append(result, false)
		logrus.Warnf("Failed test %s-0 expected %d resources, got %d instead.", test.Name, assertions[0].Count, len(objects.Items))
	} else {
		result = append(result, true)
		logrus.Debugf("Passed test %s-0 expected %d resources", test.Name, assertions[0].Count)
	}

	if len(errors) != test.ExpectedErrors {
		result = append(result, false)
		logrus.Warnf("Failed test %s-0 expected %d errors, got %d instead.", test.Name, test.ExpectedErrors, len(errors))
	} else {
		result = append(result, true)
		logrus.Debugf("Passed test %s-0 expected %d errors", test.Name, test.ExpectedErrors)
	}
	return result
}

// Create resources defined on manifests
func (c *Controller) create(objects []loader.LoadedObject) []error {

	var errors []error

	for _, obj := range objects {
		_, err := c.Provisioner.CreateOrUpdate(context.TODO(), obj)
		if err != nil {
			logrus.Debugf("Couldn't delete resource %s", obj.Object.GetName())
			logrus.Debugln(err)
			errors = append(errors, err)
		} else {
			logrus.Debugf("Resource created %s\n", obj.Object.GetName())
		}
	}

	return errors
}

// Delete resources defined on manifests
func (c *Controller) delete(objects []loader.LoadedObject) []error {

	var errors []error

	for _, obj := range objects {
		_, err := c.Provisioner.Delete(context.TODO(), obj)
		if err != nil {
			logrus.Debugf("Couldn't delete resource %s", obj.Object.GetName())
			logrus.Debugln(err)
			errors = append(errors, err)
		} else {
			logrus.Debugf("Resource deleted %s\n", obj.Object.GetName())
		}
	}

	return errors
}
