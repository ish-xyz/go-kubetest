package controller

import (
	"github.com/ish-xyz/go-kubetest/pkg/assert"
	"github.com/ish-xyz/go-kubetest/pkg/loader"
	"github.com/ish-xyz/go-kubetest/pkg/metrics"
	"github.com/ish-xyz/go-kubetest/pkg/provisioner"
)

type Controller struct {
	Loader            loader.Loader
	Provisioner       provisioner.Provisioner
	MetricsController *metrics.MetricsController
	Assert            *assert.Assert
}
