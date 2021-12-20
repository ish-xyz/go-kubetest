package provisioner

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/ish-xyz/go-kubetest/pkg/loader"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/discovery"
	"k8s.io/client-go/discovery/cached/memory"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/restmapper"
)

type Provisioner struct {
	Client    *kubernetes.Clientset
	DynClient dynamic.Interface
	Config    *rest.Config
}

// Return a provisioner instance used to create, update & delete
// 		cluster-wide or namespaced resources on Kubernetes cluster
func NewProvisioner(cfg *rest.Config) *Provisioner {

	dynclient, _ := dynamic.NewForConfig(cfg)
	client, _ := kubernetes.NewForConfig(cfg)

	return &Provisioner{
		Config:    cfg,
		Client:    client,
		DynClient: dynclient,
	}
}

// Create or update an unstructured resource
func (p *Provisioner) CreateOrUpdate(ctx context.Context, object loader.LoadedObject) (loader.LoadedObject, error) {

	var dr dynamic.ResourceInterface
	obj := object.Object
	object, err := createMapper(object, p.Config)
	if err != nil {
		return loader.LoadedObject{}, err
	}

	// Get Rest Interface (Cluster or Namespaced resource)
	mapping := object.Mapping
	dr = p.DynClient.Resource(mapping.Resource)
	if mapping.Scope.Name() == meta.RESTScopeNameNamespace {
		dr = p.DynClient.Resource(mapping.Resource).Namespace(obj.GetNamespace())
	}

	// Exec rest request to API
	data, _ := json.Marshal(obj)
	_, err = dr.Patch(ctx, obj.GetName(), types.ApplyPatchType, data, metav1.PatchOptions{
		FieldManager: "go-kubetest",
	})

	return object, err
}

// Delete an unstructured resource
func (p *Provisioner) Delete(ctx context.Context, object loader.LoadedObject) (loader.LoadedObject, error) {

	var dr dynamic.ResourceInterface
	obj := object.Object
	object, err := createMapper(object, p.Config)
	if err != nil {
		return loader.LoadedObject{}, err
	}

	// Get Rest Interface (Cluster or Namespaced resource)
	mapping := object.Mapping
	dr = p.DynClient.Resource(mapping.Resource)
	if mapping.Scope.Name() == meta.RESTScopeNameNamespace {
		dr = p.DynClient.Resource(mapping.Resource).Namespace(obj.GetNamespace())
	}

	// Exec rest request to API
	deletePolicy := metav1.DeletePropagationForeground
	deleteOptions := metav1.DeleteOptions{
		PropagationPolicy: &deletePolicy,
	}
	err = dr.Delete(ctx, obj.GetName(), deleteOptions)

	return object, err
}

// List Resources dynamically in a Kubernetes cluster using fieldselectors
func (p *Provisioner) ListWithSelectors(ctx context.Context, object loader.LoadedObject, fieldselectors string) (*unstructured.UnstructuredList, error) {

	var dr dynamic.ResourceInterface
	obj := object.Object
	object, err := createMapper(object, p.Config)
	if err != nil {
		return nil, err
	}

	// Get Rest Interface (Cluster or Namespaced resource)
	mapping := object.Mapping
	dr = p.DynClient.Resource(mapping.Resource)
	if mapping.Scope.Name() == meta.RESTScopeNameNamespace {
		dr = p.DynClient.Resource(mapping.Resource).Namespace(obj.GetNamespace())
	}

	//TODO Create field selector dinamically
	/*selectors := ""

	for k, v := range fieldselectors {
		selectors = fmt.Sprintf("%s,%v=%v", selectors, k, v)
	}*/
	o, _ := dr.List(ctx, metav1.ListOptions{FieldSelector: fieldselectors})
	fmt.Println(len(o.Items))

	return nil, nil
}

// Create Mapper and return a copy of object with the Mapper set
func createMapper(object loader.LoadedObject, cfg *rest.Config) (loader.LoadedObject, error) {

	obj := object.Object
	mapping := object.Mapping

	// Skip if mapping is passed
	if mapping == nil {
		// Init discovery client and mapper
		dc, err := discovery.NewDiscoveryClientForConfig(cfg)
		if err != nil {
			return loader.LoadedObject{}, err
		}
		mapper := restmapper.NewDeferredDiscoveryRESTMapper(memory.NewMemCacheClient(dc))

		// Get GVR
		mapping, err = mapper.RESTMapping(
			obj.GroupVersionKind().GroupKind(),
			obj.GroupVersionKind().Version,
		)
		if err != nil {
			return loader.LoadedObject{}, err
		}
	}

	// Add Mapping to the copy of the object
	object.Mapping = mapping

	return object, nil
}
