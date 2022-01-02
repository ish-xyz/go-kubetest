package controller

import (
	"context"
	"errors"
	"testing"

	"github.com/ish-xyz/go-kubetest/pkg/loader"
	"github.com/ish-xyz/go-kubetest/pkg/provisioner"
	"github.com/stretchr/testify/assert"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

func TestGetResourceDataFromPath(t *testing.T) {

	gvkData, err := getResourceDataFromPath("v1/Namespace/namespace-1")

	assert.Nil(t, err)
	assert.Equal(t, gvkData["version"], "v1")
	assert.Equal(t, gvkData["kind"], "Namespace")
	assert.Equal(t, gvkData["namespace"], "")
	assert.Equal(t, gvkData["name"], "namespace-1")

}

func TestGetResourceDataFromPathErrors(t *testing.T) {

	_, err := getResourceDataFromPath("wrong resource path")

	assert.NotNil(t, err)

}

func TestGetMaxRetriesErrors(t *testing.T) {

	// Will default to 60s
	limit := getMaxRetries("wrongString", 6)

	assert.Equal(t, limit, 10)
}

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
	ctrl := NewController(nil, prvMock, nil, nil)
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
	ctrl := NewController(nil, prvMock, nil, nil)
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
	ctrl := NewController(nil, prvMock, nil, nil)
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
	ctrl := NewController(nil, prvMock, nil, nil)
	errors := ctrl.Teardown(objects)

	assert.Len(t, errors, 1)
	prvMock.AssertNumberOfCalls(t, testedMethod, 1)
}

func TestWaitForDeletion(t *testing.T) {

	//Prepare test data & mock
	testedMethod := "ListWithSelectors"
	resources := make([]loader.WaitFor, 1)
	resources[0] = loader.WaitFor{
		Resource: "v1/Namespace/namespace-1",
		Timeout:  "5s",
	}
	returnObject := &unstructured.UnstructuredList{
		Items: []unstructured.Unstructured{},
	}
	prvMock := new(provisioner.ProvisionerMock)
	prvMock.On(
		testedMethod,
		context.TODO(),
		map[string]string{"kind": "Namespace", "name": "namespace-1", "namespace": "", "version": "v1"},
		map[string]interface{}{
			"metadata.name": "namespace-1",
		}).Return(returnObject, nil)

	// Run tests
	ctrl := NewController(nil, prvMock, nil, nil)
	result := ctrl.WaitForDeletion(resources)

	assert.True(t, result)
	prvMock.AssertNumberOfCalls(t, testedMethod, 1)

}

func TestWaitForCreation(t *testing.T) {

	//Prepare test data & mock
	testedMethod := "ListWithSelectors"
	resources := make([]loader.WaitFor, 1)
	resources[0] = loader.WaitFor{
		Resource: "v1/Namespace/namespace-1",
		Timeout:  "5s",
	}
	returnObject := &unstructured.UnstructuredList{
		Items: []unstructured.Unstructured{
			{
				Object: map[string]interface{}{
					"metadata": map[string]interface{}{
						"name": "MockTest",
					},
				},
			},
		},
	}

	prvMock := new(provisioner.ProvisionerMock)
	prvMock.On(
		testedMethod,
		context.TODO(),
		map[string]string{"kind": "Namespace", "name": "namespace-1", "namespace": "", "version": "v1"},
		map[string]interface{}{
			"metadata.name": "namespace-1",
		}).Return(returnObject, nil)

	// Run tests
	ctrl := NewController(nil, prvMock, nil, nil)
	result := ctrl.WaitForCreation(resources)

	assert.True(t, result)
	prvMock.AssertNumberOfCalls(t, testedMethod, 1)

}

func TestWaitForDeletionErrors(t *testing.T) {

	//Prepare test data & mock
	testedMethod := "ListWithSelectors"
	resources := make([]loader.WaitFor, 1)
	resources[0] = loader.WaitFor{
		Resource: "v1/Namespace/namespace-1",
		Timeout:  "5s",
	}
	returnObject := &unstructured.UnstructuredList{
		Items: []unstructured.Unstructured{
			{
				Object: map[string]interface{}{
					"metadata": map[string]interface{}{
						"name": "MockTest",
					},
				},
			},
		},
	}

	prvMock := new(provisioner.ProvisionerMock)
	prvMock.On(
		testedMethod,
		context.TODO(),
		map[string]string{"kind": "Namespace", "name": "namespace-1", "namespace": "", "version": "v1"},
		map[string]interface{}{
			"metadata.name": "namespace-1",
		}).Return(returnObject, errors.New("error retrieving object"))

	// Run tests
	ctrl := NewController(nil, prvMock, nil, nil)
	result := ctrl.WaitForDeletion(resources)

	assert.False(t, result)
	prvMock.AssertNumberOfCalls(t, testedMethod, 2)

}

func TestWaitForCreationErrors(t *testing.T) {

	//Prepare test data & mock
	testedMethod := "ListWithSelectors"
	resources := make([]loader.WaitFor, 1)
	resources[0] = loader.WaitFor{
		Resource: "v1/Namespace/namespace-1",
		Timeout:  "5s",
	}
	returnObject := &unstructured.UnstructuredList{
		Items: []unstructured.Unstructured{},
	}

	prvMock := new(provisioner.ProvisionerMock)
	prvMock.On(
		testedMethod,
		context.TODO(),
		map[string]string{"kind": "Namespace", "name": "namespace-1", "namespace": "", "version": "v1"},
		map[string]interface{}{
			"metadata.name": "namespace-1",
		}).Return(returnObject, errors.New("error retrieving object"))

	// Run tests
	ctrl := NewController(nil, prvMock, nil, nil)
	result := ctrl.WaitForCreation(resources)

	assert.False(t, result)
	prvMock.AssertNumberOfCalls(t, testedMethod, 2)

}
