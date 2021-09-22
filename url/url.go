package url // import github.com/zanloy/bms-api/url

import (
	"fmt"
	"sync"
	"time"

	"k8s.io/apimachinery/pkg/util/wait"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/spf13/viper"
	"github.com/zanloy/bms-api/kubernetes"
	"github.com/zanloy/bms-api/models"
)

var (
	logger  zerolog.Logger
	targets []models.URLCheck
	mutex   = sync.Mutex{}
)

func init() {
	logger = log.With().Str("component", "url").Logger()
	viper.SetDefault("urls", make([]models.URLCheck, 0))
}

// GetTargets will return an array of all the targets being monitored.
func GetTargets() []models.URLCheck {
	return targets
}

// Start will begin monitoring all URLCheck targets until told to stop via the
// stopCh channel.
func Start(stopCh <-chan struct{}) {
	logger.Info().Msg("Starting URL checker.")
	runChecks() // To get initial health

	go wait.Until(runChecks, time.Minute, stopCh)

	<-stopCh
	logger.Info().Msg("Stopping URL checker.")
}

func GetResults() []models.URLCheck {
	return targets
}

func Load() {
	mutex.Lock()
	defer mutex.Unlock()

	var targetsin []models.URLCheckMeta
	if err := viper.UnmarshalKey("urls", &targetsin); err != nil {
		logger.Fatal().Err(err).Msg("Failed to load URLs from config.")
	}

	targets = make([]models.URLCheck, 0)
	for _, target := range targetsin {
		if target.Type == "" {
			target.Type = models.RespTypeHTTPStatus
		}
		targets = append(targets, models.URLCheck{Meta: target})
		logger.Info().Msg(fmt.Sprintf("Loaded URL: %s", target.Name))
	}

	logger.Info().Msg(fmt.Sprintf("Loaded %d URLs.", len(targets)))
}

func runChecks() {
	var (
		wg         = sync.WaitGroup{}
		start_time = time.Now()
	)

	logger.Debug().Msg(fmt.Sprintf("Checking %d URLs.", len(targets)))
	// TODO: Add a timeout
	mutex.Lock()
	defer mutex.Unlock()

	wg.Add(len(targets))
	for idx := range targets {
		go func(target *models.URLCheck) {
			defer wg.Done()
			prevHealthReport := target.HealthReport
			logger.Debug().
				Str("previous_healthy", string(prevHealthReport.Healthy)).
				Msg(fmt.Sprintf("Checking %s", target.Meta.Url))
			start_time := time.Now()

			target.Check()

			logger.Debug().
				Str("previous_healthy", string(prevHealthReport.Healthy)).
				Str("healthy", string(target.HealthReport.Healthy)).
				Msg(fmt.Sprintf("Completed check of %s in %.2fs.", target.Meta.Url, time.Since(start_time).Seconds()))

			if target.HealthReport.Healthy != prevHealthReport.Healthy {
				// Alert the press!
				update := models.HealthUpdate{
					Kind:      "Url",
					Name:      target.Meta.Name,
					Namespace: "",
					TenantInfo: models.TenantInfo{
						Name:        "platform",
						Environment: "",
					},
					Action:               "update",
					HealthReport:         target.HealthReport,
					PreviousHealthReport: &prevHealthReport,
				}

				/*
					kubernetes.HealthUpdates.BroadcastFilter(update.ToMsg(), func(s *melody.Session) bool {
						sessKind, ok := s.Get("kind")
						if !ok || sessKind == "url" || sessKind == "all" {
							return true
						} else {
							return false
						}
					})
				*/
				kubernetes.HealthUpdates.Broadcast(update.ToMsg())
			}
		}(&targets[idx])
	}

	logger.Debug().Msg("Waiting for worker pool to finish...")
	wg.Wait()
	logger.Debug().Msg(fmt.Sprintf("Worker pool completed in %.2fs.", time.Since(start_time).Seconds()))
}
