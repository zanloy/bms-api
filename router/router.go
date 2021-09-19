package router

import (
	"net/http"
	"os"

	ginlogger "github.com/gin-contrib/logger"
	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog"
	"github.com/zanloy/bms-api/controllers"
	"github.com/zanloy/bms-api/kubernetes"
)

var logger zerolog.Logger

func SetupRouter() *gin.Engine {
	/* Setup Logger */
	logger = zerolog.New(os.Stdout).With().
		Str("component", "Router").
		Timestamp().
		Logger()

	logger.Debug().Msg("Router initializing...")

	/* Setup router */
	router := gin.New()
	router.Use(gin.Recovery())
	router.Use(ginlogger.SetLogger(ginlogger.Config{
		Logger:   &logger,
		SkipPath: []string{"/ping"},
	}))

	/* Load controllers */
	var (
		namespaceCtl = new(controllers.NamespaceController)
		nodeCtl      = new(controllers.NodeController)
		podCtl       = new(controllers.PodController)
		reportCtl    = new(controllers.ReportController)
		urlCtl       = new(controllers.URLController)
		veleroCtl    = new(controllers.VeleroController)
	)

	/* Setup routes */
	router.GET("/ping", func(ctx *gin.Context) { // For our own health checks
		ctx.JSON(http.StatusOK, gin.H{"message": "pong"})
	})

	healthGrp := router.Group("/health")
	{
		healthGrp.GET("/namespaces", namespaceCtl.GetNamespacesHealth)
	}

	socketGrp := router.Group("/ws")
	{
		// This endpoint has no filter and will notify on all health updates
		socketGrp.GET("/", func(ctx *gin.Context) {
			kubernetes.HealthUpdates.HandleRequest(ctx.Writer, ctx.Request)
			//wsrouter.HandleRequest("all", ctx.Writer, ctx.Request)
		})
		socketGrp.GET("/namespaces", namespaceCtl.WatchNamespace)
		socketGrp.GET("/ns", namespaceCtl.WatchNamespace)
		socketGrp.GET("/nodes", nodeCtl.WatchNodes)
		socketGrp.GET("/urls", func(ctx *gin.Context) {
			kubernetes.HealthUpdates.HandleRequestWithKeys(ctx.Writer, ctx.Request, map[string]interface{}{"kind": "url"})
		})
	}

	//router.GET("/namespaces", namespaceCtl.GetAll) // Get all namespaces
	namespaceGrp := router.Group("/ns")
	{
		namespaceGrp.GET("/", namespaceCtl.GetNamespaces)
		namespaceGrp.GET("/:name", namespaceCtl.GetNamespace)
		namespaceGrp.GET("/:name/pods", namespaceCtl.GetPods)
		namespaceGrp.GET("/:name/ws", namespaceCtl.WatchNamespace)
	}

	nodeGrp := router.Group("/nodes")
	{
		nodeGrp.GET("/", nodeCtl.GetNodes)
		nodeGrp.GET("/:name", nodeCtl.GetNode)
	}

	router.GET("/pods", podCtl.GetPods)

	router.GET("/report", reportCtl.Create)
	urlGrp := router.Group("/urls")
	{
		urlGrp.GET("/", urlCtl.GetURLs)
	}

	veleroGrp := router.Group("/velero")
	{
		veleroGrp.GET("/backups", veleroCtl.GetBackups)
		veleroGrp.GET("/backups/:namespace", veleroCtl.GetBackups)
		veleroGrp.GET("/schedules", veleroCtl.GetSchedules)
		veleroGrp.GET("/schedules/:namespace", veleroCtl.GetSchedules)
	}

	logger.Debug().Msg("Router successfully initialized.")
	return router
}
