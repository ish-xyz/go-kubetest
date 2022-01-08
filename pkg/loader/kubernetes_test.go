package loader

import (
	"context"
	"errors"
	"testing"

	"github.com/ish-xyz/go-kubetest/pkg/provisioner"
	"github.com/stretchr/testify/assert"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

func TestK8SLoadManifestsErrors(t *testing.T) {

	// Prepare mock and data
	resourcePath := "namespace:name-of-resource"

	prvMock := new(provisioner.ProvisionerMock)
	prvMock.On(
		"ListWithSelectors",
		context.TODO(),
		map[string]string{
			"apiVersion": "go-kubetest.io/v1",
			"kind":       "TestResource",
			"namespace":  "namespace",
		},
		map[string]interface{}{
			"metadata.name": "name-of-resource",
		},
	).Return(&unstructured.UnstructuredList{}, errors.New("random error"))

	// Execute
	ldr := NewKubernetesLoader(prvMock)
	_, err := ldr.LoadManifests(resourcePath)

	// Assertions
	assert.NotNil(t, err)

	prvMock.AssertNumberOfCalls(t, "ListWithSelectors", 1)
}

func TestK8SLoadManifestsEmptyResult(t *testing.T) {

	// Prepare mock and data
	resourcePath := "namespace:name-of-resource"
	retObjects := &unstructured.UnstructuredList{
		Items: []unstructured.Unstructured{},
	}

	prvMock := new(provisioner.ProvisionerMock)
	prvMock.On(
		"ListWithSelectors",
		context.TODO(),
		map[string]string{
			"apiVersion": "go-kubetest.io/v1",
			"kind":       "TestResource",
			"namespace":  "namespace",
		},
		map[string]interface{}{
			"metadata.name": "name-of-resource",
		},
	).Return(retObjects, nil)

	// Execute
	ldr := NewKubernetesLoader(prvMock)
	_, err := ldr.LoadManifests(resourcePath)

	// Assertions
	assert.NotNil(t, err)
	prvMock.AssertNumberOfCalls(t, "ListWithSelectors", 1)
}

func TestK8SLoadManifests(t *testing.T) {

	// Prepare mock and data
	resourcePath := "namespace:name-of-resource"
	retObjects := &unstructured.UnstructuredList{
		Items: []unstructured.Unstructured{
			{
				Object: map[string]interface{}{
					"spec": map[string]interface{}{
						"data": "apiVersion: v1\nkind: Namespace\nmetadata:\n  labels:\n    myCustomLabel: myCustomValue\n    name: namespace-1\nspec: {}",
					},
				},
			},
		},
	}

	prvMock := new(provisioner.ProvisionerMock)
	prvMock.On(
		"ListWithSelectors",
		context.TODO(),
		map[string]string{
			"apiVersion": "go-kubetest.io/v1",
			"kind":       "TestResource",
			"namespace":  "namespace",
		},
		map[string]interface{}{
			"metadata.name": "name-of-resource",
		},
	).Return(retObjects, nil)

	// Execute
	ldr := NewKubernetesLoader(prvMock)
	res, err := ldr.LoadManifests(resourcePath)

	// Assertions
	assert.Nil(t, err)
	assert.Equal(t, res[0].GetKind(), "Namespace")
	assert.Equal(t, res[0].GetAPIVersion(), "v1")
	prvMock.AssertNumberOfCalls(t, "ListWithSelectors", 1)
}

func TesGetTestDefinition(t *testing.T) {

	inputData := map[string]interface{}{
		"apiVersion": "go-kubetest.io/v1",
		"kind":       "TestDefinition",
		"metadata": map[string]interface{}{
			"name": "test-input-data",
		},
		"spec": map[string]interface{}{
			"assert":   map[string]interface{}{},
			"teardown": []map[string]interface{}{},
			"setup":    map[string]interface{}{},
		},
	}

	res, err := getTestDefinition(inputData)
	assert.NotNil(t, err)
	assert.IsType(t, &TestDefinition{}, res)
}
