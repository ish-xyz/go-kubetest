package controller

import "github.com/ish-xyz/go-kubetest/pkg/provisioner"

type Controller struct {
	Provisioner *provisioner.Provisioner
}

type TestReport struct {
	Name          string
	Failed        int
	Passed        int
	ErrorMessages []string
}
