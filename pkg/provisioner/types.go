package provisioner

import (
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

type Provisioner struct {
	Client    *kubernetes.Clientset
	DynClient dynamic.Interface
	Config    *rest.Config
}
