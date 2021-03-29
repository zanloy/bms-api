package router

import (
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
	)

	/* Setup routes */
	router.GET("/ping", func(ctx *gin.Context) { // For our own health checks
		ctx.JSON(200, gin.H{
			"message": "pong",
		})
	})

	healthGrp := router.Group("/health")
	{
		healthGrp.GET("/namespaces", namespaceCtl.GetAllHealth)
		healthGrp.GET("/namespaces/ws", namespaceCtl.WatchHealth)
		healthGrp.GET("/nodes", nodeCtl.GetAllHealth)
		healthGrp.GET("/nodes/ws", nodeCtl.WatchHealth)
		healthGrp.GET("/pods", podCtl.GetAllHealth)
		healthGrp.GET("/pods/ws", podCtl.WatchHealth)
		healthGrp.GET("/urls", urlCtl.GetAll)
		healthGrp.GET("/urls/ws", urlCtl.WatchHealth)
		// This endpoint has no filter and will notify on all health updates
		healthGrp.GET("/ws", func(ctx *gin.Context) {
			kubernetes.HealthUpdates.HandleRequest(ctx.Writer, ctx.Request)
		})
	}

	//router.GET("/namespaces", namespaceCtl.GetAll) // Get all namespaces
	namespaceGrp := router.Group("/ns")
	{
		namespaceGrp.GET("/:name", namespaceCtl.GetNS)
	}

	reportsGrp := router.Group("/reports")
	{
		reportsGrp.GET("/create", reportCtl.Create)
	}

	logger.Debug().Msg("Router successfully initialized.")
	return router
}
