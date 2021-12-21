package controller

import (
	"context"
	"fmt"
	"time"

	"github.com/ish-xyz/go-kubetest/pkg/loader"
	"github.com/ish-xyz/go-kubetest/pkg/provisioner"
	"github.com/sirupsen/logrus"
)

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
			fmt.Println(c.Assert(test, errors))

			// Delete resources and wait for deletion
			c.delete(test.ObjectsList)
			time.Sleep(test.Timeout * time.Second)
		}
		logrus.Infof("Waiting for next execution (%s)", wait)
		time.Sleep(wait)
	}
}

func (c *Controller) Assert(test loader.TestDefinition, errors []error) TestReport {

	var result TestReport
	assertions := test.Assert

	for _, assertion := range assertions {
		objects, err := c.Provisioner.ListWithSelectors(
			context.TODO(),
			assertion.ApiVersion,
			assertion.Kind,
			assertion.Namespace,
			assertion.Selectors,
		)
		if err != nil {
			logrus.Debugln(err)
			result.Failed += 1
			result.ErrorMessages = append(
				result.ErrorMessages,
				fmt.Sprintf("%v", err),
			)
			continue
		}
		if len(objects.Items) != assertion.Count {
			result.Failed += 1
			result.ErrorMessages = append(
				result.ErrorMessages,
				fmt.Sprintf(
					"Failed test %s-0 expected %d resources, got %d instead.",
					test.Name,
					assertion.Count,
					len(objects.Items),
				),
			)
			logrus.Warnf("Failed test %s-0 expected %d resources, got %d instead.", test.Name, assertion.Count, len(objects.Items))
		} else {
			result.Passed += 1
			logrus.Debugf("Passed test %s-0 expected %d resources", test.Name, assertion.Count)
		}

		if len(errors) != test.ExpectedErrors {
			result.Failed += 1
			result.ErrorMessages = append(
				result.ErrorMessages,
				fmt.Sprintf(
					"Failed test %s-0 expected %d errors, got %d instead.",
					test.Name,
					test.ExpectedErrors,
					len(errors),
				),
			)
			logrus.Warnf("Failed test %s-0 expected %d errors, got %d instead.", test.Name, test.ExpectedErrors, len(errors))
		} else {
			result.Passed += 1
			logrus.Debugf("Passed test %s-0 expected %d errors", test.Name, test.ExpectedErrors)
		}
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
