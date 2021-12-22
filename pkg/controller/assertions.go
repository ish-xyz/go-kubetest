package controller

import (
	"context"
	"fmt"
	"regexp"

	"github.com/ish-xyz/go-kubetest/pkg/loader"
	"github.com/sirupsen/logrus"
)

func (c *Controller) Assert(test loader.TestDefinition, errors []string) []AssertionResult {

	var result []AssertionResult

	result = append(
		result,
		c.expectedErrors(test.Name, test.ExpectedErrors, errors),
		c.expectedObjects(test.Name, test),
		c.expectedErrorsCount(test.Name, test.ExpectedErrors, errors)...,
	)

	return result
}

// Check if the expected Errors match the actual errors
func (c *Controller) expectedErrors(testName string, expErrors, actErrors []string) []AssertionResult {

	var result []AssertionResult

	for index, errorMessage := range expErrors {
		assertRes := AssertionResult{
			ID:       index,
			Type:     "expected_errors",
			TestName: testName,
			Passed:   true,
			Message:  "OK",
		}
		match, _ := regexp.MatchString(errorMessage, actErrors[index])
		if !match {
			assertRes.Passed = false
			assertRes.Message = fmt.Sprintf(
				"expected error: %s \nreturned error: %s",
				errorMessage,
				actErrors[index],
			)
		}
		result = append(result, assertRes)
	}
	return result
}

// Check if the number of errors occured during the resource creation is expected
func (c *Controller) expectedErrorsCount(testName string, expErrors, actErrors []string) []AssertionResult {
	//return len(expErrors) == len(actErrors)
	assertRes := AssertionResult{
		ID:       0,
		Type:     "expected_errors_count",
		TestName: testName,
		Message:  "OK",
		Passed:   true,
	}
	if len(expErrors) != len(actErrors) {
		assertRes.Message = fmt.Sprintf(
			"expected %d objects, got %d instead.",
			len(expErrors),
			len(actErrors),
		)
		assertRes.Passed = false
	}

	return []AssertionResult{assertRes}
}

// Check if the retrieved objects match the expected count
func (c *Controller) expectedObjects(testName string, test loader.TestDefinition) []AssertionResult {

	var result []AssertionResult

	for index, assertion := range test.Assert {

		assertRes := AssertionResult{
			ID:       index,
			Type:     "expected_objects",
			TestName: testName,
			Message:  "OK",
			Passed:   true,
		}

		objects, err := c.Provisioner.ListWithSelectors(
			context.TODO(),
			assertion.ApiVersion,
			assertion.Kind,
			assertion.Namespace,
			assertion.Selectors,
		)
		if err != nil {
			message := "Failed to retrieve objects, possible internal error"
			assertRes.Message = message
			assertRes.Passed = false
			logrus.Debugln(message)
			logrus.Debugln(err)
		} else {
			if len(objects.Items) != assertion.ExpectedResources {
				assertRes.Passed = false
				assertRes.Message = fmt.Sprintf(
					"expected %d objects, got %d instead.",
					assertion.ExpectedResources,
					len(objects.Items),
				)
			}
		}

		result = append(result, assertRes)
	}
	return result
}
