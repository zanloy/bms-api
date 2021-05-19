package models

import (
	extensionsv1beta1 "k8s.io/api/extensions/v1beta1"
)

type DaemonSet struct {
	Name        string        `json:"name"`
	Namespace   string        `json:"namespace"`
	Tenant      string        `json:"tenant,omitempty"`
	Environment string        `json:"environment,omitempty"`
	Healthy     HealthyStatus `json:"healthy"`
	Errors      []string      `json:"errors,omitempty"`
}

func FromK8DaemonSet(daemonset extensionsv1beta1.DaemonSet) DaemonSet {
	var (
		report              HealthReport
		tenant, environment string
	)

	// Get tenant info
	tenant, environment = parseTenantAndEnv(daemonset.Namespace)

	// Get health report
	report = HealthReportForDaemonSet(daemonset)

	return DaemonSet{
		Name:        daemonset.Name,
		Namespace:   daemonset.Namespace,
		Tenant:      tenant,
		Environment: environment,
		Healthy:     report.Healthy,
		Errors:      report.Errors,
	}
}
