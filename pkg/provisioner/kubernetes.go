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
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/restmapper"
)

const defaultNamespace = "default"

// Return a provisioner instance used to create, update & delete
// 		cluster-wide or namespaced resources on Kubernetes cluster
func NewProvisioner(cfg *rest.Config, client *kubernetes.Clientset, dynClient dynamic.Interface) *Kubernetes {

	return &Kubernetes{
		Config:    cfg,
		Client:    client,
		DynClient: dynClient,
	}
}

// Create or update an unstructured resource
func (k *Kubernetes) CreateOrUpdate(ctx context.Context, obj *unstructured.Unstructured) error {

	var dr dynamic.ResourceInterface

	// Init discovery client and mapper
	dc, err := discovery.NewDiscoveryClientForConfig(k.Config)
	if err != nil {
		logrus.Debugln(err)
		return err
	}

	// Get GVR
	groupResources, err := restmapper.GetAPIGroupResources(dc)
	if err != nil {
		logrus.Debugln(err)
		return err
	}

	mapper := restmapper.NewDiscoveryRESTMapper(groupResources)
	mapping, err := mapper.RESTMapping(schema.ParseGroupKind(obj.GroupVersionKind().GroupKind().String()))
	if err != nil {
		logrus.Debugln(err)
		return err
	}

	// Default to "default" namespace if not specified
	namespace := obj.GetNamespace()
	if mapping.Scope.Name() == meta.RESTScopeNameNamespace && namespace == "" {
		namespace = defaultNamespace
	}

	dr = k.DynClient.Resource(mapping.Resource)
	if mapping.Scope.Name() == meta.RESTScopeNameNamespace {
		dr = k.DynClient.Resource(mapping.Resource).Namespace(namespace)
	}

	// Check if namespace is empty and if resource is namespaced or not
	data, _ := json.Marshal(obj)
	_, err = dr.Patch(
		ctx,
		obj.GetName(),
		types.ApplyPatchType,
		data,
		metav1.PatchOptions{
			FieldManager: "go-kubetest",
		},
	)

	return err
}

// Delete an unstructured resource
func (k *Kubernetes) Delete(ctx context.Context, obj *unstructured.Unstructured) error {

	var dr dynamic.ResourceInterface

	// Init discovery client and mapper
	dc, err := discovery.NewDiscoveryClientForConfig(k.Config)
	if err != nil {
		logrus.Debugln(err)
		return err
	}

	// Get GVR
	groupResources, err := restmapper.GetAPIGroupResources(dc)
	if err != nil {
		logrus.Debugln(err)
		return err
	}

	mapper := restmapper.NewDiscoveryRESTMapper(groupResources)
	mapping, err := mapper.RESTMapping(schema.ParseGroupKind(obj.GroupVersionKind().GroupKind().String()))
	if err != nil {
		logrus.Debugln(err)
		return err
	}

	// Default to "default" namespace if not specified
	namespace := obj.GetNamespace()
	if mapping.Scope.Name() == meta.RESTScopeNameNamespace && namespace == "" {
		namespace = defaultNamespace
	}

	dr = k.DynClient.Resource(mapping.Resource)
	if mapping.Scope.Name() == meta.RESTScopeNameNamespace {
		dr = k.DynClient.Resource(mapping.Resource).Namespace(namespace)
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
func (k *Kubernetes) ListWithSelectors(ctx context.Context, objData map[string]string, selectors map[string]interface{}) (*unstructured.UnstructuredList, error) {

	var labelSelector string
	var fieldSelector string
	var dr dynamic.ResourceInterface

	apiVersion := objData["apiVersion"]
	kind := objData["kind"]
	namespace := objData["namespace"]

	// Init discovery client and mapper
	dc, err := discovery.NewDiscoveryClientForConfig(k.Config)
	if err != nil {
		logrus.Debugln(err)
		return nil, err
	}

	// Get GVR
	groupResources, err := restmapper.GetAPIGroupResources(dc)
	if err != nil {
		logrus.Debugln(err)
		return nil, err
	}

	// Use empty group name if root apiversion
	group := strings.Split(apiVersion, "/")[0]
	if group == apiVersion {
		group = ""
	}

	mapper := restmapper.NewDiscoveryRESTMapper(groupResources)
	mapping, err := mapper.RESTMapping(schema.GroupKind{Kind: kind, Group: group})
	if err != nil {
		logrus.Debugln(err)
		return nil, err
	}

	if mapping.Scope.Name() == meta.RESTScopeNameNamespace && namespace == "" {
		namespace = defaultNamespace
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
