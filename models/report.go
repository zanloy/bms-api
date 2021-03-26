package models

import (
	"time"

	"k8s.io/apimachinery/pkg/api/resource"
)

type Report struct {
	Date           time.Time       `json:"date"`
	Errors         []string        `json:"errors"`
	Nodes          []ReportNode    `json:"nodes"`
	Restarts       []ReportRestart `json:"restarts"`
	NonhealthyPods []ReportPod     `json:"nonhealthy_pods"`
	URLs           []URLCheck      `json:"urlchecks"`
}

func NewReport() Report {
	return Report{
		Date:           time.Now(),
		Errors:         make([]string, 0),
		Nodes:          make([]ReportNode, 0),
		Restarts:       make([]ReportRestart, 0),
		NonhealthyPods: make([]ReportPod, 0),
	}
}

type ReportNode struct {
	Name           string             `json:"name"`
	Conditions     []string           `json:"conditions,omitempty"`
	KernelVersion  string             `json:"kernel_version,omitempty"`
	KubeletVersion string             `json:"kubelet_version,omitempty"`
	CPU            ReportResourceData `json:"cpu"`
	Memory         ReportResourceData `json:"memory"`
	CPUAllocation  string             `json:"cpu_allocation"`
	RAMAllocation  string             `json:"ram_allocation"`
	CPUUtilization string             `json:"cpu_utilization"`
	RAMUtilization string             `json:"ram_utilization"`
}

type ReportPod struct {
	Namespace string `json:"namespace"`
	Name      string `json:"name"`
	Ready     string `json:"ready"`
	Restarts  uint   `json:"restarts"`
}

type ReportResourceData struct {
	Allocatable resource.Quantity `json:"allocatable"`
	Allocated   resource.Quantity `json:"allocated"`
	Utilized    resource.Quantity `json:"utilized"`
}

type ReportRestart struct {
	Namespace    string    `json:"namespace"`
	Name         string    `json:"name"`
	RestartCount uint      `json:"restart_count"`
	LastRestart  time.Time `json:"last_restart,omitempty"`
}
