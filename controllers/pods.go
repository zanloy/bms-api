package controllers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/zanloy/bms-api/kubernetes"
)

type PodController struct{}

type podParams struct {
	Name      string `form:"name" json:"name" binding:"required"`
	Namespace string `form:"namespace" json:"namespace" binding:"required"`
}

func (ctl *PodController) GetPod(ctx *gin.Context) {
	var podparams podParams
	if err := ctx.ShouldBindJSON(&podparams); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if pod, err := kubernetes.GetPod(podparams.Namespace, podparams.Name); err != nil {
		ctx.JSON(http.StatusOK, pod)
	} else {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
	}
}

func (ctl *PodController) GetPods(ctx *gin.Context) {
	if pods, err := kubernetes.GetPods(""); err == nil {
		ctx.JSON(http.StatusOK, pods)
	} else {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err})
	}
}

func (ctl *PodController) WatchHealth(ctx *gin.Context) {
	kubernetes.HealthUpdates.HandleRequestWithKeys(ctx.Writer, ctx.Request, map[string]interface{}{"kind": "pod"})
}
