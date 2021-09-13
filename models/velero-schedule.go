package models // import "github.com/zanloy/bms-api/models"

import (
	velerov1 "github.com/vmware-tanzu/velero/pkg/apis/velero/v1"
)

type VeleroSchedule struct {
	velerov1.Schedule `json:",inline"`
	TenantInfo        `json:"tenant"`
	HealthReport      `json:"health"`
}

func NewVeleroSchedule(raw velerov1.Schedule, checkHealth bool) VeleroSchedule {
	vs := VeleroSchedule{
		Schedule:     raw,
		TenantInfo:   ParseTenantInfo(raw.Namespace),
		HealthReport: HealthReport{},
	}

	if checkHealth {
		vs.CheckHealth()
	}

	return vs
}

func (vs *VeleroSchedule) CheckHealth() {
	report := NewHealthReport()
	if vs.Status.Phase == velerov1.SchedulePhaseFailedValidation {
		report.AddError("failed validation phase")
	}

	// Since we explicitly set Healthy, we don't need to fallback to any state
	report.FailHealthy()
	vs.Healthy = report.Healthy
	vs.Errors = report.Errors
	vs.Warnings = report.Warnings
	vs.HealthReport = report
}
