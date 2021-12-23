package assert

import (
	"context"
	"fmt"
	"regexp"

	"github.com/ish-xyz/go-kubetest/pkg/loader"
	"github.com/ish-xyz/go-kubetest/pkg/provisioner"
	"github.com/sirupsen/logrus"
)

func NewAssert(prv *provisioner.Provisioner) *Assert {
	return &Assert{
		Provisioner: prv,
	}
}

func (a *Assert) Run(test loader.TestDefinition, errors []string) TestResult {

	testResult := &TestResult{
		Name:              test.Name,
		Passed:            true,
		AssertionsResults: []AssertionResult{},
	}

	testResult.AssertionsResults = append(
		testResult.AssertionsResults,
		testResult.expectedErrors(test.ExpectedErrors, errors)...,
	)
	testResult.AssertionsResults = append(
		testResult.AssertionsResults,
		testResult.expectedObjects(a.Provisioner, test)...,
	)

	return *testResult
}

// Check if the expected Errors match the actual errors
func (t *TestResult) expectedErrorMsgs(expErrors, actErrors []string) []AssertionResult {

	var result []AssertionResult

	for index, errorMessage := range expErrors {
		assertRes := AssertionResult{
			ID:      index,
			Type:    "expected_errors",
			Passed:  true,
			Message: "OK",
		}
		match, _ := regexp.MatchString(errorMessage, actErrors[index])
		if !match {
			t.Passed = false
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
func (t *TestResult) expectedErrors(expErrors, actErrors []string) []AssertionResult {

	var result []AssertionResult

	assertRes := AssertionResult{
		ID:      0,
		Type:    "expected_errors_count",
		Message: "OK",
		Passed:  true,
	}
	if len(expErrors) != len(actErrors) {
		t.Passed = false
		assertRes.Passed = false
		assertRes.Message = fmt.Sprintf(
			"expected %d errors, got %d instead.",
			len(expErrors),
			len(actErrors),
		)
	} else {
		result = append(result, t.expectedErrorMsgs(expErrors, actErrors)...)
	}

	result = append(result, assertRes)

	return result
}

// Check if the retrieved objects match the expected count
func (t *TestResult) expectedObjects(prv *provisioner.Provisioner, test loader.TestDefinition) []AssertionResult {

	var result []AssertionResult

	for index, assertion := range test.Assert {

		assertRes := AssertionResult{
			ID:      index,
			Type:    "expected_objects",
			Message: "OK",
			Passed:  true,
		}

		objects, err := prv.ListWithSelectors(
			context.TODO(),
			assertion.ApiVersion,
			assertion.Kind,
			assertion.Namespace,
			assertion.Selectors,
		)
		if err != nil {
			t.Passed = false
			assertRes.Passed = false
			message := "Failed to retrieve objects, possible internal error"
			assertRes.Message = message
			logrus.Debugln(message)
			logrus.Debugln(err)
		} else {
			if len(objects.Items) != assertion.ExpectedResources {
				t.Passed = false
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
