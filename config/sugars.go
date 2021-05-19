package config

import "github.com/zanloy/bms-api/models"

func Filters() []models.Filter {
	return Config.Filters
}

func MaxReports() int {
	if Config.MaxReports == 0 {
		return 48 // 2 days worth of reports as default
	} else {
		return Config.MaxReports
	}
}

func Namespace() string {
	if Config.Namespace == "" {
		return "bms"
	} else {
		return Config.Namespace
	}
}
