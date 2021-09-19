package config // import "github.com/zanloy/bms-api/config"

/*
	This package is responsible for monitoring the config file and updating the
	various elements of the bms-api application.
*/

import (
	"bytes"
	_ "embed"
	"fmt"

	"github.com/fsnotify/fsnotify"
	"github.com/rs/zerolog/log"
	"github.com/spf13/viper"
	"github.com/zanloy/bms-api/models"
	"github.com/zanloy/bms-api/url"
	"k8s.io/client-go/util/homedir"
)

var logger = log.With().Str("component", "config").Logger()

//go:embed default.yaml
var defaultConfig []byte

func Load() {
	logger.Info().Msg("Loading config file...")

	viper.SetConfigName("bms-api")

	home := homedir.HomeDir()
	if home != "" {
		viper.AddConfigPath(home)
	}

	viper.AddConfigPath("/etc")
	viper.AddConfigPath("/")

	// Find the config file.
	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			logger.Info().Msg("No config file was found... loading sane defaults.")
			viper.ReadConfig(bytes.NewBuffer(defaultConfig))
		} else {
			logger.Fatal().Err(err).Msg("Found config file but an error occured while trying to load it.")
		}
	} else {
		logger.Info().Msg(fmt.Sprintf("Loaded config file at %s.", viper.ConfigFileUsed()))
	}

	var tmp models.Config
	viper.UnmarshalExact(&tmp)
	logger.Info().Interface("config", tmp).Msg("WORK!")

	if err := viper.UnmarshalExact(&models.Config{}); err != nil {
		logger.Fatal().Err(err).Msg("Failed to parse config file.")
	}

	viper.WatchConfig()
	viper.OnConfigChange(reload)

	// Tell other components to load resources from Config
	announce()

	// Celebrate!
	logger.Info().Msg(fmt.Sprintf("Config file successfully loaded from %s.", viper.ConfigFileUsed()))
}

func announce() {
	url.Load()
	//wsrouter.LoadFilters()
}

func reload(e fsnotify.Event) {
	logger.Info().Msg("Config file changed. Reloading...")
	if err := viper.UnmarshalExact(&models.Config{}); err == nil {
		announce() // Tell the world we got new datas!
	} else {
		logger.Err(err).Msg("Failed to parse config file. Retaining previous config.")
	}
}
