package loader

import (
	"context"
	"fmt"
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
	).Return(&unstructured.UnstructuredList{}, fmt.Errorf("random error"))

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

func TestGetTestDefinition(t *testing.T) {

	inputData := map[string]interface{}{
		"name":      "input-data",
		"resources": []string{},
		"assert":    []map[string]interface{}{},
		"teardown":  map[string]interface{}{},
		"setup":     map[string]interface{}{},
	}

	res, err := getTestDefinition(inputData)
	assert.Nil(t, err)
	assert.IsType(t, &TestDefinition{}, res)
	assert.Equal(t, res.Name, "input-data")
}

func TestGetTestDefinitionErrors(t *testing.T) {

	wrongData := map[string]interface{}{
		"name":      123,
		"resources": "wrong-data",
	}

	res, err := getTestDefinition(wrongData)
	assert.NotNil(t, err)
	assert.Nil(t, res)
}

func TestLoadTestsEmpty(t *testing.T) {

	// Prepare mock and data
	namespace := "default"
	selectors := map[string]interface{}{
		"metadata.labels.type": "soft",
	}

	prvMock := new(provisioner.ProvisionerMock)
	prvMock.On(
		"ListWithSelectors",
		context.TODO(),
		map[string]string{
			"apiVersion": "go-kubetest.io/v1",
			"kind":       "TestDefinition",
			"namespace":  namespace,
		},
		selectors,
	).Return(&unstructured.UnstructuredList{}, nil)

	// Execute
	ldr := NewKubernetesLoader(prvMock)
	res, err := ldr.LoadTests(namespace, selectors)

	// Assertions
	assert.NotNil(t, err)
	assert.Nil(t, res)

	prvMock.AssertNumberOfCalls(t, "ListWithSelectors", 1)
}

func TestLoadTestsErrors(t *testing.T) {

	// Prepare mock and data
	namespace := "default"
	selectors := map[string]interface{}{
		"metadata.labels.type": "soft",
	}

	prvMock := new(provisioner.ProvisionerMock)
	prvMock.On(
		"ListWithSelectors",
		context.TODO(),
		map[string]string{
			"apiVersion": "go-kubetest.io/v1",
			"kind":       "TestDefinition",
			"namespace":  namespace,
		},
		selectors,
	).Return(&unstructured.UnstructuredList{}, fmt.Errorf("failed to fetch data."))

	// Execute
	ldr := NewKubernetesLoader(prvMock)
	res, err := ldr.LoadTests(namespace, selectors)

	// Assertions
	assert.NotNil(t, err)
	assert.Nil(t, res)

	prvMock.AssertNumberOfCalls(t, "ListWithSelectors", 1)
}

func TestLoadTests(t *testing.T) {

	innerTestName := "inner-test-name"
	returnData := &unstructured.UnstructuredList{
		Items: []unstructured.Unstructured{
			{
				Object: map[string]interface{}{
					"apiVersion": "go-kubetest.io/v1",
					"kind":       "TestDefinition",
					"metadata": map[string]interface{}{
						"name": "test-input-data",
					},
					"spec": []interface{}{
						map[string]interface{}{"name": innerTestName},
					},
				},
			},
		},
	}
	// Prepare mock and data
	namespace := "default"
	selectors := map[string]interface{}{
		"metadata.labels.type": "soft",
	}

	prvMock := new(provisioner.ProvisionerMock)
	prvMock.On(
		"ListWithSelectors",
		context.TODO(),
		map[string]string{
			"apiVersion": "go-kubetest.io/v1",
			"kind":       "TestDefinition",
			"namespace":  namespace,
		},
		selectors,
	).Return(returnData, nil)

	// Execute
	ldr := NewKubernetesLoader(prvMock)
	res, err := ldr.LoadTests(namespace, selectors)

	// Assertions
	assert.Nil(t, err)
	assert.IsType(t, []*TestDefinition{}, res)
	assert.NotNil(t, res)
	assert.Equal(t, res[0].Name, innerTestName)

	prvMock.AssertNumberOfCalls(t, "ListWithSelectors", 1)
}
