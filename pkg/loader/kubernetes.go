package loader

import (
	"context"
	"encoding/json"

	"github.com/ish-xyz/go-kubetest/pkg/provisioner"
	"github.com/sirupsen/logrus"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

// Return a new Loader instance
func NewKubernetesLoader(prv provisioner.Provisioner) *KubernetesLoader {
	return &KubernetesLoader{
		Provisioner: prv,
	}
}

// Load testData manifests
func (ldr *KubernetesLoader) LoadManifests(testResource string) ([]*unstructured.Unstructured, error) {
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

		for _, testDef := range manifest.Object["spec"].([]interface{}) {
			testDefStruct := &TestDefinition{}
			uobj := unstructured.Unstructured{Object: testDef.(map[string]interface{})}
			testDefJson, err := uobj.MarshalJSON()
			if err != nil {
				logrus.Warningf("Can't Marshal JSON for object")
				logrus.Debugln(err)
				continue
			}

			err = json.Unmarshal(testDefJson, testDefStruct)
			if err != nil {
				logrus.Warningf("Can't unmarshal test %s", string(testDefJson))
				logrus.Debugln(err)
				continue
			}

			testDefStruct.ObjectsList, err = ldr.LoadManifests("")
			if err != nil {
				logrus.Warningf("Error while loading manifests object in test %s", testDefStruct.Name)
				logrus.Debugln(err)
				continue
			}
			tests = append(tests, testDefStruct)
		}
	}

	return tests, nil
}
