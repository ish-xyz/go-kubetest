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

type TestDefinition struct {
	Name        string `yaml:"name"`
	Manifest    string `yaml:"manifest"`
	ObjectsList []*LoadedObject

	Setup struct {
		WaitFor []struct {
			Resource string `yaml:"resource"`
			Timeout  string `yaml:"timeout"`
		} `yaml:"waitFor"`
	} `yaml:"setup"`

	Teardown struct {
		WaitFor []struct {
			Resource string `yaml:"resource"`
			Timeout  string `yaml:"timeout"`
		} `yaml:"waitFor"`
	} `yaml:"teardown"`

	Assert []Assertion `yaml:"assert"`
}

type LoadedObject struct {
	Object  *unstructured.Unstructured
	Mapping *meta.RESTMapping
}

type Assertion struct {
	Type       string                 `yaml:"type"`
	ApiVersion string                 `yaml:"apiVersion"`
	Kind       string                 `yaml:"kind"`
	Namespace  string                 `yaml:"namespace"`
	Selectors  map[string]interface{} `yaml:"selectors"`
	Count      int                    `yaml:"count"`
	Errors     []string               `yaml:"expectedErrors"`
}
