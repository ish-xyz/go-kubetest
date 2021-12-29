package provisioner

import (
	"context"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

func (_m *ProvisionerMock) CreateOrUpdate(ctx context.Context, object *unstructured.Unstructured) error {
	args := _m.Called(ctx, object)
	return args.Error(0)
}

func (_m *ProvisionerMock) Delete(ctx context.Context, object *unstructured.Unstructured) error {
	args := _m.Called(ctx, object)
	return args.Error(0)
}

func (_m *ProvisionerMock) ListWithSelectors(
	ctx context.Context,
	apiVersion, kind, namespace string,
	selectors map[string]interface{}) (*unstructured.UnstructuredList, error) {

	args := _m.Called(ctx, apiVersion, kind, namespace, selectors)
	return args.Get(0).(*unstructured.UnstructuredList), args.Error(1)
}
