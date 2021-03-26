package models

import (
	corev1 "k8s.io/api/core/v1"
)

//type Node interface {
//	WithHealthReport() Node
//}

type Node struct {
	Object  corev1.Node            `json:",inline"`
	Name    string                 `json:"name"`
	Healthy corev1.ConditionStatus `json:"healthy"`
	Report  HealthReport           `json:"report,omitempty"`
}

func FromK8Node(obj corev1.Node) Node {
	return Node{
		Object: obj,
		Name:   obj.Name,
	}
}
