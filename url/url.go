package url // import github.com/zanloy/bms-api/url

import (
	"fmt"
	"sync"
	"time"

	"k8s.io/apimachinery/pkg/util/wait"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/zanloy/bms-api/models"
	"gopkg.in/olahol/melody.v1"
)

var (
	logger        zerolog.Logger
	targets       []models.URLCheck
	mutex         = sync.Mutex{}
	HealthUpdates = melody.New()
)

func Start(targetsin []models.URLCheck, stopCh <-chan struct{}) {
	logger = log.With().
		Timestamp().
		Str("component", "url").
		Logger()

	logger.Info().Msg("Starting URL checker.")
	Reload(targetsin)
	runChecks() // To get initial health

	go wait.Until(runChecks, time.Minute, stopCh)

	<-stopCh
	logger.Info().Msg("Stopping URL checker.")
}

func GetResults() []models.URLCheck {
	return targets
}

func Reload(targetsin []models.URLCheck) {
	mutex.Lock()
	defer mutex.Unlock()

	targets = targetsin
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
	for idx, _ := range targets {
		go func(target *models.URLCheck) {
			defer wg.Done()
			var prevHealthy models.HealthyStatus = models.StatusUnknown
			if target.Report != nil {
				prevHealthy = target.Report.Healthy
			}
			logger.Debug().
				Str("previous_healthy", string(prevHealthy)).
				Msg(fmt.Sprintf("Checking %s", target.Url))
			start_time := time.Now()
			report := target.Check()
			logger.Debug().
				Str("previous_healthy", string(prevHealthy)).
				Str("healthy", string(report.Healthy)).
				Msg(fmt.Sprintf("Completed check of %s in %.2fs.", target.Url, time.Since(start_time).Seconds()))
			if report.Healthy != prevHealthy {
				// Alert the press!
				update := models.HealthUpdate{
					Action:          "update",
					Kind:            "url",
					Name:            target.Name,
					Healthy:         report.Healthy,
					PreviousHealthy: prevHealthy,
					Errors:          report.Errors,
				}

				HealthUpdates.BroadcastFilter(update.ToMsg(), func(s *melody.Session) bool {
					sessKind, ok := s.Get("kind")
					if !ok || sessKind == "url" || sessKind == "all" {
						return true
					} else {
						return false
					}
				})
			}
		}(&targets[idx])
	}

	logger.Debug().Msg("Waiting for worker pool to finish...")
	wg.Wait()
	logger.Debug().Msg(fmt.Sprintf("Worker pool completed in %.2fs.", time.Since(start_time).Seconds()))
}
