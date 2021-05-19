package models

import corev1 "k8s.io/api/core/v1"

type Pod struct {
	Name        string        `json:"name"`
	Namespace   string        `json:"namespace"`
	Tenant      string        `json:"tenant,omitempty"`
	Environment string        `json:"environment,omitempty"`
	Healthy     HealthyStatus `json:"healthy"`
	Errors      []string      `json:"errors,omitempty"`
}

func FromK8Pod(pod corev1.Pod) Pod {
	var (
		report              HealthReport
		tenant, environment string
	)

	// Get tenant info
	tenant, environment = parseTenantAndEnv(pod.Namespace)

	// Get health report
	report = HealthReportForPod(pod)

	return Pod{
		Name:        pod.Name,
		Namespace:   pod.Namespace,
		Tenant:      tenant,
		Environment: environment,
		Healthy:     report.Healthy,
		Errors:      report.Errors,
	}
}
