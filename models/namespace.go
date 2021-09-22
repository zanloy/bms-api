package models

import (
	"fmt"
	"sort"
	"time"

	velerov1 "github.com/vmware-tanzu/velero/pkg/apis/velero/v1"
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

	DaemonSets   []DaemonSet         `json:"daemonsets,omitempty"`
	Deployments  []Deployment        `json:"deployments,omitempty"`
	Events       []*corev1.Event     `json:"events,omitempty"`
	Pods         []Pod               `json:"pods,omitempty"`
	Services     []Service           `json:"services,omitempty"`
	StatefulSets []StatefulSet       `json:"statefulsets,omitempty"`
	Velero       NamespaceVeleroInfo `json:"velero,omitempty"`
}

func NewNamespace(raw *corev1.Namespace) Namespace {
	ns := Namespace{
		Namespace:    *raw,
		TenantInfo:   ParseTenantInfo(raw.Name),
		HealthReport: HealthReport{},
		DaemonSets:   make([]DaemonSet, 0),
		Deployments:  make([]Deployment, 0),
		Pods:         make([]Pod, 0),
		Services:     make([]Service, 0),
		StatefulSets: make([]StatefulSet, 0),
		Velero:       NamespaceVeleroInfo{},
	}

	ns.Kind = "Namespace"

	return ns
}

func NewNamespaceWithEvents(raw *corev1.Namespace, events []*corev1.Event) Namespace {
	ns := NewNamespace(raw)
	ns.Events = events
	return ns
}

func (ns *Namespace) CheckHealth(addons []string) {
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
		if len(pod.ObjectMeta.OwnerReferences) == 0 {
			report.FoldIn(pod.HealthReport, fmt.Sprintf("Pod[%s]", pod.Name))
		}
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
	if idx := sort.SearchStrings(addons, "velero"); idx < len(addons) && addons[idx] == "velero" {
		// Make sure we are scheduled for backups
		if len(ns.Velero.Schedules) < 1 {
			report.AddWarning("There are no Velero Schedules for this namespace")
		}

		// Find the most recent backup
		var latest *velerov1.Backup
		for _, backup := range ns.Velero.Backups {
			if latest == nil {
				latest = backup.DeepCopy()
			} else {
				if backup.Backup.Status.CompletionTimestamp != nil {
					if latest.Status.CompletionTimestamp.Before(backup.Status.CompletionTimestamp) {
						latest = backup.DeepCopy()
					}
				}
			}
		}

		// Check the time and see if < 24h
		if latest != nil {
			latestBackup := NewVeleroBackup(latest, true)
			if timestamp := latestBackup.Status.CompletionTimestamp; timestamp != nil {
				if diff := time.Since(timestamp.Time).Hours(); diff <= 24.0 {
					// Check the backup status
					if latestBackup.HealthReport.Healthy != StatusHealthy {
						report.AddWarning(fmt.Sprintf("The latest backup has the status: %s", latestBackup.Status.Phase))
					}
				} else {
					report.AddWarning(fmt.Sprintf("The latest backup is %d days old", int(diff/24)))
				}
			}
		} else {
			report.AddWarning("No recent Velero backups found.")
		}
	}

	// If nobody said we're unhealthy, that must mean we are health, right?
	report.FailHealthy()

	ns.HealthReport = report
}
