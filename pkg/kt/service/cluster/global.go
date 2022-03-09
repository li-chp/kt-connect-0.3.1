package cluster

import (
	"context"
	"github.com/alibaba/kt-connect/pkg/kt/util"
	coreV1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/tools/cache"
	"strings"
	"time"
)

// ResourceMeta ...
type ResourceMeta struct {
	Name        string
	Namespace   string
	Labels      map[string]string
	Annotations map[string]string
}

// SSHkeyMeta ...
type SSHkeyMeta struct {
	SshConfigMapName string
	PrivateKeyPath   string
}

// GetAllNamespaces get all namespaces
func (k *Kubernetes) GetAllNamespaces() (*coreV1.NamespaceList, error) {
	return k.Clientset.CoreV1().Namespaces().List(context.TODO(), metav1.ListOptions{})
}

// GetKtResources fetch all kt pods and deployments
func (k *Kubernetes) GetKtResources(namespace string) ([]coreV1.Pod, []coreV1.ConfigMap, []coreV1.Service, error) {
	pods, err := Ins().GetPodsByLabel(map[string]string{util.ControlBy: util.KubernetesToolkit}, namespace)
	if err != nil {
		return nil, nil, nil, err
	}
	configmaps, err := Ins().GetConfigMapsByLabel(map[string]string{util.ControlBy: util.KubernetesToolkit}, namespace)
	if err != nil {
		return nil, nil, nil, err
	}
	services, err := Ins().GetServicesByLabel(map[string]string{util.ControlBy: util.KubernetesToolkit}, namespace)
	if err != nil {
		return nil, nil, nil, err
	}
	return pods.Items, configmaps.Items, services.Items, nil
}

// watchResource watch for change
// name: empty for any name
// namespace: empty for all namespace
// fAdd, fDel, fMod: nil for ignore
func (k *Kubernetes) watchResource(name, namespace, resourceType string, objType runtime.Object, fAdd, fDel, fMod func(interface{})) {
	selector := fields.Nothing()
	if name != "" {
		selector = fields.OneTermEqualSelector("metadata.name", name)
	}
	watchlist := cache.NewListWatchFromClient(
		k.Clientset.CoreV1().RESTClient(),
		resourceType,
		namespace,
		selector,
	)
	_, controller := cache.NewInformer(
		watchlist,
		objType,
		0,
		cache.ResourceEventHandlerFuncs{
			AddFunc:    func(obj interface{}) { fAdd(obj) },
			DeleteFunc: func(obj interface{}) { fDel(obj) },
			UpdateFunc: func(oldObj, newObj interface{}) { fMod(newObj) },
		},
	)

	stop := make(chan struct{})
	defer close(stop)
	go controller.Run(stop)
	for {
		time.Sleep(1000 * time.Second)
	}
}

func isSingleIp(ipRange string) bool {
	return !strings.Contains(ipRange, "/") || strings.Split(ipRange, "/")[1] == "32"
}
