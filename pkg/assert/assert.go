package assert

import (
	"context"
	"regexp"

	"github.com/ish-xyz/go-kubetest/pkg/loader"
	"github.com/ish-xyz/go-kubetest/pkg/provisioner"
	"github.com/sirupsen/logrus"
)

func NewAssert(prv provisioner.Provisioner) *Assert {
	return &Assert{
		Provisioner: prv,
	}
}

func (a *Assert) Run(test *loader.TestDefinition, errors []string) (bool, map[string]bool) {

	testResult := true
	assertResults := map[string]bool{}
	for _, assertion := range test.Assert {
		var assertRes bool

		switch assertion.Type {
		case "expectedResources":
			assertRes = expectedResources(a.Provisioner, assertion)
		case "expectedErrors":
			assertRes = expectedErrors(assertion.Errors, errors)
		}

		if !assertRes {
			testResult = false
		}

		assertResults[assertion.Name] = assertRes
	}

	// TODO return bool []map[string]bool
	return testResult, assertResults
}

// Check if the errors throwed during setup are expected or not
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
func expectedResources(prv provisioner.Provisioner, assertion loader.Assertion) bool {

	apiVersion, kind, namespace := unpackResource(assertion.Resource)
	limit := getMaxRetries(assertion.Timeout, 2)
	passed := false

	for x := 0; x < limit; x++ {

		objects, err := prv.ListWithSelectors(
			context.TODO(),
			map[string]string{
				"apiVersion": apiVersion,
				"kind":       kind,
				"namespace":  namespace,
			},
			assertion.Selectors,
		)
		if err != nil {
			logrus.Debugln(err)
			continue
		}
		if len(objects.Items) != assertion.Count {
			logrus.Debugln(err)
			continue
		}
		passed = true
		break
	}

	return passed
}
