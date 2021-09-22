package wsrouter

import (
	"github.com/rs/zerolog/log"
	"github.com/spf13/viper"
	"github.com/zanloy/bms-api/models"
	corev1 "k8s.io/api/core/v1"
)

var (
	logger = log.With().Str("component", "wsrouter").Logger()

	EventsChan chan corev1.Event
	HealthChan chan models.HealthUpdate
	eventsCh   channel
	healthCh   channel
)

func init() {
	// Set sane defaults in viper
	viper.SetDefault("filters", make([]models.Filter, 0))

	// Init (empty) data
	filters = make([]models.Filter, 0)

	EventsChan = make(chan corev1.Event, 10)
	HealthChan = make(chan models.HealthUpdate, 10)
}

func Start(quitCh <-chan struct{}) {
	logger.Debug().Msg("Websocket Router Initializing...")
	go handleEvents()
	go handleHealthUpdates()
}

func handleEvents() {
	for {
		msg := <-EventsChan
		eventsCh.Iter(func(c *Consumer) { c.Source <- msg })
	}
}

func handleHealthUpdates() {
	for {
		msg := <-HealthChan
		if AllowBroadcast(msg) {
			healthCh.Iter(func(c *Consumer) { c.Source <- msg })
		}
	}
}

func SubscribeToEvents(consumer *Consumer) {
	eventsCh.Push(consumer)
}

func SubscribeToHealth(consumer *Consumer) {
	healthCh.Push(consumer)
}
