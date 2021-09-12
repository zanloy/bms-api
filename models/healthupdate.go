package models

import (
	"encoding/json"
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type HealthColor string

const (
	HealthGreen   HealthColor = "green"
	HealthYellow  HealthColor = "yellow"
	HealthRed     HealthColor = "red"
	HealthUnknown HealthColor = "unknown"
)

type HealthUpdate struct {
	metav1.TypeMeta      `json:",inline"`
	Kind                 string `json:"kind"`
	Name                 string `json:"name"`
	Namespace            string `json:"namespace"`
	HealthReport         `json:",inline"`
	Action               string        `json:"action"`
	PreviousHealthReport *HealthReport `json:"previousHealthReport,omitempty"`
}

// Timestamps the payload and returns the byte string
func (hu *HealthUpdate) ToMsg() []byte {
	hu.Timestamp = time.Now().Unix()

	if bytes, err := json.Marshal(hu); err == nil {
		return bytes
	} else {
		return []byte{}
	}
}
