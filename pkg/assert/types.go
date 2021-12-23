package assert

import (
	"github.com/ish-xyz/go-kubetest/pkg/provisioner"
)

type TestResult struct {
	Name              string
	AssertionsResults []AssertionResult
	Passed            bool
}

type AssertionResult struct {
	ID      int
	Type    string
	Message string
	Passed  bool
}

type Assert struct {
	Provisioner *provisioner.Provisioner
}
