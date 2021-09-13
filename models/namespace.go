package models

import (
	"fmt"
	"time"

	corev1 "k8s.io/api/core/v1"
)

type NamespaceVeleroInfo struct {
	Schedules []VeleroSchedule `json:"schedules,omitempty"`
	Backups   []VeleroBackup   `json:"backups,omitempty"`
}

type Namespace struct {
	corev1.Namespace `json:",inline"`
	TenantInfo       TenantInfo   `json:"tenant"`
	HealthReport     HealthReport `json:"health"`

	DaemonSets   []DaemonSet         `json:"daemonsets"`
	Deployments  []Deployment        `json:"deployments"`
	Pods         []Pod               `json:"pods"`
	Services     []Service           `json:"services,omitempty"`
	StatefulSets []StatefulSet       `json:"statefulsets"`
	Velero       NamespaceVeleroInfo `json:"velero"`
}

func NewNamespace(raw *corev1.Namespace) Namespace {
	ns := Namespace{
		Namespace:    *raw,
		TenantInfo:   ParseTenantInfo(raw.Namespace),
		HealthReport: HealthReport{},
		DaemonSets:   make([]DaemonSet, 0),
		Deployments:  make([]Deployment, 0),
		Pods:         make([]Pod, 0),
		Services:     make([]Service, 0),
		StatefulSets: make([]StatefulSet, 0),
		Velero:       NamespaceVeleroInfo{},
	}
	return ns
}

func (ns *Namespace) CheckHealth() {
	report := NewHealthReport()

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
			if backup.Backup.Status.CompletionTimestamp != nil {
				if latest.Backup.Status.CompletionTimestamp.Before(backup.Status.CompletionTimestamp) {
					latest = &backup
				}
			}
		}
	}

	// Check the time and see if < 24h
	if latest != nil {
		if timestamp := latest.Backup.Status.CompletionTimestamp; timestamp != nil {
			if diff := time.Since(timestamp.Time).Hours(); diff <= 24.0 {
				// Check the backup status
				if latest.HealthReport.Healthy != StatusHealthy {
					report.AddWarning(fmt.Sprintf("The latest backup has the status: %s", latest.Backup.Status.Phase))
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
	*/

	// If nobody said we're unhealthy, that must mean we are health, right?
	report.FailHealthy()

	ns.HealthReport = report
}
