package controller

import (
	"context"
	"errors"
	"testing"

	"github.com/ish-xyz/go-kubetest/pkg/provisioner"
	"github.com/stretchr/testify/assert"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

func TestSetup(t *testing.T) {

	// Prepare test data & mock
	testedMethod := "CreateOrUpdate"
	objects := make([]*unstructured.Unstructured, 1)
	objects[0] = &unstructured.Unstructured{
		Object: map[string]interface{}{
			"metadata": map[string]interface{}{
				"name": "MockTest",
			},
		},
	}
	prvMock := new(provisioner.ProvisionerMock)
	prvMock.On(testedMethod, context.TODO(), objects[0]).Return(nil)

	// Run tests
	ctrl := NewController(prvMock, nil, nil)
	errors := ctrl.Setup(objects)

	assert.Len(t, errors, 0)
	prvMock.AssertNumberOfCalls(t, testedMethod, 1)
}

func TestSetupWithErrors(t *testing.T) {

	// Prepare test data & mock
	testedMethod := "CreateOrUpdate"
	objects := make([]*unstructured.Unstructured, 1)
	objects[0] = &unstructured.Unstructured{
		Object: map[string]interface{}{
			"metadata": map[string]interface{}{
				"name": "MockTest",
			},
		},
	}
	prvMock := new(provisioner.ProvisionerMock)
	prvMock.On(testedMethod, context.TODO(), objects[0]).Return(
		errors.New("failed to create"),
	)

	// Run tests
	ctrl := NewController(prvMock, nil, nil)
	errors := ctrl.Setup(objects)

	assert.Len(t, errors, 1)
	prvMock.AssertNumberOfCalls(t, testedMethod, 1)
}

func TestTeardown(t *testing.T) {

	// Prepare test data & mock
	testedMethod := "Delete"
	objects := make([]*unstructured.Unstructured, 1)
	objects[0] = &unstructured.Unstructured{
		Object: map[string]interface{}{
			"metadata": map[string]interface{}{
				"name": "MockTest",
			},
		},
	}
	prvMock := new(provisioner.ProvisionerMock)
	prvMock.On(testedMethod, context.TODO(), objects[0]).Return(nil)

	// Run tests
	ctrl := NewController(prvMock, nil, nil)
	errors := ctrl.Teardown(objects)

	assert.Len(t, errors, 0)
	prvMock.AssertNumberOfCalls(t, testedMethod, 1)
}

func TestTeardownWithErrors(t *testing.T) {

	// Prepare test data & mock
	testedMethod := "Delete"
	objects := make([]*unstructured.Unstructured, 1)
	objects[0] = &unstructured.Unstructured{
		Object: map[string]interface{}{
			"metadata": map[string]interface{}{
				"name": "MockTest",
			},
		},
	}
	prvMock := new(provisioner.ProvisionerMock)
	prvMock.On(testedMethod, context.TODO(), objects[0]).Return(
		errors.New("failed to delete"),
	)

	// Run tests
	ctrl := NewController(prvMock, nil, nil)
	errors := ctrl.Teardown(objects)

	assert.Len(t, errors, 1)
	prvMock.AssertNumberOfCalls(t, testedMethod, 1)
}
