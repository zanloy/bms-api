package wsrouter // import "github.com/zanloy/bms-api/wsrouter"

import (
	"github.com/spf13/viper"
	"github.com/zanloy/bms-api/models"
)

var (
	filters []models.Filter
)

func LoadFilters() {
	var tmpFilters []models.Filter

	if err := viper.UnmarshalKey("filters", tmpFilters); err == nil {
		logger.Warn().Msg("Failed to load filter from config file.")
	} else {
		filters = tmpFilters
	}
}

func AllowBroadcast(update models.HealthUpdate) bool {
	for _, filter := range filters {
		if filter.Kind != "" && update.Kind != filter.Kind {
			return false
		}
		if filter.Namespace != "" && update.Namespace != filter.Namespace {
			return false
		}
		if filter.Name != "" && update.Name != filter.Name {
			return false
		}
	}
	return true
}
