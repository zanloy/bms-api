package config

import "github.com/zanloy/bms-api/models"

func Filters() []models.Filter {
	return Config.Filters
}

func Namespace() string {
	if Config.Namespace == "" {
		return "bms"
	} else {
		return Config.Namespace
	}
}
