package models

type FilterAction string

const (
	// Ignore will ignore the obj completely and not show up in the UI
	FilterIgnore FilterAction = "ignore"
	// Mute will still have the obj in the UI but not broadcast health updates
	FilterMute FilterAction = "mute"
)

type Filter struct {
	Action    FilterAction `json:"action"`
	Kind      string       `json:"kind"`
	Name      string       `json:"name"`
	Namespace string       `json:"namespace"`
}

// This is the structure of our bms-api config file and will be used to
// marshal our config file.
type Config struct {
	Namespace         string     `json:"namespace"`
	Filters           []Filter   `json:"filters,omitempty"`
	Urls              []URLCheck `json:"urls,omitempty"`
	Environments      []string   `json:"environments,omitempty"`
	NotificationDelay string     `json:"notification_delay"`
}

func (config *Config) ShouldNotify(kind string, name string, namespace string) (result bool) {
	result = true
	for _, filter := range config.Filters {
		if filter.Kind == kind && filter.Name == name && filter.Namespace == namespace {
			if filter.Action == FilterIgnore || filter.Action == FilterMute {
				return false
			}
		}
	}
	return
}
