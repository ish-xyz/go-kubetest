package controller

import (
	"context"
	"testing"

	"github.com/ish-xyz/go-kubetest/pkg/provisioner"
	"github.com/stretchr/testify/assert"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

/*
Make controller.Run() testable with stopCh
*/

func TestSetup(t *testing.T) {

	// Prepare mock
	argObj := &unstructured.Unstructured{
		Object: map[string]interface{}{
			"metadata": map[string]interface{}{
				"name": "MockTest",
			},
		},
	}
	objects := make([]*unstructured.Unstructured, 1)
	prvMock := new(provisioner.ProvisionerMock)
	prvMock.On("CreateOrUpdate", context.TODO(), argObj).Return(nil)

	// Run tests
	ctrl := NewController(prvMock, nil, nil)
	err := ctrl.Setup(objects)

	assert.Nil(t, err)
	prvMock.AssertNumberOfCalls(t, "CreateOrUpdate", 1)
}
