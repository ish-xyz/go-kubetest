package loader

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/ish-xyz/go-kubetest/pkg/provisioner"
	"github.com/sirupsen/logrus"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/serializer/yaml"
)

// Return a new Loader instance
func NewKubernetesLoader(prv provisioner.Provisioner) *KubernetesLoader {
	return &KubernetesLoader{
		Provisioner: prv,
	}
}

// Load testData manifests
func (ldr *KubernetesLoader) LoadManifests(resourcePath string) ([]*unstructured.Unstructured, error) {

	var objects []*unstructured.Unstructured

	namespace := strings.Split(":", resourcePath)[0]
	name := strings.Split(":", resourcePath)[1]

	testResources, err := ldr.Provisioner.ListWithSelectors(
		context.TODO(),
		map[string]string{
			"apiVersion": "go-kubetest.io/v1",
			"kind":       "TestResource",
			"namespace":  namespace,
		},
		map[string]interface{}{
			"metadata.name": name,
		},
	)
	if err != nil {
		return nil, err
	} else if len(testResources.Items) < 1 {
		return nil, fmt.Errorf("no resource with name %s", name)
	}

	spec := testResources.Items[0].Object["spec"].(map[string]interface{})
	manifests := strings.Split(spec["data"].(string), YAMLDelimiter)
	decUnstructured := yaml.NewDecodingSerializer(unstructured.UnstructuredJSONScheme)

	for _, manifest := range manifests {
		unstructObject := &unstructured.Unstructured{}
		_, _, err = decUnstructured.Decode([]byte(manifest), nil, unstructObject)
		if err != nil {
			logrus.Debugln(err)
			return nil, err
		}

		objects = append(objects, unstructObject)
	}

	return objects, err
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
		map[string]interface{}{
			"metadata.name": "my-test-definition",
		},
	)
	fmt.Println(testDefinitions)
	if err != nil {
		return nil, err
	} else if len(testDefinitions.Items) < 1 {
		return nil, fmt.Errorf("can't retrieve any tests from the Kubernetes API")
	}

	for _, testDef := range testDefinitions.Items[0].Object["spec"].([]interface{}) {

		testSpec, err := getManifestSpec(testDef)
		if err != nil {
			logrus.Warningf("Can't convert manifest.spec into TestDefinition")
		}

		for _, resource := range testSpec.Resources {
			objects, err := ldr.LoadManifests(fmt.Sprintf("%s:%s", namespace, resource))
			if err != nil {
				logrus.Warningf("Error while loading manifests object in test %s", testSpec.Name)
				logrus.Debugln(err)
				continue
			}
			testSpec.ObjectsList = append(testSpec.ObjectsList, objects...)
		}

		tests = append(tests, testSpec)
	}

	return tests, nil
}

func getManifestSpec(testDef interface{}) (*TestDefinition, error) {

	testDefStruct := &TestDefinition{}
	uobj := unstructured.Unstructured{Object: testDef.(map[string]interface{})}
	testDefJson, err := uobj.MarshalJSON()
	if err != nil {
		return nil, err
	}

	err = json.Unmarshal(testDefJson, testDefStruct)
	if err != nil {
		return nil, err
	}

	return testDefStruct, nil
}
