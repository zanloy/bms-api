package models

import (
	"time"
)

type ReportSummary struct {
	Timestamp int64          `json:"timestamp"`
	Date      time.Time      `json:"date"`
	Errors    []string       `json:"errors"`
	Counts    map[string]int `json:"counts"`
}

type Report struct {
	Date                  time.Time       `json:"date"`
	Errors                []string        `json:"errors"`
	Nodes                 []Node          `json:"nodes"`
	UnhealthyDaemonSets   []DaemonSet     `json:"unhealthy_daemonsets"`
	UnhealthyDeployments  []Deployment    `json:"unhealthy_deployments"`
	UnhealthyPods         []Pod           `json:"unhealthy_pods"`
	UnhealthyStatefulSets []StatefulSet   `json:"unhealthy_statefulsets"`
	Restarts              []ReportRestart `json:"restarts"`
	URLs                  []URLCheck      `json:"urlchecks"`
}

func NewReport() Report {
	return Report{
		Date:                  time.Now(),
		Errors:                make([]string, 0),
		Nodes:                 make([]Node, 0),
		UnhealthyDaemonSets:   make([]DaemonSet, 0),
		UnhealthyDeployments:  make([]Deployment, 0),
		UnhealthyPods:         make([]Pod, 0),
		UnhealthyStatefulSets: make([]StatefulSet, 0),
		Restarts:              make([]ReportRestart, 0),
		URLs:                  make([]URLCheck, 0),
	}
}

func (r *Report) Summary() ReportSummary {
	return ReportSummary{
		Timestamp: r.Date.Unix(),
		Date:      r.Date,
		Errors:    r.Errors,
		Counts: map[string]int{
			"nodes":                  len(r.Nodes),
			"unhealthy_daemonsets":   len(r.UnhealthyDaemonSets),
			"unhealthy_deployments":  len(r.UnhealthyDeployments),
			"unhealthy_pods":         len(r.UnhealthyPods),
			"unhealthy_statefulsets": len(r.UnhealthyStatefulSets),
		},
	}
}

type ReportRestart struct {
	Namespace    string    `json:"namespace"`
	Name         string    `json:"name"`
	RestartCount uint      `json:"restart_count"`
	LastRestart  time.Time `json:"last_restart,omitempty"`
}
