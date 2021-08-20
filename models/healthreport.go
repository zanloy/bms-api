package models

import (
	"fmt"
	"time"

	"github.com/zanloy/bms-api/helpers"
)

type HealthyStatus string

const (
	StatusHealthy   HealthyStatus = "True"
	StatusIgnored   HealthyStatus = "Ignored"
	StatusUnhealthy HealthyStatus = "False"
	StatusUnknown   HealthyStatus = "Unknown"
	StatusWarn      HealthyStatus = "Warn"
)

type HealthReport struct {
	Timestamp   int64         `json:"timestamp"`
	Action      string        `json:"action,omitempty"`
	Kind        string        `json:"kind"`
	Namespace   string        `json:"namespace,omitempty"`
	Name        string        `json:"name"`
	Tenant      string        `json:"tenant,omitempty"`
	Environment string        `json:"environment,omitempty"`
	Healthy     HealthyStatus `json:"healthy"`
	Text        string        `json:"text,omitempty"`
	Errors      []string      `json:"errors,omitempty"`
	Warnings    []string      `json:"warnings,omitempty"`
}

func NewHealthReport() HealthReport {
	return HealthReport{
		Timestamp: time.Now().Unix(),
		Healthy:   StatusUnknown,
	}
}

func NewHealthReportFor(kind string, name string, namespace string) HealthReport {
	tenant, env := helpers.ParseTenantAndEnv(namespace)
	return HealthReport{
		Timestamp:   time.Now().Unix(),
		Action:      "",
		Kind:        kind,
		Namespace:   namespace,
		Name:        name,
		Tenant:      tenant,
		Environment: env,
		Healthy:     StatusUnknown,
		Text:        "",
		Errors:      []string{},
		Warnings:    []string{},
	}
}

func (hr *HealthReport) AddWarning(msg string) {
	if hr.Healthy != StatusUnhealthy {
		hr.Healthy = StatusWarn
	}
	hr.Warnings = append(hr.Warnings, msg)
}

func (hr *HealthReport) AddWarnings(msgs []string, prefix string) {
	if prefix != "" {
		prefix = fmt.Sprintf("%s: ", prefix)
	}
	for msg := range msgs {
		hr.AddWarning(fmt.Sprintf("%v%v", prefix, msg))
	}
}

func (hr *HealthReport) AddError(msg string) {
	hr.Healthy = StatusUnhealthy
	hr.Errors = append(hr.Errors, msg)
}

func (hr *HealthReport) AddErrors(msgs []string, prefix string) {
	if prefix != "" {
		prefix = fmt.Sprintf("%s: ", prefix)
	}
	for _, msg := range msgs {
		hr.AddError(fmt.Sprintf("%s%s", prefix, msg))
	}
}

func (hr *HealthReport) FailHealthy() {
	if hr.Healthy == StatusUnknown {
		hr.Healthy = StatusHealthy
	}
}

func (hr *HealthReport) FailUnhealthy() {
	if hr.Healthy == StatusUnknown {
		hr.Healthy = StatusUnhealthy
	}
}

func (hr *HealthReport) FoldIn(subhr HealthReport, prefix string) {
	hr.AddErrors(subhr.Errors, prefix)
	hr.AddWarnings(subhr.Warnings, prefix)
}
