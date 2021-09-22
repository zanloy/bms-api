package controllers

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"
	"github.com/zanloy/bms-api/kubernetes"
	"github.com/zanloy/bms-api/models"
	"github.com/zanloy/bms-api/url"
)

type ReportController struct{}

// Create is an api endpoint that is called by an external entity (usually a
// cron) to create a report and optionally store it in kubernetes.
func (ctl *ReportController) Create(ctx *gin.Context) {
	report := models.NewReport()

	// Nodes
	if nodes, err := kubernetes.GetNodes(); err == nil {
		// Attach nodelist to report
		report.Nodes = nodes
	} else {
		err = fmt.Errorf("failed to get nodes from kubernetes: %w", err)
		logAndAppendError(err, &report)
	}

	// DaemonSets
	if daemonsets, err := kubernetes.GetAllDaemonSets(); err == nil {
		for _, daemonset := range daemonsets {
			if daemonset.HealthReport.Healthy != models.StatusHealthy {
				report.UnhealthyDaemonSets = append(report.UnhealthyDaemonSets, daemonset)
			}
		}
	} else {
		err = fmt.Errorf("failed to get daemonsets from kubernetes: %w", err)
		logAndAppendError(err, &report)
	}

	// Deployments
	if deployments, err := kubernetes.GetAllDeployments(); err == nil {
		for _, deployment := range deployments {
			if deployment.HealthReport.Healthy != models.StatusHealthy {
				report.UnhealthyDeployments = append(report.UnhealthyDeployments, deployment)
			}
		}
	} else {
		err = fmt.Errorf("failed to get deployments from kubernetes: %w", err)
		logAndAppendError(err, &report)
	}

	// Pods
	if pods, err := kubernetes.GetAllPods(); err == nil {
		for _, pod := range pods {
			if pod.HealthReport.Healthy != models.StatusHealthy {
				report.UnhealthyPods = append(report.UnhealthyPods, pod)
			}
		}
	} else {
		err = fmt.Errorf("failed to get pods from kubernetes: %w", err)
		logAndAppendError(err, &report)
	}

	// StatefulSets
	if statefulsets, err := kubernetes.GetAllStatefulSets(); err == nil {
		for _, statefulset := range statefulsets {
			if statefulset.HealthReport.Healthy != models.StatusHealthy {
				report.UnhealthyStatefulSets = append(report.UnhealthyStatefulSets, statefulset)
			}
		}
	} else {
		err = fmt.Errorf("failed to get statefulsets from kubernetes: %w", err)
		logAndAppendError(err, &report)
	}

	// URLs
	report.URLs = url.GetTargets()

	// Return results to client
	ctx.JSON(http.StatusOK, report)
}

// DISABLED
/*
func (ctl *ReportController) List(ctx *gin.Context) {
	reports, err := storage.ListReports()
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err})
		return
	}

	ctx.JSON(http.StatusOK, reports)
}
*/

func logAndAppendError(err error, report *models.Report) {
	logger := log.With().
		Timestamp().
		Str("component", "ReportController").
		Logger()
	logger.Err(err)
	report.Errors = append(report.Errors, err.Error())
}
