package cluster

import (
	"context"
	"github.com/alibaba/kt-connect/pkg/kt/util"
	"github.com/rs/zerolog/log"
	coreV1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	labelApi "k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/types"
)

// SvcMetaAndSpec ...
type SvcMetaAndSpec struct {
	Meta      *ResourceMeta
	External  bool
	Ports     map[int]int
	Selectors map[string]string
}

// GetService get service
func (k *Kubernetes) GetService(name, namespace string) (*coreV1.Service, error) {
	return k.Clientset.CoreV1().Services(namespace).Get(context.TODO(), name, metav1.GetOptions{})
}

// GetServicesBySelector get services by selector
func (k *Kubernetes) GetServicesBySelector(matchLabels map[string]string, namespace string) ([]coreV1.Service, error) {
	var matchedSvcs []coreV1.Service
	svcList, err := k.GetAllServiceInNamespace(namespace)
	if err != nil {
		return nil, err
	}
	for _, svc := range svcList.Items {
		if util.MapContains(svc.Spec.Selector, matchLabels) {
			matchedSvcs = append(matchedSvcs, svc)
		}
	}
	return matchedSvcs, nil
}

// GetServicesByLabel get services by label
func (k *Kubernetes) GetServicesByLabel(labels map[string]string, namespace string) (svcs *coreV1.ServiceList, err error) {
	return k.Clientset.CoreV1().Services(namespace).List(context.TODO(), metav1.ListOptions{
		LabelSelector: labelApi.SelectorFromSet(labels).String(),
	})
}

// GetAllServiceInNamespace get all services in specified namespace
func (k *Kubernetes) GetAllServiceInNamespace(namespace string) (*coreV1.ServiceList, error) {
	return k.Clientset.CoreV1().Services(namespace).List(context.TODO(), metav1.ListOptions{})
}

// CreateService create kubernetes service
func (k *Kubernetes) CreateService(metaAndSpec *SvcMetaAndSpec) (*coreV1.Service, error) {
	SetupHeartBeat(metaAndSpec.Meta.Name, metaAndSpec.Meta.Namespace, k.UpdateServiceHeartBeat)
	return k.Clientset.CoreV1().Services(metaAndSpec.Meta.Namespace).
		Create(context.TODO(), createService(metaAndSpec), metav1.CreateOptions{})
}

// UpdateService ...
func (k *Kubernetes) UpdateService(svc *coreV1.Service) (*coreV1.Service, error) {
	return k.Clientset.CoreV1().Services(svc.Namespace).Update(context.TODO(), svc, metav1.UpdateOptions{})
}

// RemoveService remove service
func (k *Kubernetes) RemoveService(name, namespace string) (err error) {
	client := k.Clientset.CoreV1().Services(namespace)
	return client.Delete(context.TODO(), name, metav1.DeleteOptions{})
}

func (k *Kubernetes) UpdateServiceHeartBeat(name, namespace string) {
	log.Debug().Msgf("Heartbeat service %s ticked at %s", name, formattedTime())
	if _, err := k.Clientset.CoreV1().Services(namespace).
		Patch(context.TODO(), name, types.JSONPatchType, []byte(resourceHeartbeatPatch()), metav1.PatchOptions{}); err != nil {
		log.Warn().Err(err).Msgf("Failed to update service heart beat")
	}
}

// WatchService ...
func (k *Kubernetes) WatchService(name, namespace string, fAdd, fDel, fMod func(*coreV1.Service)) {
	k.watchResource(name, namespace, string(coreV1.ResourceServices), &coreV1.Service{},
		func(obj interface{}) {
			if fAdd != nil {
				log.Debug().Msgf("Service %s added", obj.(*coreV1.Service).Name)
				fAdd(obj.(*coreV1.Service))
			}
		},
		func(obj interface{}) {
			if fDel != nil {
				log.Debug().Msgf("Service %s deleted", obj.(*coreV1.Service).Name)
				fDel(obj.(*coreV1.Service))
			}
		},
		func(obj interface{}) {
			if fMod != nil {
				log.Debug().Msgf("Service %s modified", obj.(*coreV1.Service).Name)
				fMod(obj.(*coreV1.Service))
			}
		},
	)
}
