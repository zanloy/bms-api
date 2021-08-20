package models

import (
	"fmt"
	"time"

	"github.com/zanloy/bms-api/helpers"

	corev1 "k8s.io/api/core/v1"
)

type NSVelero struct {
	Schedules []VeleroSchedule `json:"schedules,omitempty"`
	Backups   []VeleroBackup   `json:"backups,omitempty"`
}

type Namespace struct {
	Name         string        `json:"name"`
	Tenant       string        `json:"tenant"`
	Environment  string        `json:"env,omitempty"`
	HealthReport HealthReport  `json:"-"`
	Healthy      HealthyStatus `json:"healthy"`
	Errors       []string      `json:"errors,omitempty"`
	Warnings     []string      `json:"warnings,omitempty"`

	DaemonSets   []DaemonSet   `json:"daemonsets"`
	Deployments  []Deployment  `json:"deployments"`
	Pods         []Pod         `json:"pods"`
	Services     []Service     `json:"services,omitempty"`
	StatefulSets []StatefulSet `json:"statefulsets"`
	Velero       NSVelero      `json:"velero"`

	raw *corev1.Namespace
}

func NewNamespace(raw *corev1.Namespace) Namespace {
	tenant, env := helpers.ParseTenantAndEnv(raw.Name)
	ns := Namespace{
		Name:         raw.Name,
		Tenant:       tenant,
		Environment:  env,
		HealthReport: HealthReport{},
		Healthy:      StatusUnknown,
		Errors:       []string{},
		Warnings:     []string{},
		DaemonSets:   []DaemonSet{},
		Deployments:  []Deployment{},
		StatefulSets: []StatefulSet{},
		Velero:       NSVelero{},
		raw:          raw,
	}
	return ns
}

func (ns *Namespace) CheckHealth() {
	report := NewHealthReport()
	report.Kind = "namespace"
	report.Name = ns.Name
	report.Tenant, report.Environment = helpers.ParseTenantAndEnv(ns.Name)

	// DaemonSets
	for _, daemonset := range ns.DaemonSets {
		report.FoldIn(daemonset.HealthReport, fmt.Sprintf("DaemonSet[%s]", daemonset.Name))
	}
	// Deployments
	for _, deployment := range ns.Deployments {
		report.FoldIn(deployment.HealthReport, fmt.Sprintf("Deployment[%s]", deployment.Name))
	}
	// Pods
	for _, pod := range ns.Pods {
		report.FoldIn(pod.HealthReport, fmt.Sprintf("Pod[%s]", pod.Name))
	}
	// Services
	for _, service := range ns.Services {
		report.FoldIn(service.HealthReport, fmt.Sprintf("Service[%s]", service.Name))
	}
	// StatefulSets
	for _, statefulset := range ns.StatefulSets {
		report.FoldIn(statefulset.HealthReport, fmt.Sprintf("StatefulSet[%s]", statefulset.Name))
	}
	// Velero
	if len(ns.Velero.Schedules) < 1 {
		report.AddWarning("There are no Velero Schedules for this namespace")
	}
	// Find the most recent backup
	var latest *VeleroBackup
	for _, backup := range ns.Velero.Backups {
		if latest == nil {
			latest = &backup
		} else {
			if backup.CompletedAt != nil {
				if latest.CompletedAt.Before(backup.CompletedAt) {
					latest = &backup
				}
			}
		}
	}

	// Check the time and see if < 24h
	if latest != nil {
		if timestamp := latest.CompletedAt; timestamp != nil {
			if diff := time.Since(timestamp.Time).Hours(); diff <= 24.0 {
				// Check the backup status
				if latest.Healthy != StatusHealthy {
					report.AddWarning(fmt.Sprintf("The latest backup has the status: %s", latest.Phase))
				}
			} else {
				report.AddWarning(fmt.Sprintf("The latest backup is %d days old", int(diff/24)))
			}
		}
	} else {
		report.AddWarning("No recent Velero backups found.")
	}
	/*
		// Check Services
		if services, err := Services(namespace.Name).List(labels.Everything()); err == nil {
			unhealthyServices := make([]string, 0, len(services))
			for _, service := range services {
				report := ReportForService(*service)
				if report.Healthy != models.StatusHealthy {
					unhealthyServices = append(unhealthyServices, service.Name)
				}
			}
			if len(unhealthyServices) > 0 {
				nsreport.Healthy = models.StatusUnhealthy
				nsreport.Errors = append(nsreport.Errors, fmt.Sprintf("Services with unhealthy status: [%s].", strings.Join(unhealthyServices, ",")))
			}
		} else {
			nsreport.Errors = append(nsreport.Errors, "Failed to fetch Services from Kubernetes.")
		}

		// Get the latest backup
		if backups, err := VeleroBackupsForNamespace(namespace.Name); err == nil {
			var latest *velerov1.Backup

			for _, backup := range backups {
				if latest == nil {
					latest = &backup
				} else {
					if timestamp := backup.Status.CompletionTimestamp; timestamp != nil {
						if latest.Status.CompletionTimestamp.Before(backup.Status.CompletionTimestamp) {
							latest = &backup
						}
					}
				}
			}

			// Check the time and see if < 24h
			if latest != nil {
				if timestamp := latest.Status.CompletionTimestamp; timestamp != nil {
					if time.Since(timestamp.Time).Hours() < 24.0 {
						// Check the backup status
						if latest.Status.Phase != velerov1.BackupPhaseCompleted {
							nsreport.AddWarning(fmt.Sprintf("The latest backup (%s) has the status: %s", latest.Name, latest.Status.Phase))
						}
					}
				}
			} else {
				nsreport.AddWarning("No recent Velero backups found.")
			}
		} else {
			nsreport.AddWarning(err.Error())
		}
	*/

	// If nobody said we're unhealthy, that must mean we are health, right?
	report.FailHealthy()

	// Save everything to the object
	ns.Healthy = report.Healthy
	ns.Errors = report.Errors
	ns.Warnings = report.Warnings
	ns.HealthReport = report
}
