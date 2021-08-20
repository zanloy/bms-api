package models

import (
	"fmt"

	"github.com/zanloy/bms-api/helpers"

	appsv1 "k8s.io/api/apps/v1"
)

type StatefulSet struct {
	Name         string        `json:"name"`
	Namespace    string        `json:"namespace"`
	Tenant       string        `json:"tenant,omitempty"`
	Environment  string        `json:"environment,omitempty"`
	HealthReport HealthReport  `json:"-"`
	Healthy      HealthyStatus `json:"healthy"`
	Errors       []string      `json:"errors,omitempty"`
	Warnings     []string      `json:"warnings,omitempty"`
	raw          *appsv1.StatefulSet
}

func NewStatefulSet(raw *appsv1.StatefulSet, checkHealth bool) StatefulSet {
	tenant, env := helpers.ParseTenantAndEnv(raw.Name)
	ss := StatefulSet{
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
		ss.CheckHealth()
	}
	return ss
}

func (ss *StatefulSet) CheckHealth() {
	report := NewHealthReport()
	report.Kind = "StatefulSet"
	report.Name = ss.Name
	report.Namespace = ss.Namespace
	report.Tenant = ss.Tenant
	report.Environment = ss.Environment

	if int32(*ss.raw.Spec.Replicas) != int32(ss.raw.Status.ReadyReplicas) {
		msg := fmt.Sprintf("The number of desired replicas [%d] does not match the number of ready replicas [%d].", *ss.raw.Spec.Replicas, ss.raw.Status.ReadyReplicas)
		if int32(ss.raw.Status.ReadyReplicas) == 0 {
			report.AddError(msg)
		} else {
			report.AddWarning(msg)
		}
	}

	report.FailHealthy()
	ss.Healthy = report.Healthy
	ss.Errors = report.Errors
	ss.Warnings = report.Warnings
	ss.HealthReport = report
}
