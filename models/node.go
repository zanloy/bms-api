package models

import (
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metricsv1beta1 "k8s.io/metrics/pkg/apis/metrics/v1beta1"
)

type ResourceQuantities struct {
	Allocatable resource.Quantity `json:"allocatable"`
	Allocated   resource.Quantity `json:"allocated"`
	Utilized    resource.Quantity `json:"utilized"`
}

type Node struct {
	Name           string             `json:"name"`
	Healthy        HealthyStatus      `json:"healthy"`
	Errors         []string           `json:"errors,omitempty"`
	Conditions     []string           `json:"conditions,omitempty"`
	KernelVersion  string             `json:"kernel_version,omitempty"`
	KubeletVersion string             `json:"kubelet_version,omitempty"`
	CPU            ResourceQuantities `json:"cpu"`
	Memory         ResourceQuantities `json:"memory"`
}

func FromK8Node(node corev1.Node) Node {
	var (
		report     HealthReport
		conditions = make([]string, 0, len(node.Status.Conditions))
	)

	// Get health report
	report = HealthReportForNode(node)

	// Get node conditions
	for _, condition := range node.Status.Conditions {
		if condition.Status == corev1.ConditionTrue {
			conditions = append(conditions, string(condition.Type))
		}
	}

	return Node{
		Name:           node.Name,
		Healthy:        report.Healthy,
		Errors:         report.Errors,
		Conditions:     conditions,
		KernelVersion:  node.Status.NodeInfo.KernelVersion,
		KubeletVersion: node.Status.NodeInfo.KubeletVersion,
		CPU: ResourceQuantities{
			Allocatable: node.Status.Allocatable["cpu"],
		},
		Memory: ResourceQuantities{
			Allocatable: node.Status.Allocatable["memory"],
		},
	}
}

func (n *Node) AddMetrics(metrics metricsv1beta1.NodeMetrics) {
	if usage, ok := metrics.Usage["cpu"]; ok {
		n.CPU.Utilized = usage
	}
	if usage, ok := metrics.Usage["memory"]; ok {
		n.Memory.Utilized = usage
	}
}
