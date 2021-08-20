package models // import "github.com/zanloy/bms-api/models"

import (
	"fmt"

	velerov1 "github.com/vmware-tanzu/velero/pkg/apis/velero/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type VeleroBackup struct {
	Name               string               `json:"name"`
	Namespace          string               `json:"namespace"`
	HealthReport       HealthReport         `json:"-"`
	Healthy            HealthyStatus        `json:"healthy"`
	Errors             []string             `json:"errors,omitempty"`
	Warnings           []string             `json:"warnings,omitempty"`
	IncludedNamespaces []string             `json:"included_namespaces"`
	Phase              velerov1.BackupPhase `json:"phase"`
	CompletedAt        *metav1.Time         `json:"completed_at"`
	raw                velerov1.Backup      `json:"-"`
}

func NewVeleroBackup(raw velerov1.Backup, checkHealth bool) VeleroBackup {
	vb := VeleroBackup{
		Name:               raw.Name,
		Namespace:          raw.Namespace,
		HealthReport:       HealthReport{},
		Healthy:            StatusUnknown,
		Errors:             []string{},
		Warnings:           []string{},
		IncludedNamespaces: raw.Spec.IncludedNamespaces,
		Phase:              raw.Status.Phase,
		CompletedAt:        raw.Status.CompletionTimestamp,
		raw:                raw,
	}
	if checkHealth {
		vb.CheckHealth()
	}
	return vb
}

func (vb *VeleroBackup) CheckHealth() {
	report := NewHealthReportFor("VeleroBackup", vb.Name, vb.Namespace)
	switch vb.Phase {
	case velerov1.BackupPhaseCompleted:
		report.Healthy = StatusHealthy
	case velerov1.BackupPhaseFailed, velerov1.BackupPhaseFailedValidation:
		report.AddError(fmt.Sprintf("Backup failed in state: %s", vb.Phase))
	case velerov1.BackupPhasePartiallyFailed:
		report.AddWarning("Backup partially failed. See logs for details.")
	default: // Includes Phases: BackupPhaseNew, BackupPhaseInProgress, BackupPhaseDeleting
		report.Healthy = StatusUnknown
	}

	// Since we explicitly set Healthy, we don't need to fallback to any state
	//report.FailHealthy()
	vb.Healthy = report.Healthy
	vb.Errors = report.Errors
	vb.Warnings = report.Warnings
	vb.HealthReport = report
}
