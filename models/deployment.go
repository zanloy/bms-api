package models

import (
	"fmt"

	corev1 "k8s.io/api/core/v1"
	extensionsv1beta1 "k8s.io/api/extensions/v1beta1"
)

type Deployment struct {
	extensionsv1beta1.Deployment `json:",inline"`
	TenantInfo                   `json:"tenant"`
	HealthReport                 `json:"health"`
}

func NewDeployment(raw *extensionsv1beta1.Deployment, checkHealth bool) Deployment {
	deployment := Deployment{
		Deployment:   *raw,
		TenantInfo:   ParseTenant(raw.Namespace),
		HealthReport: HealthReport{},
	}

	if checkHealth {
		deployment.CheckHealth()
	}

	return deployment
}

func (d *Deployment) CheckHealth() {
	report := NewHealthReport()

	for _, condition := range d.Status.Conditions {
		if condition.Type == extensionsv1beta1.DeploymentAvailable && condition.Status == corev1.ConditionFalse {
			// K8 states Availability as binary and I want to be able to have Warn/Yellow
			// so I check the replicas to see if we have any vs 0.
			msg := fmt.Sprintf("The number of desired pods [%v] does not match the number of ready pods [%v].", *d.Spec.Replicas, d.Status.ReadyReplicas)
			if d.Status.ReadyReplicas == 0 {
				report.AddError(msg)
			} else {
				report.AddWarning(msg)
			}
		}
	}

	// If nobody said we're unhealthy, that must mean we are health, right?
	report.FailHealthy()

	d.HealthReport = report
}
