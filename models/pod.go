package models

import (
	"github.com/zanloy/bms-api/helpers"

	corev1 "k8s.io/api/core/v1"
)

type Pod struct {
	Name         string        `json:"name"`
	Namespace    string        `json:"namespace"`
	Tenant       string        `json:"tenant,omitempty"`
	Environment  string        `json:"environment,omitempty"`
	HealthReport HealthReport  `json:"-"`
	Healthy      HealthyStatus `json:"healthy"`
	Errors       []string      `json:"errors,omitempty"`
	Warnings     []string      `json:"warnings,omitempty"`
	raw          *corev1.Pod
}

func NewPod(raw *corev1.Pod, checkHealth bool) Pod {
	tenant, env := helpers.ParseTenantAndEnv(raw.Name)
	pod := Pod{
		Name:         raw.Name,
		Namespace:    raw.Namespace,
		Tenant:       tenant,
		Environment:  env,
		HealthReport: HealthReport{},
		Healthy:      StatusUnknown,
		Errors:       []string{},
		Warnings:     []string{},
		raw:          raw,
	}
	if checkHealth {
		pod.CheckHealth()
	}
	return pod
}

func (p *Pod) CheckHealth() {
	report := NewHealthReport()
	report.Kind = "Pod"
	report.Namespace = p.Namespace
	report.Name = p.Name
	report.Tenant = p.Tenant
	report.Environment = p.Environment

	for name, value := range p.raw.Labels {
		if name == "jenkins" && value == "slave" {
			// This pod is part of a jenkins job
			report.Healthy = StatusIgnored
		}
	}

	if report.Healthy != StatusIgnored && p.raw.Status.Phase != corev1.PodSucceeded {
		for _, condition := range p.raw.Status.Conditions {
			if condition.Type == corev1.PodReady && condition.Status == corev1.ConditionFalse {
				report.AddError(condition.Message)
			}
		}
	}

	report.FailHealthy()
	p.Healthy = report.Healthy
	p.Errors = report.Errors
	p.Warnings = report.Warnings
	p.HealthReport = report
}
