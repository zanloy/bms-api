package models // import "github.com/zanloy/bms-api/models"

import (
	velerov1 "github.com/vmware-tanzu/velero/pkg/apis/velero/v1"
)

type VeleroSchedule struct {
	velerov1.Schedule `json:",inline"`
	TenantInfo        TenantInfo   `json:"tenant"`
	HealthReport      HealthReport `json:"health"`
}

func NewVeleroSchedule(raw *velerov1.Schedule, checkHealth bool) VeleroSchedule {
	vs := VeleroSchedule{
		Schedule:     *raw,
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

	report.FailHealthy()

	vs.HealthReport = report
}
