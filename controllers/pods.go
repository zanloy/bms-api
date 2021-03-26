package controllers

import (
	"net/http"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/labels"

	"github.com/gin-gonic/gin"
	"github.com/zanloy/bms-api/kubernetes"
	"github.com/zanloy/bms-api/models"
)

type PodWithReport struct {
	corev1.Pod
	Health models.HealthReport `json:"health"`
}

type PodController struct{}

func (ctl *PodController) GetAllHealth(ctx *gin.Context) {
	// Get all Pods
	results, err := kubernetes.Pods("").List(labels.Everything())
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error": err,
		})
		logger.Err(err)
		return
	}

	pods := make([]models.HealthReport, len(results))
	for idx, pod := range results {
		pods[idx], _ = models.HealthReportFor(pod, kubernetes.Factory)
	}

	ctx.JSON(http.StatusOK, pods)
}

func (ctl *PodController) GetSingle(ctx *gin.Context) {
	name := ctx.Param("name")
	namespace := ctx.Param("ns")

	if name == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"errors": "name param can not be nil.",
		})
		return
	}

	pod, err := kubernetes.Pods(namespace).Get(name)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"errors": err.Error(),
		})
	}

	report := models.HealthReportForPod(pod)
	result := PodWithReport{
		Pod:    *pod,
		Health: report,
	}

	// Celebrate!
	ctx.JSON(http.StatusOK, result)
}

func (ctl *PodController) WatchHealth(ctx *gin.Context) {
	kubernetes.HealthUpdates.HandleRequestWithKeys(ctx.Writer, ctx.Request, map[string]interface{}{"kind": "pod"})
}
