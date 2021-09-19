package models

import (
	corev1 "k8s.io/api/core/v1"
)

type Pod struct {
	corev1.Pod   `json:",inline"`
	TenantInfo   TenantInfo   `json:"tenant"`
	HealthReport HealthReport `json:"health"`
	Restarts     int32        `json:"restarts"`
}

func NewPod(raw *corev1.Pod, checkHealth bool) Pod {
	pod := Pod{
		Pod:          *raw,
		TenantInfo:   ParseTenantInfo(raw.Namespace),
		HealthReport: HealthReport{},
	}
	for _, cs := range pod.Status.ContainerStatuses {
		pod.Restarts += cs.RestartCount
	}

	pod.Kind = "Pod"

	if checkHealth {
		pod.CheckHealth()
	}

	return pod
}

func (p *Pod) CheckHealth() {
	report := NewHealthReport()

	for name, value := range p.Labels {
		if name == "jenkins" && value == "slave" {
			// This pod is part of a jenkins job
			report.Healthy = StatusIgnored
		}
	}

	if report.Healthy != StatusIgnored && p.Status.Phase != corev1.PodSucceeded {
		for _, condition := range p.Status.Conditions {
			if condition.Type == corev1.PodReady && condition.Status == corev1.ConditionFalse {
				report.AddError(condition.Message)
			}
		}
	}

	// If nobody said we're unhealthy, that must mean we are health, right?
	report.FailHealthy()

	p.HealthReport = report
}
