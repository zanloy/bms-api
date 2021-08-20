package wsrouter

import (
	"fmt"
	"net/http"
	"sort"
	"strings"

	"github.com/zanloy/bms-api/models"
	"gopkg.in/olahol/melody.v1"
)

var (
	Objects = []string{
		"daemonset",
		"deployment",
		"namespace",
		"node",
		"pod",
		"statefulset",
		"url",
	}
	outboxes = map[string]melody.Melody{"all": *melody.New()}
)

func init() {
	// Setup outboxes for all our monitored objects
	for _, obj := range Objects {
		outboxes[obj] = *melody.New()
	}
}

var filters = map[string][]string{}

func HandleRequest(kind string, w http.ResponseWriter, r *http.Request) error {
	if !hasOutbox(kind) {
		return fmt.Errorf("there is no outbox in wsrouter for %s", kind)
	}
	box := outboxes[kind]
	return box.HandleRequest(w, r)
}

func LoadFilters(input []models.Filter) {
	// Initialize our data structure
	filters = map[string][]string{} // (re)init map
	for _, obj := range Objects {
		filters[obj] = make([]string, 0)
	}

	// Process filters from Config
	for _, filter := range input {
		if filter.Kind == "" {
			// A filter is invalid if Name is missing
			continue
		}

		filter.Kind = strings.ToLower(filter.Kind)

		if _, ok := filters[filter.Kind]; ok {
			filters[filter.Kind] = append(filters[filter.Kind], filter.Name)
		}
	}

	// Sort our arrays for easy binary searching later
	for key := range filters {
		sort.Strings(filters[key])
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
	if outbox, ok := outboxes[update.Kind]; ok {
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
