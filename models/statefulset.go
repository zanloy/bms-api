package models

import (
	"fmt"

	appsv1 "k8s.io/api/apps/v1"
)

type StatefulSet struct {
	appsv1.StatefulSet `json:",inline"`
	TenantInfo         TenantInfo   `json:"tenant"`
	HealthReport       HealthReport `json:"health"`
}

func NewStatefulSet(raw *appsv1.StatefulSet, checkHealth bool) StatefulSet {
	ss := StatefulSet{
		StatefulSet:  *raw,
		TenantInfo:   ParseTenantInfo(raw.Namespace),
		HealthReport: HealthReport{},
	}

	ss.Kind = "StatefulSet"

	if checkHealth {
		ss.CheckHealth()
	}

	return ss
}

func (ss *StatefulSet) CheckHealth() {
	report := NewHealthReport()

	if int32(*ss.Spec.Replicas) != int32(ss.Status.ReadyReplicas) {
		msg := fmt.Sprintf("The number of desired replicas [%d] does not match the number of ready replicas [%d].", *ss.Spec.Replicas, ss.Status.ReadyReplicas)
		if int32(ss.Status.ReadyReplicas) == 0 {
			report.AddError(msg)
		} else {
			report.AddWarning(msg)
		}
	}

	// If nobody said we're unhealthy, that must mean we are health, right?
	report.FailHealthy()

	ss.HealthReport = report
}
