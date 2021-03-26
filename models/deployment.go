package models

import (
	corev1 "k8s.io/api/core/v1"
	extensionsv1beta1 "k8s.io/api/extensions/v1beta1"
)

type Deployment struct {
	object    extensionsv1beta1.Deployment
	Name      string                 `json:"name"`
	Namespace string                 `json:"namespace"`
	Healthy   corev1.ConditionStatus `json:"healthy"`
	Report    *HealthReport          `json:"report,omitempty"`
}

func FromK8Deployment(input extensionsv1beta1.Deployment) Deployment {
	return Deployment{
		object:    input,
		Name:      input.Name,
		Namespace: input.Namespace,
		Healthy:   corev1.ConditionUnknown,
	}
}
