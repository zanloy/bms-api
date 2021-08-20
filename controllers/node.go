package controllers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/zanloy/bms-api/kubernetes"
)

type NodeController struct{}

func (ctl *NodeController) GetNode(ctx *gin.Context) {
	if node, err := kubernetes.GetNode(ctx.Param("name")); err == nil {
		ctx.JSON(http.StatusOK, node)
	} else {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err})
	}
}

func (ctl *NodeController) GetNodes(ctx *gin.Context) {
	nodes, err := kubernetes.GetNodes()
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error": err,
		})
		return
	}
	ctx.JSON(http.StatusOK, nodes)
}

func (ctl *NodeController) WatchHealth(ctx *gin.Context) {
	kubernetes.HealthUpdates.HandleRequestWithKeys(ctx.Writer, ctx.Request, map[string]interface{}{"kind": "node"})
}
