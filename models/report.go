package models

import (
	"time"
)

type Report struct {
	Date           time.Time       `json:"date"`
	Errors         []string        `json:"errors"`
	Nodes          []Node          `json:"nodes"`
	Restarts       []ReportRestart `json:"restarts"`
	NonhealthyPods []ReportPod     `json:"nonhealthy_pods"`
	URLs           []URLCheck      `json:"urlchecks"`
}

func NewReport() Report {
	return Report{
		Date:           time.Now(),
		Errors:         make([]string, 0),
		Nodes:          make([]Node, 0),
		Restarts:       make([]ReportRestart, 0),
		NonhealthyPods: make([]ReportPod, 0),
	}
}

type ReportPod struct {
	Namespace string `json:"namespace"`
	Name      string `json:"name"`
	Ready     string `json:"ready"`
	Restarts  uint   `json:"restarts"`
}

type ReportRestart struct {
	Namespace    string    `json:"namespace"`
	Name         string    `json:"name"`
	RestartCount uint      `json:"restart_count"`
	LastRestart  time.Time `json:"last_restart,omitempty"`
}
