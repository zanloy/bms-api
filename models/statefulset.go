package models

import (
	appsv1 "k8s.io/api/apps/v1"
)

type StatefulSet struct {
	Name        string        `json:"name"`
	Namespace   string        `json:"namespace"`
	Tenant      string        `json:"tenant,omitempty"`
	Environment string        `json:"environment,omitempty"`
	Healthy     HealthyStatus `json:"healthy"`
	Errors      []string      `json:"errors,omitempty"`
}

func FromK8StatefulSet(statefulset appsv1.StatefulSet) StatefulSet {
	var (
		report              HealthReport
		tenant, environment string
	)

	// Get tenant info
	tenant, environment = parseTenantAndEnv(statefulset.Namespace)

	// Get health report
	report = HealthReportForStatefulSet(statefulset)

	return StatefulSet{
		Name:        statefulset.Name,
		Namespace:   statefulset.Namespace,
		Tenant:      tenant,
		Environment: environment,
		Healthy:     report.Healthy,
		Errors:      report.Errors,
	}
}
