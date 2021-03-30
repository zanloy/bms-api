package models

import (
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
)

type ResourceQuantities struct {
	Allocatable resource.Quantity `json:"allocatable"`
	Allocated   resource.Quantity `json:"allocated"`
	Utilized    resource.Quantity `json:"utilized"`
}

type Node struct {
	Object         corev1.Node        `json:",inline"`
	Name           string             `json:"name"`
	Healthy        HealthyStatus      `json:"healthy"`
	Conditions     []string           `json:"conditions,omitempty"`
	KernelVersion  string             `json:"kernel_version,omitempty"`
	KubeletVersion string             `json:"kubelet_version,omitempty"`
	CPU            ResourceQuantities `json:"cpu"`
	Memory         ResourceQuantities `json:"memory"`
	CPUAllocation  string             `json:"cpu_allocation"`
	RAMAllocation  string             `json:"ram_allocation"`
	CPUUtilization string             `json:"cpu_utilization"`
	RAMUtilization string             `json:"ram_utilization"`
	Report         HealthReport       `json:"report,omitempty"`
}

func FromK8Node(obj corev1.Node) Node {
	return Node{
		Object: obj,
		Name:   obj.Name,
	}
}
