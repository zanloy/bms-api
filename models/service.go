package models

import (
	"fmt"

	corev1 "k8s.io/api/core/v1"
)

type Service struct {
	corev1.Service `json:",inline"`
	TenantInfo     `json:"tenant"`
	HealthReport   `json:"health"`

	Pods []Pod `json:"pods"`
}

func NewService(raw *corev1.Service, checkHealth bool) Service {
	return NewServiceWithPods(raw, make([]Pod, 0), false)
}

func NewServiceWithPods(raw *corev1.Service, pods []Pod, checkHealth bool) Service {
	service := Service{
		Service:      *raw,
		TenantInfo:   ParseTenant(raw.Namespace),
		HealthReport: HealthReport{},
		Pods:         pods,
	}

	if checkHealth {
		service.CheckHealth()
	}

	return service
}

func (s *Service) CheckHealth() {
	//   - A service is considered unhealthy if no pods are handling requests
	report := NewHealthReport()

	var total, healthy int
	for _, pod := range s.Pods {
		total++
		if pod.HealthReport.Healthy == StatusHealthy {
			healthy++
		}
	}
	if total != healthy {
		if healthy == 0 {
			report.AddError("There are no ready pods handling request")
		} else {
			report.AddWarning(fmt.Sprintf("There are %d pods matching selector but not ready.", int(total-healthy)))
		}
	}

	// If nobody said we're unhealthy, that must mean we are health, right?
	report.FailHealthy()

	s.HealthReport = report
}
