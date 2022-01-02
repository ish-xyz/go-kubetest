package loader

import (
	"context"
	"fmt"

	"github.com/ish-xyz/go-kubetest/pkg/provisioner"
	"github.com/mitchellh/mapstructure"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

// Return a new Loader instance
func NewKubernetesLoader(prv provisioner.Provisioner) *KubernetesLoader {
	return &KubernetesLoader{
		Provisioner: prv,
	}
}

// Load testData manifests
func (ldr *KubernetesLoader) LoadManifests(testDataName string) ([]*unstructured.Unstructured, error) {
	/*
		TODO:
		get test-resource
		get data
		load yaml into struct with delimiter
	*/
	return nil, nil
}

// Load TestDefinition resources for a given namespace
func (ldr *KubernetesLoader) LoadTests(namespace string) ([]*TestDefinition, error) {
	var tests []*TestDefinition
	testDefinitions, err := ldr.Provisioner.ListWithSelectors(
		context.TODO(),
		map[string]string{
			"apiVersion": "go-kubetest.io/v1",
			"kind":       "TestDefinition",
			"namespace":  namespace,
		},
		map[string]interface{}{},
	)
	if err != nil {
		return nil, err
	}

	for _, manifest := range testDefinitions.Items {
		testDefinition := &TestDefinition{}

		for _, x := range manifest.Object["spec"].([]interface{}) {
			mapstructure.Decode(x, testDefinition)
		}
		//TODO: Load Object

		tests = append(tests, testDefinition)
	}

	fmt.Println(tests)

	return nil, nil
}
