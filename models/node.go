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
	HealthReport   HealthReport       `json:"-"`
	Healthy        HealthyStatus      `json:"healthy"`
	Errors         []string           `json:"errors,omitempty"`
	Warnings       []string           `json:"warnings,omitempty"`
	Conditions     []string           `json:"conditions,omitempty"`
	KernelVersion  string             `json:"kernel_version,omitempty"`
	KubeletVersion string             `json:"kubelet_version,omitempty"`
	CPU            ResourceQuantities `json:"cpu"`
	Memory         ResourceQuantities `json:"memory"`
	raw            *corev1.Node       `json:"-"`
}

func NewNode(raw *corev1.Node, checkHealth bool) Node {
	conditions := make([]string, 0, len(raw.Status.Conditions))
	for _, condition := range raw.Status.Conditions {
		if condition.Status == corev1.ConditionTrue {
			conditions = append(conditions, string(condition.Type))
		}
	}
	node := Node{
		Name:           raw.Name,
		HealthReport:   HealthReport{},
		Healthy:        StatusUnknown,
		Errors:         []string{},
		Warnings:       []string{},
		Conditions:     conditions,
		KernelVersion:  raw.Status.NodeInfo.KernelVersion,
		KubeletVersion: raw.Status.NodeInfo.KubeletVersion,
		CPU:            ResourceQuantities{Allocatable: raw.Status.Allocatable["cpu"]},
		Memory:         ResourceQuantities{Allocatable: raw.Status.Allocatable["memory"]},
		raw:            raw,
	}
	if checkHealth {
		node.CheckHealth()
	}
	return node
}

func (n *Node) AddMetrics(metrics metricsv1beta1.NodeMetrics) {
	if usage, ok := metrics.Usage["cpu"]; ok {
		n.CPU.Utilized = usage
	}
	if usage, ok := metrics.Usage["memory"]; ok {
		n.Memory.Utilized = usage
	}
}

func (n *Node) CheckHealth() {
	report := NewHealthReportFor("node", n.Name, "")

	for _, condition := range n.raw.Status.Conditions {
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

	report.FailHealthy()
	n.Healthy = report.Healthy
	n.Errors = report.Errors
	n.Warnings = report.Warnings
	n.HealthReport = report
}
