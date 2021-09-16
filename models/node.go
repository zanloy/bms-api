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
	Percentage  uint8             `json:"percentage"`
}

type NodeResources struct {
	CPU    ResourceQuantities `json:"cpu"`
	Memory ResourceQuantities `json:"memory"`
}

type Node struct {
	corev1.Node  `json:",inline"`
	HealthReport HealthReport  `json:"health"`
	Conditions   []string      `json:"conditions"`
	Resources    NodeResources `json:"resources"`
}

func NewNode(raw *corev1.Node, checkHealth bool) Node {
	conditions := make([]string, 0, len(raw.Status.Conditions))
	for _, condition := range raw.Status.Conditions {
		if condition.Status == corev1.ConditionTrue {
			conditions = append(conditions, string(condition.Type))
		}
	}

	node := Node{
		Node:         *raw,
		HealthReport: HealthReport{},
		Conditions:   conditions,
	}

	node.Kind = "Node"

	node.Resources.CPU.Allocatable = node.Status.Allocatable["cpu"]
	node.Resources.Memory.Allocatable = node.Status.Allocatable["memory"]

	if checkHealth {
		node.CheckHealth()
	}

	return node
}

func (n *Node) AddMetrics(metrics *metricsv1beta1.NodeMetrics) {
	if usage, ok := metrics.Usage["cpu"]; ok {
		n.Resources.CPU.Utilized = usage
		n.Resources.CPU.Percentage = uint8(n.Resources.CPU.Utilized.AsApproximateFloat64() / n.Resources.CPU.Allocatable.AsApproximateFloat64() * 100)
	}
	if usage, ok := metrics.Usage["memory"]; ok {
		n.Resources.Memory.Utilized = usage
		n.Resources.Memory.Percentage = uint8(n.Resources.Memory.Utilized.AsApproximateFloat64() / n.Resources.Memory.Allocatable.AsApproximateFloat64() * 100)
	}
}

func (n *Node) CheckHealth() {
	report := NewHealthReport()

	for _, condition := range n.Status.Conditions {
		switch condition.Type {
		case corev1.NodeReady:
			if condition.Status != corev1.ConditionTrue {
				report.AddError(condition.Message)
			}
		default:
			if condition.Status != corev1.ConditionFalse {
				report.AddError(condition.Message)
			}
		}
	}

	// If nobody said we're unhealthy, that must mean we are health, right?
	report.FailHealthy()

	n.HealthReport = report
}
