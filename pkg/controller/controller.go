package controller

import (
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
		// wait for channel
		time.Sleep(wait)

		// for TestSuite in TestSuitesList
		// 		NEW GO ROUTINE ->
		/*
			//		for test in TestSuite.Tests:
			//			for object in test.Objects
			//				create(object)
			//				wait(object)
			//
			//			for assertion in test.Assertion
			//				validate it
			//
			//			for object in test.Objects
			//				delete(object)
		*/
	}
}
