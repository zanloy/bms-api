package controllers // import "github.com/zanloy/bms-api/controllers"

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/zanloy/bms-api/kubernetes"
)

type VeleroController struct{}

func (ctl *VeleroController) GetBackups(ctx *gin.Context) {
	namespace := ctx.Param("namespace")
	if backups, err := kubernetes.GetVeleroBackups(namespace); err == nil {
		ctx.JSON(http.StatusOK, backups)
	} else {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err})
	}
}

func (ctl *VeleroController) GetSchedules(ctx *gin.Context) {
	namespace := ctx.Param("namespace")
	if schedules, err := kubernetes.GetVeleroSchedules(namespace); err == nil {
		ctx.JSON(http.StatusOK, schedules)
	} else {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err})
	}
}
