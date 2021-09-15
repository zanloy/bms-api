package models

import (
	"fmt"

	extensionsv1beta1 "k8s.io/api/extensions/v1beta1"
)

type DaemonSet struct {
	extensionsv1beta1.DaemonSet `json:",inline"`
	TenantInfo                  TenantInfo   `json:"tenant"`
	HealthReport                HealthReport `json:"health"`
}

func NewDaemonSet(raw *extensionsv1beta1.DaemonSet, checkHealth bool) DaemonSet {
	ds := DaemonSet{
		DaemonSet:    *raw,
		TenantInfo:   ParseTenantInfo(raw.Namespace),
		HealthReport: HealthReport{},
	}

	if checkHealth {
		ds.CheckHealth()
	}

	return ds
}

func (ds *DaemonSet) CheckHealth() {
	report := NewHealthReport()

	if ds.Status.DesiredNumberScheduled != ds.Status.NumberReady {
		msg := fmt.Sprintf("The number of desired pods [%d] does not match the number of ready pods [%d].", ds.Status.DesiredNumberScheduled, ds.Status.NumberReady)
		if ds.Status.NumberReady == 0 {
			report.AddError(msg)
		} else {
			report.AddWarning(msg)
		}
	}

	// If nobody said we're unhealthy, that must mean we are health, right?
	report.FailHealthy()

	ds.HealthReport = report
}
