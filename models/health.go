package models

import (
	"encoding/json"
	"time"
)

type HealthColor string

const (
	HealthGreen   HealthColor = "green"
	HealthYellow  HealthColor = "yellow"
	HealthRed     HealthColor = "red"
	HealthUnknown HealthColor = "unknown"
)

type Health struct {
	Timestamp int64       `json:"timestamp,omitempty"`
	Error     string      `json:"error,omitempty"`
	Healthy   bool        `json:"healthy"`
	Light     HealthColor `json:"light"`
	Status    string      `json:"status"`
}

type HealthUpdate struct {
	Timestamp       int64         `json:"timestamp"`
	Action          string        `json:"action"`
	Kind            string        `json:"kind"`
	Namespace       string        `json:"namespace"`
	Name            string        `json:"name"`
	Healthy         HealthyStatus `json:"healthy"`
	PreviousHealthy HealthyStatus `json:"previous_healthy,omitempty"`
	Errors          []string      `json:"errors,omitempty"`
}

func (hu *HealthUpdate) ToMsg() []byte {
	hu.Timestamp = time.Now().Unix()

	bytes, err := json.Marshal(hu)
	if err != nil {
		return []byte{}
	}

	return bytes
}
