package controller

import (
	"context"
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

func (c *Controller) Run(testsList loader.TestSuitesList, wait time.Duration) {

	logrus.Info("Starting controller")
	for {
		for _, testSuite := range testsList.TestSuites {
			logrus.Info("Starting tests on test suite: ", testSuite.Name)
			for _, test := range testSuite.Tests {
				logrus.Info("Running test with ID: ", test.ID)
				errors := c.create(test)

				time.Sleep(test.Timeout * time.Second)
				c.execTest(test.Assert, errors)
				c.delete(test)
				time.Sleep(test.Timeout * time.Second)
			}
		}
		time.Sleep(wait)
	}
}

func (c *Controller) execTest(assertion loader.Assertion, errors []error) bool {
	objects, err := c.Provisioner.ListWithSelectors(
		context.TODO(),
		assertion.ApiVersion,
		assertion.Kind,
		assertion.Namespace,
		assertion.Selectors,
	)
	if err != nil {
		logrus.Debugln(err)
		return false
	}

	logrus.Debugln("Retrieved objects from listWithSelectors() ", objects)

	return true
}

// Create resources defined on manifests
func (c *Controller) create(test loader.TestDefinition) []error {

	var errors []error

	for _, object := range test.ObjectsList {
		_, err := c.Provisioner.CreateOrUpdate(context.TODO(), object)
		if err != nil {
			logrus.Debugf("Couldn't delete resource %s", object.Object.GetName())
			logrus.Debugln(err)
			errors = append(errors, err)
		} else {
			logrus.Debugf("Resource created %s\n", object.Object.GetName())
		}
	}

	return errors
}

// Delete resources defined on manifests
func (c *Controller) delete(test loader.TestDefinition) []error {

	var errors []error

	for _, object := range test.ObjectsList {
		_, err := c.Provisioner.Delete(context.TODO(), object)
		if err != nil {
			logrus.Debugf("Couldn't delete resource %s", object.Object.GetName())
			logrus.Debugln(err)
			errors = append(errors, err)
		} else {
			logrus.Debugf("Resource deleted %s\n", object.Object.GetName())
		}
	}

	return errors
}
