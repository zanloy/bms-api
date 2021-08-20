package models

import (
	"fmt"

	"github.com/zanloy/bms-api/helpers"

	corev1 "k8s.io/api/core/v1"
	extensionsv1beta1 "k8s.io/api/extensions/v1beta1"
)

type Deployment struct {
	Name         string        `json:"name"`
	Namespace    string        `json:"namespace"`
	Tenant       string        `json:"tenant,omitempty"`
	Environment  string        `json:"environment,omitempty"`
	HealthReport HealthReport  `json:"-"`
	Healthy      HealthyStatus `json:"healthy"`
	Errors       []string      `json:"errors,omitempty"`
	Warnings     []string      `json:"warnings,omitempty"`
	raw          *extensionsv1beta1.Deployment
}

func NewDeployment(raw *extensionsv1beta1.Deployment, checkHealth bool) Deployment {
	tenant, env := helpers.ParseTenantAndEnv(raw.Namespace)
	d := Deployment{
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
		d.CheckHealth()
	}
	return d
}

func (d *Deployment) CheckHealth() {
	report := NewHealthReport()
	report.Kind = "Deployment"
	report.Namespace = d.Namespace
	report.Name = d.Name
	report.Tenant = d.Tenant
	report.Environment = d.Environment

	for _, condition := range d.raw.Status.Conditions {
		if condition.Type == extensionsv1beta1.DeploymentAvailable && condition.Status == corev1.ConditionFalse {
			// K8 states Availability as binary and I want to be able to have Warn/Yellow
			// so I check the replicas to see if we have any vs 0.
			msg := fmt.Sprintf("The number of desired pods [%v] does not match the number of ready pods [%v].", *d.raw.Spec.Replicas, d.raw.Status.ReadyReplicas)
			if d.raw.Status.ReadyReplicas == 0 {
				report.AddError(msg)
			} else {
				report.AddWarning(msg)
			}
		}
	}

	report.FailHealthy()

	d.Healthy = report.Healthy
	d.Errors = report.Errors
	d.Warnings = report.Warnings
	d.HealthReport = report
}
