package models

import (
	"fmt"

	"github.com/zanloy/bms-api/helpers"
	extensionsv1beta1 "k8s.io/api/extensions/v1beta1"
)

type DaemonSet struct {
	Name         string        `json:"name"`
	Namespace    string        `json:"namespace"`
	Tenant       string        `json:"tenant,omitempty"`
	Environment  string        `json:"environment,omitempty"`
	HealthReport HealthReport  `json:"-"`
	Healthy      HealthyStatus `json:"healthy"`
	Errors       []string      `json:"errors,omitempty"`
	Warnings     []string      `json:"warnings,omitempty"`

	raw *extensionsv1beta1.DaemonSet
}

func NewDaemonSet(raw *extensionsv1beta1.DaemonSet, checkHealth bool) DaemonSet {
	tenant, env := helpers.ParseTenantAndEnv(raw.Namespace)
	ds := DaemonSet{
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
		ds.CheckHealth()
	}
	return ds
}

func (ds *DaemonSet) CheckHealth() {
	report := NewHealthReport()
	report.Kind = "DaemonSet"
	report.Namespace = ds.Namespace
	report.Name = ds.Name
	report.Tenant, report.Environment = helpers.ParseTenantAndEnv(ds.Namespace)

	if ds.raw.Status.DesiredNumberScheduled != ds.raw.Status.NumberReady {
		msg := fmt.Sprintf("The number of desired pods [%d] does not match the number of ready pods [%d].", ds.raw.Status.DesiredNumberScheduled, ds.raw.Status.NumberReady)
		if ds.raw.Status.NumberReady == 0 {
			report.AddError(msg)
		} else {
			report.AddWarning(msg)
		}
	}

	report.FailHealthy()
	ds.Healthy = report.Healthy
	ds.Errors = report.Errors
	ds.Warnings = report.Warnings
}
