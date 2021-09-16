package main

import (
	"flag"
	"io"
	"os"
	"os/signal"
	"syscall"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog"
	"github.com/spf13/pflag"
	"github.com/zanloy/bms-api/config"
	"github.com/zanloy/bms-api/kubernetes"
	"github.com/zanloy/bms-api/router"
	"github.com/zanloy/bms-api/url"
	"k8s.io/klog"
)

func main() {
	/* Setup logging */

	zerolog.SetGlobalLevel(zerolog.InfoLevel)
	if gin.IsDebugging() {
		zerolog.SetGlobalLevel(zerolog.DebugLevel)
	}

	logger := zerolog.New(os.Stdout).With().
		Str("component", "main").
		Timestamp().
		Logger()

	/* Parse flags */
	var kubeconfig *string = pflag.StringP("kubeconfig", "k", "", "location of kubeconfig if not ~/.kube/config")
	klog.InitFlags(flag.CommandLine)
	pflag.CommandLine.AddGoFlagSet(flag.CommandLine)
	// This turns off the logging from the client-go libraries.
	// We do not want them putting non-json logs in our output.
	klog.SetOutput(io.Discard)
	pflag.Parse()

	logger.Info().Msg("Starting bms-api...")

	/* Setup config */
	config.Load()

	/* Setup stop chan */
	stopCh := make(chan struct{})
	defer close(stopCh)

	/* Setup os.Signal catcher */
	signalCh := make(chan os.Signal, 1)
	signal.Notify(signalCh, syscall.SIGINT, syscall.SIGTERM)
	go func(stopCh chan struct{}) {
		for sig := range signalCh {
			// Graceful shutdown
			logger.Info().Str("signal", sig.String()).Msg("Gracefully shutting down...")
			close(stopCh)
			os.Exit(0)
		}
	}(stopCh)

	/* Setup kubernetes */
	if err := kubernetes.Init(*kubeconfig); err != nil {
		logger.Fatal().Str("element", "kubernetes").Err(err).Msg("Failed to initialize Kubernetes.")
	}

	kubernetes.Start(stopCh)

	/* Setup URL checker */
	go url.Start(stopCh)

	/* Setup router */
	router := router.SetupRouter()

	/* Start the http server */
	logger.Info().Msg("Starting HTTP listener...")
	router.Run()
	logger.Info().Msg("HTTP listener stopped.")
}
