package models

import (
	"fmt"
	"time"
)

type HealthyStatus string

const (
	StatusHealthy   HealthyStatus = "True"
	StatusIgnored   HealthyStatus = "Ignored"
	StatusUnhealthy HealthyStatus = "False"
	StatusUnknown   HealthyStatus = "Unknown"
	StatusWarn      HealthyStatus = "Warn"
)

type HealthReportInterface interface {
	AddError(string)
	AddErrors([]string, string)
	AddWarning(string)
	AddWarnings([]string, string)
	AddAlert(string)
	AddAlerts([]string, string)
}

type HealthReport struct {
	Timestamp int64         `json:"timestamp"`
	Healthy   HealthyStatus `json:"healthy"`
	Errors    []string      `json:"errors,omitempty"`
	Warnings  []string      `json:"warnings,omitempty"`
	Alerts    []string      `json:"alerts,omitempty"`
}

func NewHealthReport() HealthReport {
	return HealthReport{
		Timestamp: time.Now().Unix(),
		Healthy:   StatusUnknown,
		Errors:    make([]string, 0),
		Warnings:  make([]string, 0),
		Alerts:    make([]string, 0),
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
	for _, msg := range msgs {
		hr.AddWarning(fmt.Sprintf("%s%s", prefix, msg))
	}
}

func (hr *HealthReport) AddAlert(msg string) {
	hr.Alerts = append(hr.Alerts, msg)
}

func (hr *HealthReport) AddAlerts(msgs []string, prefix string) {
	if prefix != "" {
		prefix = fmt.Sprintf("%s: ", prefix)
	}
	for _, msg := range msgs {
		hr.AddAlert(fmt.Sprintf("%s%s", prefix, msg))
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
