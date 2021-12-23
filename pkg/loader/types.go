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

type Assertion struct {
	ApiVersion        string                 `yaml:"apiVersion"`
	Kind              string                 `yaml:"kind"`
	Namespace         string                 `yaml:"namespace"`
	Selectors         map[string]interface{} `yaml:"selectors"`
	ExpectedResources int                    `yaml:"expectedResources"`
}

type TestDefinition struct {
	Name     string `yaml:"name"`
	Manifest string `yaml:"manifest"`
	Setup    struct {
		ExpectedErrors []string `yaml:"expectedErrors`
		WaitFor        []struct {
			Resource string `yaml:"resource"`
			Timeout  string `yaml:"timeout"`
		}
	}
	ObjectsList []LoadedObject
	Assert      []Assertion `yaml:"assert"`
}
