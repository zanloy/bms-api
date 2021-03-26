package controllers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"
	"github.com/zanloy/bms-api/kubernetes"
	"github.com/zanloy/bms-api/models"
	"k8s.io/apimachinery/pkg/labels"
)

var logger = log.With().Timestamp().Str("component", "NodeController").Logger()

type NodeController struct{}

func (ctl *NodeController) GetAllHealth(ctx *gin.Context) {
	logger := logger.With().
		Timestamp().
		Str("function", "GetAllHealth").
		Logger()

	// Get all Nodes
	results, err := kubernetes.Nodes().List(labels.Everything())
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error": err,
		})
		logger.Err(err).Msg("An error occurred while trying to pull nodes from kubernetes.")
		return
	}

	nodes := make([]models.HealthReport, len(results))
	for idx, node := range results {
		nodes[idx], _ = models.HealthReportFor(node, kubernetes.Factory)
	}

	ctx.JSON(http.StatusOK, nodes)
}

func (ctl *NodeController) WatchHealth(ctx *gin.Context) {
	kubernetes.HealthUpdates.HandleRequestWithKeys(ctx.Writer, ctx.Request, map[string]interface{}{"kind": "node"})
}
