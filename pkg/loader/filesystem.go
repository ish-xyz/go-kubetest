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

// Return a new Loader instance
func NewFileSystemLoader() *FileSystemLoader {
	return &FileSystemLoader{}
}

// Takes in the filepath to a YAML file and returns an unstructured object
// TODO: load yaml file with multiple resources
func (ldr *FileSystemLoader) LoadManifests(filepath string) ([]*unstructured.Unstructured, error) {

	var objects []*unstructured.Unstructured

	decUnstructured := yaml.NewDecodingSerializer(unstructured.UnstructuredJSONScheme)
	data, err := ioutil.ReadFile(filepath)
	if err != nil {
		return nil, err
	}

	manifests := strings.Split(string(data), YAMLDelimiter)
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

// Load multiple test files as one big array
func (ldr *FileSystemLoader) LoadTests(testsDir string) ([]*TestDefinition, error) {

	var tests []*TestDefinition

	logrus.Debugf("searching for yaml files in %s", testsDir)
	match := testsDir + "/*.yaml"
	files, err := filepath.Glob(match)
	if err != nil {
		return nil, err
	}

	logrus.Infof("files found: %v", files)
	for _, file := range files {

		test := make([]*TestDefinition, 10)
		data, err := ioutil.ReadFile(file)
		if err != nil {
			logrus.Warningf("Failed to load test file %s", file)
			logrus.Debug(err)
			continue
		}

		err = stdyaml.Unmarshal(data, &test)
		if err != nil {
			logrus.Warningf("Error while during yaml unmarshal for file %s", file)
			logrus.Debug(err)
			continue
		}

		// Load the linked manifest
		for _, singleTest := range test {

			for _, resource := range singleTest.Resources {
				objects, err := ldr.LoadManifests(fmt.Sprintf("%s/%s", testsDir, resource))
				if err != nil {
					logrus.Warningf("Error while loading resource %s in test %s", resource, singleTest.Name)
					logrus.Debug(err)
					continue
				}
				singleTest.ObjectsList = append(singleTest.ObjectsList, objects...)
			}
			tests = append(tests, singleTest)
		}
	}

	logrus.Debugln("Loaded tests: ", tests)

	return tests, nil
}
