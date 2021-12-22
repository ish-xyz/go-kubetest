package controller

import "github.com/ish-xyz/go-kubetest/pkg/provisioner"

type Controller struct {
	Provisioner *provisioner.Provisioner
}

type AssertionResult struct {
	ID       int
	Type     string
	TestName string
	Message  string
	Passed   bool
}
