package models

import (
	v1 "k8s.io/api/core/v1"
	extensionsv1beta1 "k8s.io/api/extensions/v1beta1"
)

type DaemonSet struct {
	object    extensionsv1beta1.DaemonSet
	Name      string             `json:"name"`
	Namespace string             `json:"namespace"`
	Healthy   v1.ConditionStatus `json:"healthy"`
	Report    *HealthReport      `json:"report,omitempty"`
}

func FromK8DaemonSet(input extensionsv1beta1.DaemonSet) DaemonSet {
	return DaemonSet{
		object:    input,
		Name:      input.Name,
		Namespace: input.Namespace,
		Healthy:   v1.ConditionUnknown,
	}
}
