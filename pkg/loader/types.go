package loader

import (
	"github.com/ish-xyz/go-kubetest/pkg/provisioner"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

// Interfaces
type Loader interface {
	LoadManifests(string) ([]*unstructured.Unstructured, error)
	LoadTests(string) ([]*TestDefinition, error)
}

// Data
type Config struct {
	TestsDir string
	Interval int
}

type TestDefinition struct {
	Name        string   `yaml:"name" json:"name"`
	Resources   []string `yaml:"resources" json:"resources"`
	ObjectsList []*unstructured.Unstructured

	Setup struct {
		WaitFor []WaitFor `yaml:"waitFor" json:"waitFor"`
	} `yaml:"setup" json:"setup"`

	Teardown struct {
		WaitFor []WaitFor `yaml:"waitFor" json:"waitFor"`
	} `yaml:"teardown" json:"teardown"`

	Assert []Assertion `yaml:"assert" json:"assert"`
}

type WaitFor struct {
	Resource string `yaml:"resource" json:"resource"`
	Timeout  string `yaml:"timeout" json:"timeout"`
}

type Assertion struct {
	Type       string                 `yaml:"type" json:"type"`
	ApiVersion string                 `yaml:"apiVersion" json:"apiVersion"`
	Kind       string                 `yaml:"kind" json:"kind"`
	Namespace  string                 `yaml:"namespace" json:"namespace"`
	Selectors  map[string]interface{} `yaml:"selectors" json:"selectors"`
	Count      int                    `yaml:"count" json:"count"`
	Errors     []string               `yaml:"expectedErrors" json:"expectedErrors"`
}

// Loaders
type FileSystemLoader struct{}

type KubernetesLoader struct {
	Provisioner provisioner.Provisioner
}
