package loader

import (
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
	ObjectsList []*unstructured.Unstructured

	Setup struct {
		WaitFor []WaitFor `yaml:"waitFor"`
	} `yaml:"setup"`

	Teardown struct {
		WaitFor []WaitFor `yaml:"waitFor"`
	} `yaml:"teardown"`

	Assert []Assertion `yaml:"assert"`
}

type WaitFor struct {
	Resource string `yaml:"resource"`
	Timeout  string `yaml:"timeout"`
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
