package loader

import (
	"fmt"
	"io/ioutil"
	"path/filepath"
	"strings"

	"github.com/sirupsen/logrus"
	stdyaml "gopkg.in/yaml.v2"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/serializer/yaml"
)

const YAMLDelimiter = "---"

// Return a new Loader instance
func NewLoader() *Loader {
	return &Loader{}
}

// Takes in the filepath to a YAML file and returns an unstructured object
// TODO: load yaml file with multiple resources
func (ldr *Loader) LoadManifests(filepath string) ([]LoadedObject, error) {

	var objects []LoadedObject

	decUnstructured := yaml.NewDecodingSerializer(unstructured.UnstructuredJSONScheme)
	data, err := ioutil.ReadFile(filepath)
	if err != nil {
		return nil, err
	}

	manifests := strings.Split(string(data), YAMLDelimiter)
	for _, manifest := range manifests {
		object := LoadedObject{}
		unstructObject := &unstructured.Unstructured{}
		_, _, err = decUnstructured.Decode([]byte(manifest), nil, unstructObject)
		if err != nil {
			return nil, err
		}

		object.Object = unstructObject
		objects = append(objects, object)
	}

	return objects, err
}

// Load multiple testSuite files and related Manifests/Objects
func (ldr *Loader) LoadTestSuites(testsDir string) ([]TestDefinition, error) {

	var testSuite []TestDefinition

	match := testsDir + "/*yaml"
	logrus.Debugf("searching for yaml files in %s", match)
	files, err := filepath.Glob(match)
	if err != nil {
		return []TestDefinition{}, err
	}
	logrus.Infof("files found: %v", files)

	for _, file := range files {

		test := []TestDefinition{}
		data, err := ioutil.ReadFile(file)
		if err != nil {
			logrus.Errorf("Failed to load test file %s", file)
			logrus.Debug(err)
			continue
		}

		err = stdyaml.Unmarshal(data, &test)
		if err != nil {
			logrus.Errorf("Error while during yaml unmarshal for file %s", file)
			logrus.Debug(err)
		}

		for index, testDef := range test {
			testDef.ObjectsList, err = ldr.LoadManifests(fmt.Sprintf("%s/%s", testsDir, testDef.Manifest))
			if err != nil {
				logrus.Errorf("Error while loading manifests object in test %s", testDef.Name)
				logrus.Debug(err)
				continue
			}
			test[index] = testDef
		}

		testSuite = append(testSuite, test...)
	}

	logrus.Debugf("Loaded tests: %v", testSuite)

	return testSuite, nil
}
