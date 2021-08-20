package models

import (
	"fmt"

	"github.com/zanloy/bms-api/helpers"

	corev1 "k8s.io/api/core/v1"
)

type Service struct {
	Name         string          `json:"name"`
	Namespace    string          `json:"namespace"`
	Tenant       string          `json:"tenant,omitempty"`
	Environment  string          `json:"environment,omitempty"`
	HealthReport HealthReport    `json:"-"`
	Healthy      HealthyStatus   `json:"healthy"`
	Errors       []string        `json:"errors,omitempty"`
	Warnings     []string        `json:"warnings,omitempty"`
	raw          *corev1.Service `json:"-"`
	Pods         []Pod           `json:"pods,omitempty"`
}

func NewService(raw *corev1.Service, checkHealth bool) Service {
	return NewServiceWithPods(raw, make([]Pod, 0), false)
}

func NewServiceWithPods(raw *corev1.Service, pods []Pod, checkHealth bool) Service {
	tenant, env := helpers.ParseTenantAndEnv(raw.Namespace)
	service := Service{
		Name:         raw.Name,
		Namespace:    raw.Namespace,
		Tenant:       tenant,
		Environment:  env,
		HealthReport: HealthReport{},
		Healthy:      StatusUnknown,
		Errors:       []string{},
		Warnings:     []string{},
		raw:          raw,
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
	report.Kind = "Service"
	report.Namespace = s.Namespace
	report.Name = s.Name
	report.Tenant = s.Tenant
	report.Environment = s.Environment

	var total, healthy int
	for _, pod := range s.Pods {
		total++
		if pod.Healthy == StatusHealthy {
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

	report.FailHealthy()
	s.Healthy = report.Healthy
	s.Errors = report.Errors
	s.Warnings = report.Warnings
	s.HealthReport = report
}
