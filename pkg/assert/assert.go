package assert

import (
	"context"
	"regexp"

	"github.com/ish-xyz/go-kubetest/pkg/loader"
	"github.com/ish-xyz/go-kubetest/pkg/provisioner"
)

func NewAssert(prv *provisioner.Provisioner) *Assert {
	return &Assert{
		Provisioner: prv,
	}
}

func (a *Assert) Run(test *loader.TestDefinition, errors []string) bool {

	testResult := true

	for _, assertion := range test.Assert {
		var assertRes bool
		if assertion.Type == "expectedResources" {
			assertRes = expectedResources(a.Provisioner, assertion)
		}

		if assertion.Type == "expectedErrors" {
			assertRes = expectedErrors(assertion.Errors, errors)
		}

		if !assertRes {
			testResult = false
		}
	}

	return testResult
}

func expectedErrors(expErrors, actErrors []string) bool {

	if len(expErrors) != len(actErrors) {
		return false
	}

	for index, errorMessage := range expErrors {
		match, _ := regexp.MatchString(errorMessage, actErrors[index])
		if !match {
			return false
		}
	}
	return true
}

// Check if the retrieved objects match the expected count
func expectedResources(prv *provisioner.Provisioner, assertion loader.Assertion) bool {

	objects, err := prv.ListWithSelectors(
		context.TODO(),
		assertion.ApiVersion,
		assertion.Kind,
		assertion.Namespace,
		assertion.Selectors,
	)
	if err != nil {
		return false
	}
	if len(objects.Items) != assertion.Count {
		return false
	}

	return true
}
