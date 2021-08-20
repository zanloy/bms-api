package controllers

import (
	"github.com/gin-gonic/gin"
	"github.com/zanloy/bms-api/kubernetes"
	"github.com/zanloy/bms-api/url"
)

type URLController struct{}

func (ctl *URLController) GetURLs(ctx *gin.Context) {
	results := url.GetResults()
	ctx.JSON(200, results)
}

func (ctl *URLController) WatchURLs(ctx *gin.Context) {
	kubernetes.HealthUpdates.HandleRequestWithKeys(ctx.Writer, ctx.Request, map[string]interface{}{"kind": "url"})
}
