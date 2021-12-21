package loader

import (
	"time"

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
	ApiVersion     string                 `yaml:"apiVersion"`
	Kind           string                 `yaml:"kind"`
	Namespace      string                 `yaml:"namespace"`
	Selectors      map[string]interface{} `yaml:"selectors"`
	Count          int                    `yaml:"count"`
	ExpectedErrors int                    `yaml:"expectedErrors"`
}

type TestDefinition struct {
	Name           string        `yaml:"name"`
	Manifest       string        `yaml:"manifest"`
	ExpectedErrors int           `yaml:"expectedErrors"`
	Timeout        time.Duration `yaml:"timeout"`
	ObjectsList    []LoadedObject
	Assert         []Assertion `yaml:"assert"`
	Status         string      // TODO DeleteError, CreateError, Fail, Success
}
