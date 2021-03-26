package models

import (
	v1 "k8s.io/api/core/v1"
)

type Pod struct {
	object    v1.Pod
	Name      string        `json:"name"`
	Namespace string        `json:"namespace"`
	Healthy   HealthyStatus `json:"healthy"`
	Report    *HealthReport `json:"report,omitempty"`
}

func FromK8Pod(input v1.Pod) Pod {
	return Pod{
		object:    input,
		Name:      input.Name,
		Namespace: input.Namespace,
		Healthy:   StatusUnknown,
	}
}
