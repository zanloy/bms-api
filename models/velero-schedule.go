package models // import "github.com/zanloy/bms-api/models"

import (
	velerov1 "github.com/vmware-tanzu/velero/pkg/apis/velero/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type VeleroSchedule struct {
	Name               string                 `json:"name"`
	Namespace          string                 `json:"namespace"`
	HealthReport       HealthReport           `json:"-"`
	Healthy            HealthyStatus          `json:"healthy"`
	Errors             []string               `json:"errors,omitempty"`
	Warnings           []string               `json:"warnings,omitempty"`
	IncludedNamespaces []string               `json:"included_namespaces"`
	Phase              velerov1.SchedulePhase `json:"phase"`
	LastBackupAt       *metav1.Time           `json:"last_backup_at"`
	raw                velerov1.Schedule      `json:"-"`
}

func NewVeleroSchedule(raw velerov1.Schedule, checkHealth bool) VeleroSchedule {
	vs := VeleroSchedule{
		Name:               raw.Name,
		Namespace:          raw.Namespace,
		HealthReport:       HealthReport{},
		Healthy:            StatusUnknown,
		Errors:             []string{},
		Warnings:           []string{},
		IncludedNamespaces: raw.Spec.Template.IncludedNamespaces,
		Phase:              raw.Status.Phase,
		LastBackupAt:       raw.Status.LastBackup,
		raw:                raw,
	}
	if checkHealth {
		vs.CheckHealth()
	}
	return vs
}

func (vs *VeleroSchedule) CheckHealth() {
	report := NewHealthReportFor("velero-schedule", vs.Name, vs.Namespace)
	if vs.Phase == velerov1.SchedulePhaseFailedValidation {
		report.AddError("failed validation phase")
	}

	// Since we explicitly set Healthy, we don't need to fallback to any state
	report.FailHealthy()
	vs.Healthy = report.Healthy
	vs.Errors = report.Errors
	vs.Warnings = report.Warnings
	vs.HealthReport = report
}
