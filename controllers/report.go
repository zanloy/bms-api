package controllers

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"
	"github.com/zanloy/bms-api/kubernetes"
	"github.com/zanloy/bms-api/models"
	"github.com/zanloy/bms-api/url"
	"k8s.io/apimachinery/pkg/labels"
)

type ReportController struct{}

// Create is an api endpoint that is called by an external entity (usually a
// cron) to create a report and optionally store it in kubernetes.
func (ctl *ReportController) Create(ctx *gin.Context) {
	report := models.NewReport()

	// Nodes
	if k8nodes, err := kubernetes.Nodes().List(labels.Everything()); err == nil {
		// Make the nodes list the size of our nodes
		nodelist := make([]models.Node, len(k8nodes))
		// Iterate nodes
		for idx, k8node := range k8nodes {
			node := models.FromK8Node(*k8node)
			if metrics, err := kubernetes.GetNodeMetrics(node); err == nil {
				node.AddMetrics(metrics)
			}
			// Add to array
			nodelist[idx] = node
		}
		// Attach nodelist to report
		report.Nodes = nodelist
	} else {
		err = fmt.Errorf("Failed to get nodes from kubernetes: %w", err)
		logAndAppendError(err, &report)
	}

	// DaemonSets
	if k8daemonsets, err := kubernetes.DaemonSets("").List(labels.Everything()); err == nil {
		for _, k8daemonset := range k8daemonsets {
			daemonset := models.FromK8DaemonSet(*k8daemonset)
			if daemonset.Healthy != models.StatusHealthy {
				report.UnhealthyDaemonSets = append(report.UnhealthyDaemonSets, daemonset)
			}
		}
	} else {
		err = fmt.Errorf("Failed to get daemonsets from kubernetes: %w", err)
		logAndAppendError(err, &report)
	}

	// Deployments
	if k8deployments, err := kubernetes.Deployments("").List(labels.Everything()); err == nil {
		// Iterate deployments
		for _, k8deployment := range k8deployments {
			deployment := models.FromK8Deployment(*k8deployment)
			if deployment.Healthy != models.StatusHealthy {
				report.UnhealthyDeployments = append(report.UnhealthyDeployments, deployment)
			}
		}
	} else {
		err = fmt.Errorf("Failed to get deployments from kubernetes: %w", err)
		logAndAppendError(err, &report)
	}

	// Pods
	if k8pods, err := kubernetes.Pods("").List(labels.Everything()); err == nil {
		for _, k8pod := range k8pods {
			pod := models.FromK8Pod(*k8pod)
			if pod.Healthy != models.StatusHealthy {
				report.UnhealthyPods = append(report.UnhealthyPods, pod)
			}
		}
	} else {
		err = fmt.Errorf("Failed to get pods from kubernetes: %w", err)
		logAndAppendError(err, &report)
	}

	// StatefulSets
	if k8statefulsets, err := kubernetes.StatefulSets("").List(labels.Everything()); err == nil {
		for _, k8statefulset := range k8statefulsets {
			statefulset := models.FromK8StatefulSet(*k8statefulset)
			if statefulset.Healthy != models.StatusHealthy {
				report.UnhealthyStatefulSets = append(report.UnhealthyStatefulSets, statefulset)
			}
		}
	} else {
		err = fmt.Errorf("Failed to get statefulsets from kubernetes: %w", err)
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
	logger = log.With().
		Timestamp().
		Str("component", "ReportController").
		Logger()
	logger.Err(err)
	report.Errors = append(report.Errors, err.Error())
}
