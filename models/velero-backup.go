package models // import "github.com/zanloy/bms-api/models"

import (
	"fmt"

	velerov1 "github.com/vmware-tanzu/velero/pkg/apis/velero/v1"
)

type VeleroBackup struct {
	velerov1.Backup `json:",inline"`
	TenantInfo      TenantInfo   `json:"tenant"`
	HealthReport    HealthReport `json:"health"`
}

func NewVeleroBackup(raw *velerov1.Backup, checkHealth bool) VeleroBackup {
	vb := VeleroBackup{
		Backup:       *raw,
		TenantInfo:   ParseTenantInfo(raw.Namespace),
		HealthReport: HealthReport{},
	}

	if checkHealth {
		vb.CheckHealth()
	}

	return vb
}

func (vb *VeleroBackup) CheckHealth() {
	report := NewHealthReport()
	switch vb.Status.Phase {
	case velerov1.BackupPhaseCompleted:
		report.Healthy = StatusHealthy
	case velerov1.BackupPhaseFailed, velerov1.BackupPhaseFailedValidation:
		report.AddError(fmt.Sprintf("Backup failed in state: %s", vb.Status.Phase))
	case velerov1.BackupPhasePartiallyFailed:
		report.AddWarning("Backup partially failed. See logs for details.")
	default: // Includes Phases: BackupPhaseNew, BackupPhaseInProgress, BackupPhaseDeleting
		report.Healthy = StatusUnknown
	}

	// Since we explicitly set Healthy, we don't need to fallback to any state
	//report.FailHealthy()

	vb.HealthReport = report
}
