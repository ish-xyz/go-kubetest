package loader

import (
	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

type Loader struct{}

type Config struct {
	TestsDir string
	Interval int
}

type LoadedObject struct {
	Object  *unstructured.Unstructured
	Mapping *meta.RESTMapping
}

type TestDefinition struct {
	ID          int    `yaml:"id"`
	Manifest    string `yaml:"manifest"`
	ObjectsList []LoadedObject
	Assert      []struct {
		Selectors map[string]interface{} `yaml:"selectors"`
		Count     int                    `yaml:"count"`
	}
	Status string // DeleteError, CreateError, Fail, Success
}

type TestSuite struct {
	Name  string           `yaml:"name"`
	Tests []TestDefinition `yaml:"tests"`
}

type TestSuitesList struct {
	TestSuites []TestSuite
}
