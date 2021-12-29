package provisioner

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/sirupsen/logrus"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/discovery"
	"k8s.io/client-go/discovery/cached/memory"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/restmapper"
)

// Return a provisioner instance used to create, update & delete
// 		cluster-wide or namespaced resources on Kubernetes cluster
func NewProvisioner(cfg *rest.Config) *Kubernetes {

	dynclient, _ := dynamic.NewForConfig(cfg)
	client, _ := kubernetes.NewForConfig(cfg)

	return &Kubernetes{
		Config:    cfg,
		Client:    client,
		DynClient: dynclient,
	}
}

// Create or update an unstructured resource
func (k *Kubernetes) CreateOrUpdate(ctx context.Context, obj *unstructured.Unstructured) error {

	var dr dynamic.ResourceInterface

	// Init discovery client and mapper
	dc, err := discovery.NewDiscoveryClientForConfig(k.Config)
	if err != nil {
		return err
	}

	// Get GVR
	mapper := restmapper.NewDeferredDiscoveryRESTMapper(memory.NewMemCacheClient(dc))
	mapping, err := mapper.RESTMapping(
		obj.GroupVersionKind().GroupKind(),
		obj.GroupVersionKind().Version,
	)
	if err != nil {
		return err
	}

	// Get Rest Interface (Cluster or Namespaced resource)
	dr = k.DynClient.Resource(mapping.Resource)
	if mapping.Scope.Name() == meta.RESTScopeNameNamespace {
		dr = k.DynClient.Resource(mapping.Resource).Namespace(obj.GetNamespace())
	}

	// Exec rest request to API
	data, _ := json.Marshal(obj)
	_, err = dr.Patch(ctx, obj.GetName(), types.ApplyPatchType, data, metav1.PatchOptions{
		FieldManager: "go-kubetest",
	})

	return err
}

// Delete an unstructured resource
func (k *Kubernetes) Delete(ctx context.Context, obj *unstructured.Unstructured) error {

	var dr dynamic.ResourceInterface

	// Init discovery client and mapper
	dc, err := discovery.NewDiscoveryClientForConfig(k.Config)
	if err != nil {
		return err
	}

	// Get GVR
	mapper := restmapper.NewDeferredDiscoveryRESTMapper(memory.NewMemCacheClient(dc))
	mapping, err := mapper.RESTMapping(
		obj.GroupVersionKind().GroupKind(),
		obj.GroupVersionKind().Version,
	)
	if err != nil {
		return err
	}

	// Get Rest Interface (Cluster or Namespaced resource)
	dr = k.DynClient.Resource(mapping.Resource)
	if mapping.Scope.Name() == meta.RESTScopeNameNamespace {
		dr = k.DynClient.Resource(mapping.Resource).Namespace(obj.GetNamespace())
	}

	// Exec rest request to API
	deletePolicy := metav1.DeletePropagationForeground
	deleteOptions := metav1.DeleteOptions{
		PropagationPolicy: &deletePolicy,
	}
	err = dr.Delete(ctx, obj.GetName(), deleteOptions)

	return err
}

// List Resources dynamically in a Kubernetes cluster using fieldselectors
func (k *Kubernetes) ListWithSelectors(
	ctx context.Context,
	apiVersion string,
	kind string,
	namespace string,
	selectors map[string]interface{}) (*unstructured.UnstructuredList, error) {

	var labelSelector string
	var fieldSelector string
	var dr dynamic.ResourceInterface

	// Init discovery client and mapper
	dc, err := discovery.NewDiscoveryClientForConfig(k.Config)
	if err != nil {
		logrus.Debugln(err)
		return nil, err
	}

	// Get GVR
	mapper := restmapper.NewDeferredDiscoveryRESTMapper(memory.NewMemCacheClient(dc))
	mapping, err := mapper.RESTMapping(
		schema.ParseGroupKind(kind),
		apiVersion,
	)
	if err != nil {
		logrus.Debugln(err)
		return nil, err
	}

	// Init dynamic client
	dr = k.DynClient.Resource(mapping.Resource)
	if namespace != "" {
		dr = k.DynClient.Resource(mapping.Resource).Namespace(namespace)
	}

	// Composing selectors
	for k, v := range selectors {
		if strings.HasPrefix(k, "metadata.labels") {
			labelSelector = fmt.Sprintf("%v=%v,%s", strings.ReplaceAll(k, "metadata.labels.", ""), v, labelSelector)
		} else {
			fieldSelector = fmt.Sprintf("%v=%v,%s", k, v, fieldSelector)
		}
	}

	fieldSelector = strings.TrimSuffix(fieldSelector, ",")
	labelSelector = strings.TrimSuffix(labelSelector, ",")

	logrus.Debugf("Using selectors: %v && %v", labelSelector, fieldSelector)

	retrievedObjects, _ := dr.List(ctx, metav1.ListOptions{
		FieldSelector: strings.TrimSuffix(fieldSelector, ","),
		LabelSelector: strings.TrimSuffix(labelSelector, ","),
	})
	if err != nil {
		logrus.Debugln(err)
		return nil, err
	}

	logrus.Debugf("Number of objects retrieved %d", len(retrievedObjects.Items))

	return retrievedObjects, nil
}
