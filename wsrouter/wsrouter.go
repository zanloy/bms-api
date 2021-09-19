package wsrouter

import (
	"fmt"
	"net/http"
	"sort"
	"strings"

	"github.com/rs/zerolog/log"
	"github.com/spf13/viper"
	"github.com/zanloy/bms-api/models"
	"gopkg.in/olahol/melody.v1"
)

var (
	logger = log.With().Str("component", "wsrouter").Logger()

	Objects = []string{
		"daemonset",
		"deployment",
		"namespace",
		"node",
		"pod",
		"statefulset",
		"url",
	}

	filters       []models.Filter
	filtersByKind map[string][]models.Filter
	outboxes      map[string]melody.Melody
)

func init() {
	// Setup logger
	logger = log.With().Str("component", "wsrouter").Logger()

	// Set sane defaults in viper
	viper.SetDefault("filters", make([]models.Filter, 0))

	// Setup outboxes for all our monitored objects
	outboxes = map[string]melody.Melody{"all": *melody.New()}
	for _, obj := range Objects {
		outboxes[obj] = *melody.New()
	}

	// Init (empty) data
	filters = make([]models.Filter, 0)
}

func HandleRequest(kind string, w http.ResponseWriter, r *http.Request) error {
	if !hasOutbox(kind) {
		return fmt.Errorf("there is no outbox in wsrouter for %s", kind)
	}
	box := outboxes[kind]
	return box.HandleRequest(w, r)
}

func LoadFilters() {
	// Just in case
	prevFilters := filters

	// Reset data
	if err := viper.UnmarshalKey("filters", filters); err == nil {
		logger.Warn().Msg("Failed to load filter from config file.")
		filters = prevFilters
	}

	// TODO: This is a fucking mess, we need to store the filters in an easily searchable way...

	// Process filters from Config
	for idx := range filters {
		if filters[idx].Kind == "" {
			// A filter is invalid if Name is missing
			continue
		}

		filters[idx].Kind = strings.ToLower(filters[idx].Kind)

		if _, ok := filtersByKind[filters[idx].Kind]; ok {
			filtersByKind[filters[idx].Kind] = append(filtersByKind[filters[idx].Kind], filters[idx])
		}
	}

	sortFilters()
}

func sortFilters() {
	// Sort our arrays for easy binary searching later
	for key := range filtersByKind {
		sort.Strings(filtersByKind[key])
	}
}

func Broadcast(update models.HealthUpdate) error {
	if idx := sort.SearchStrings(filters[update.Kind], update.Name); idx < len(filters[update.Kind]) {
		// We don't process this further
		return nil
	}

	// Broadcast to all
	{
		m := outboxes["all"]
		m.Broadcast(update.ToMsg())
	}

	// Broadcast to "kind" channel
	if outbox, ok := outboxes[strings.ToLower(update.Kind)]; ok {
		outbox.Broadcast(update.ToMsg())
	} else {
		return fmt.Errorf("there was no outbox with kind [%s] found", update.Kind)
	}

	return nil
}

func hasOutbox(name string) bool {
	for key := range outboxes {
		if key == name {
			return true
		}
	}
	return false
}

// The following is old code that is not currently used.
type filterFunc func(*melody.Session) bool

func filterKind(s *melody.Session, kind string) bool {
	kind = strings.ToLower(kind)
	if sessKind, ok := s.Get("kind"); ok {
		sessKind = strings.ToLower(sessKind.(string))
		if sessKind == strings.ToLower(kind) || sessKind == "all" {
			return true
		}
	} else {
		// The "kind" key didn't exist in the Session so assume no filter
		return true
	}

	return false
}

func filterAllowAll(s *melody.Session) bool { return true }

func filterDaemonSet(s *melody.Session) bool {
	return filterKind(s, "daemonset")
}

func filterDeployment(s *melody.Session) bool {
	return filterKind(s, "deployment")
}

func filterPod(s *melody.Session) bool {
	return filterKind(s, "pod")
}

func filterNamespace(s *melody.Session) bool {
	return filterKind(s, "namespace")
}

func filterNode(s *melody.Session) bool {
	return filterKind(s, "node")
}

func filterStatefulSet(s *melody.Session) bool {
	return filterKind(s, "statefulset")
}

func filterURL(s *melody.Session) bool {
	return filterKind(s, "url")
}
