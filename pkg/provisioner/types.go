package provisioner

import (
	"context"

	"github.com/ish-xyz/go-kubetest/pkg/loader"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

type Kubernetes struct {
	Client    *kubernetes.Clientset
	DynClient dynamic.Interface
	Config    *rest.Config
}

type Provisioner interface {
	CreateOrUpdate(ctx context.Context, object *loader.LoadedObject) error
	Delete(ctx context.Context, object *loader.LoadedObject) error
	ListWithSelectors(ctx context.Context, apiVersion, kind, namespace string, selectors map[string]interface{}) (*unstructured.UnstructuredList, error)
}
