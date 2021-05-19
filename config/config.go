package config

import (
	"fmt"

	"github.com/fsnotify/fsnotify"
	"github.com/rs/zerolog/log"
	"github.com/spf13/viper"
	"github.com/zanloy/bms-api/models"
	"github.com/zanloy/bms-api/url"
	"github.com/zanloy/bms-api/wsrouter"
	"k8s.io/client-go/util/homedir"
)

var Config = models.Config{}
var logger = log.With().
	Timestamp().
	Str("component", "Config").
	Logger()

func Load() {
	logger.Debug().Msg("Loading config file...")

	viper.SetConfigName("bms-api")

	home := homedir.HomeDir()
	if home != "" {
		viper.AddConfigPath(home)
	}

	viper.AddConfigPath("/etc")
	viper.AddConfigPath("/")

	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			logger.Fatal().Err(err).Msg("Failed to find config file. Searched for /etc/bms-api/bms-api.yaml, /bms-api.yaml, and $HOME/bms-api.yaml.")
		} else {
			logger.Fatal().Err(err).Msg("Found config file but an error occured while trying to load it.")
		}
	} else {
		logger.Info().Msg(fmt.Sprintf("Loaded config file at %s.", viper.ConfigFileUsed()))
	}

	if err := viper.Unmarshal(&Config); err != nil {
		logger.Fatal().Err(err).Msg("Failed to parse config file.")
	}

	viper.WatchConfig()
	viper.OnConfigChange(reload)

	// Tell other components to load resources from Config
	wsrouter.LoadFilters(Config.Filters)

	// Celebrate!
	logger.Info().Msg(fmt.Sprintf("Config file successfully loaded from %s.", viper.ConfigFileUsed()))
}

func reload(e fsnotify.Event) {
	logger.Info().Msg("Config file changed. Reloading...")
	var newconfig = models.Config{}
	if err := viper.Unmarshal(&newconfig); err == nil {
		Config = newconfig
		url.Reload(Config.Urls) // Reload our url checks
		wsrouter.LoadFilters(Config.Filters)
	} else {
		logger.Err(err).Msg("Failed to parse config file. Retaining previous config.")
	}
}
