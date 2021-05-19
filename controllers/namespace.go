package controllers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/zanloy/bms-api/kubernetes"
	"github.com/zanloy/bms-api/models"
	"github.com/zanloy/bms-api/wsrouter"
	"k8s.io/apimachinery/pkg/labels"
)

type NamespaceController struct{}

func (ctl *NamespaceController) GetAllHealth(ctx *gin.Context) {
	namespaces, err := kubernetes.Namespaces().List(labels.Everything())
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err})
		return
	}

	var ( // Setup vars for loop to keep memory allocation down.
		ns     models.Namespace
		report models.HealthReport
	)

	results := make([]models.Namespace, len(namespaces))
	for idx, namespace := range namespaces {
		ns = models.FromK8Namespace(namespace)
		report = models.HealthReportForNamespace(*namespace, kubernetes.Factory)
		ns.Healthy = report.Healthy
		ns.Errors = report.Errors
		results[idx] = ns
	}

	ctx.JSON(http.StatusOK, results)
}

func (ctl *NamespaceController) GetNS(ctx *gin.Context) {
	namespace, err := kubernetes.Factory.Core().V1().Namespaces().Lister().Get(ctx.Param("name"))
	if err != nil {
		ctx.JSON(500, gin.H{"error": err})
		return
	}

	ns := models.FromK8Namespace(namespace)
	report := models.HealthReportForNamespace(*namespace, kubernetes.Factory)
	ns.Healthy = report.Healthy
	ns.Errors = report.Errors

	ctx.JSON(200, ns)
}

func (ctl *NamespaceController) WatchHealth(ctx *gin.Context) {
	//kubernetes.HealthUpdates.HandleRequestWithKeys(ctx.Writer, ctx.Request, map[string]interface{}{"kind": "namespace"})
	wsrouter.HandleRequest("namespace", ctx.Writer, ctx.Request)
}
