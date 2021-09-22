package controllers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/gobwas/ws"
	"github.com/zanloy/bms-api/kubernetes"
	"github.com/zanloy/bms-api/models"
	"github.com/zanloy/bms-api/wsrouter"
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

func (ctl *NamespaceController) GetNamespacesHealth(ctx *gin.Context) {
	if namespaces, errs := kubernetes.GetNamespaces(); len(errs) == 0 {
		var reports []models.HealthReport
		for _, ns := range namespaces {
			reports = append(reports, ns.HealthReport)
		}
		ctx.JSON(http.StatusOK, reports)
	} else {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": errs})
	}
}

func (ctl *NamespaceController) WatchNamespace(ctx *gin.Context) {
	namespace := ctx.Param("name")
	conn, _, _, err := ws.UpgradeHTTP(ctx.Request, ctx.Writer)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err})
		return
	}

	consumer := wsrouter.Consumer{
		Connection: conn,
		Filter: &models.Filter{
			Action:    "",
			Kind:      "Namespace",
			Name:      namespace,
			Namespace: "",
		},
	}
	consumer.Start()

	go func() {
		defer conn.Close()

		wsrouter.SubscribeToHealth(&consumer)
		<-consumer.Quit
	}()
}

func (ctl *NamespaceController) WatchNamespaces(ctx *gin.Context) {
	kubernetes.HealthUpdates.HandleRequestWithKeys(ctx.Writer, ctx.Request, map[string]interface{}{"kind": "namespace"})
	//wsrouter.HandleRequest("namespace", ctx.Writer, ctx.Request)
}

func (ctl *NamespaceController) WatchNamespaceEvents(ctx *gin.Context) {
	conn, _, _, err := ws.UpgradeHTTP(ctx.Request, ctx.Writer)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err})
		return
	}

	consumer := wsrouter.Consumer{Connection: conn}
	consumer.Start()

	go func() {
		defer conn.Close()

		wsrouter.SubscribeToEvents(&consumer)
		<-consumer.Quit
	}()
}

func (ctl *NamespaceController) GetPods(ctx *gin.Context) {
	if pods, err := kubernetes.GetPods(ctx.Param("name")); err == nil {
		ctx.JSON(http.StatusOK, pods)
	} else {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err})
	}
}
