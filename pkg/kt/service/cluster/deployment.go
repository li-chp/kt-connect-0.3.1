package cluster

import (
	"context"
	"github.com/rs/zerolog/log"
	appV1 "k8s.io/api/apps/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	labelApi "k8s.io/apimachinery/pkg/labels"
)

// GetDeployment ...
func (k *Kubernetes) GetDeployment(name string, namespace string) (*appV1.Deployment, error) {
	return k.Clientset.AppsV1().Deployments(namespace).Get(context.TODO(), name, metav1.GetOptions{})
}

// GetDeploymentsByLabel get deployments by label
func (k *Kubernetes) GetDeploymentsByLabel(labels map[string]string, namespace string) (pods *appV1.DeploymentList, err error) {
	return k.Clientset.AppsV1().Deployments(namespace).List(context.TODO(), metav1.ListOptions{
		LabelSelector: labelApi.SelectorFromSet(labels).String(),
	})
}

// GetAllDeploymentInNamespace get all deployment in specified namespace
func (k *Kubernetes) GetAllDeploymentInNamespace(namespace string) (*appV1.DeploymentList, error) {
	return k.Clientset.AppsV1().Deployments(namespace).List(context.TODO(), metav1.ListOptions{})
}

// UpdateDeployment ...
func (k *Kubernetes) UpdateDeployment(deployment *appV1.Deployment) (*appV1.Deployment, error) {
	return k.Clientset.AppsV1().Deployments(deployment.Namespace).Update(context.TODO(), deployment, metav1.UpdateOptions{})
}

// ScaleTo scale deployment to
func (k *Kubernetes) ScaleTo(name, namespace string, replicas *int32) (err error) {
	deployment, err := k.GetDeployment(name, namespace)
	if err != nil {
		return
	}

	if deployment.Spec.Replicas == replicas {
		log.Warn().Msgf("Deployment %s already having %d replicas, not need to scale", name, replicas)
		return nil
	}

	log.Info().Msgf("Scaling deployment %s from %d to %d", deployment.Name, deployment.Spec.Replicas, replicas)
	deployment.Spec.Replicas = replicas

	if _, err = k.UpdateDeployment(deployment); err != nil {
		log.Error().Err(err).Msgf("Failed to scale deployment %s", deployment.Name)
		return
	}
	log.Info().Msgf("Deployment %s successfully scaled to %d replicas", name, replicas)
	return
}
