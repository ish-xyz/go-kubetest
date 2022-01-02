package provisioner

import (
	"context"

	"github.com/stretchr/testify/mock"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

// Interfaces
type Provisioner interface {
	CreateOrUpdate(context.Context, *unstructured.Unstructured) error
	Delete(context.Context, *unstructured.Unstructured) error
	ListWithSelectors(context.Context, map[string]string, map[string]interface{}) (*unstructured.UnstructuredList, error)
}

// Provisioners
type Kubernetes struct {
	Client    *kubernetes.Clientset
	DynClient dynamic.Interface
	Config    *rest.Config
}

type ProvisionerMock struct {
	mock.Mock
}
