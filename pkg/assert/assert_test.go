package assert

import (
	"context"
	"fmt"
	"testing"

	"github.com/ish-xyz/go-kubetest/pkg/loader"
	"github.com/ish-xyz/go-kubetest/pkg/provisioner"
	"github.com/stretchr/testify/assert"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

// TODO

func TestGetMaxRetries(t *testing.T) {
	res := getMaxRetries("20s", 2)

	assert.Equal(t, 10, res)
}

func TestUnpackResourceClusterWide(t *testing.T) {

	version, kind, namespace, err := unpackResource("v1:Namespace")

	assert.Equal(t, "v1", version)
	assert.Equal(t, "Namespace", kind)
	assert.Equal(t, "", namespace)
	assert.Nil(t, err)
}

func TestUnpackResourceNamespaced(t *testing.T) {

	version, kind, namespace, err := unpackResource("v1:Pod:default")

	assert.Equal(t, "v1", version)
	assert.Equal(t, "Pod", kind)
	assert.Equal(t, "default", namespace)
	assert.Nil(t, err)
}

func TestExpectedResources(t *testing.T) {

	retObjects := &unstructured.UnstructuredList{
		Items: []unstructured.Unstructured{
			{
				Object: map[string]interface{}{
					"spec": map[string]interface{}{},
				},
			},
		},
	}

	prvMock := new(provisioner.ProvisionerMock)
	prvMock.On(
		"ListWithSelectors",
		context.TODO(),
		map[string]string{
			"apiVersion": "v1",
			"kind":       "Pod",
			"namespace":  "default",
		},
		map[string]interface{}{
			"metadata.name": "resource",
		},
	).Return(retObjects, nil)

	asrt := loader.Assertion{
		Resource: "v1:Pod:default",
		Selectors: map[string]interface{}{
			"metadata.name": "resource",
		},
		Timeout: "6s",
		Count:   1,
	}

	res := expectedResources(prvMock, asrt)

	assert.True(t, res)
	prvMock.AssertNumberOfCalls(t, "ListWithSelectors", 1)
}

func TestExpectedResourcesFailed(t *testing.T) {

	retObjects := &unstructured.UnstructuredList{
		Items: []unstructured.Unstructured{
			{
				Object: map[string]interface{}{
					"spec": map[string]interface{}{},
				},
			},
		},
	}

	prvMock := new(provisioner.ProvisionerMock)
	prvMock.On(
		"ListWithSelectors",
		context.TODO(),
		map[string]string{
			"apiVersion": "v1",
			"kind":       "Pod",
			"namespace":  "default",
		},
		map[string]interface{}{
			"metadata.name": "resource",
		},
	).Return(retObjects, nil)

	asrt := loader.Assertion{
		Resource: "v1:Pod:default",
		Selectors: map[string]interface{}{
			"metadata.name": "resource",
		},
		Timeout: "6s",
		Count:   100000,
	}

	res := expectedResources(prvMock, asrt)

	assert.False(t, res)
	prvMock.AssertNumberOfCalls(t, "ListWithSelectors", 3)
}

func TestExpectedResourcesFailedWithErrors(t *testing.T) {

	retObjects := &unstructured.UnstructuredList{
		Items: []unstructured.Unstructured{
			{
				Object: map[string]interface{}{
					"spec": map[string]interface{}{},
				},
			},
		},
	}

	prvMock := new(provisioner.ProvisionerMock)
	prvMock.On(
		"ListWithSelectors",
		context.TODO(),
		map[string]string{
			"apiVersion": "v1",
			"kind":       "Pod",
			"namespace":  "default",
		},
		map[string]interface{}{
			"metadata.name": "resource",
		},
	).Return(retObjects, fmt.Errorf("some random error"))

	asrt := loader.Assertion{
		Resource: "v1:Pod:default",
		Selectors: map[string]interface{}{
			"metadata.name": "resource",
		},
		Timeout: "6s",
		Count:   1,
	}

	res := expectedResources(prvMock, asrt)

	assert.False(t, res)
	prvMock.AssertNumberOfCalls(t, "ListWithSelectors", 3)
}

func TestExpectedResourcesWrongPath(t *testing.T) {

	prvMock := new(provisioner.ProvisionerMock)
	asrt := loader.Assertion{
		Resource: "wrong-path",
		Selectors: map[string]interface{}{
			"metadata.name": "resource",
		},
		Timeout: "6s",
		Count:   100000,
	}

	res := expectedResources(prvMock, asrt)

	assert.False(t, res)
	prvMock.AssertNumberOfCalls(t, "ListWithSelectors", 0)
}

func TestExpectedErrorsRegex(t *testing.T) {

	expErrors := []string{".*SecurityContext.*"}
	actErrors := []string{"something SecurityContext something"}

	res := expectedErrors(expErrors, actErrors)

	assert.True(t, res)

}

func TestExpectedErrors(t *testing.T) {

	expErrors := []string{"something SecurityContext something"}
	actErrors := []string{"something SecurityContext something"}

	res := expectedErrors(expErrors, actErrors)

	assert.True(t, res)

}

func TestExpectedErrorsFailed(t *testing.T) {

	expErrors := []string{}
	actErrors := []string{"some random error"}

	res := expectedErrors(expErrors, actErrors)

	assert.False(t, res)

}
