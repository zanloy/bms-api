package controllers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/zanloy/bms-api/kubernetes"
)

type NamespaceController struct{}

func (ctl *NamespaceController) GetNamespace(ctx *gin.Context) {
	if ns, err := kubernetes.GetNamespaceWithEvents(ctx.Param("name")); err == nil {
		ctx.JSON(http.StatusOK, ns)
	} else {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err})
	}
}

func (ctl *NamespaceController) GetNamespaces(ctx *gin.Context) {
	if results, errs := kubernetes.GetNamespaces(); len(errs) == 0 {
		ctx.JSON(http.StatusOK, results)
	} else {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": errs})
	}
}

func (ctl *NamespaceController) WatchNamespace(ctx *gin.Context) {
	namespace := ctx.Param("name")
	kubernetes.HealthUpdates.HandleRequestWithKeys(ctx.Writer, ctx.Request, map[string]interface{}{"namespace": namespace})
}

func (ctl *NamespaceController) WatchNamespaces(ctx *gin.Context) {
	kubernetes.HealthUpdates.HandleRequestWithKeys(ctx.Writer, ctx.Request, map[string]interface{}{"kind": "namespace"})
	//wsrouter.HandleRequest("namespace", ctx.Writer, ctx.Request)
}

func (ctl *NamespaceController) GetPods(ctx *gin.Context) {
	if pods, err := kubernetes.GetPods(ctx.Param("name")); err == nil {
		ctx.JSON(http.StatusOK, pods)
	} else {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err})
	}
}
