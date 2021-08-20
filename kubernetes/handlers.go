package kubernetes

import (
	"strings"

	"github.com/zanloy/bms-api/models"

	"gopkg.in/olahol/melody.v1"
	"k8s.io/client-go/tools/cache"
)

func setupInformers() {
	handlers := cache.ResourceEventHandlerFuncs{
		AddFunc:    handleAdd,
		UpdateFunc: handleUpdate,
		DeleteFunc: handleDelete,
	}

	Factory.Extensions().V1beta1().
		DaemonSets().Informer().AddEventHandler(handlers)

	Factory.Extensions().V1beta1().
		Deployments().Informer().AddEventHandler(handlers)

	Factory.Core().V1().
		Namespaces().Informer().AddEventHandler(handlers)

	Factory.Core().V1().
		Nodes().Informer().AddEventHandler(handlers)

	Factory.Core().V1().
		Pods().Informer().AddEventHandler(handlers)

	Factory.Core().V1().
		Services().Informer().AddEventHandler(handlers)

	Factory.Apps().V1().
		StatefulSets().Informer().AddEventHandler(handlers)
}

type filterFunc func(*melody.Session) bool

func filterKind(s *melody.Session, kind string) bool {
	kind = strings.ToLower(kind)
	if sessKind, ok := s.Get("kind"); ok {
		sessKind = strings.ToLower(sessKind.(string))
		if sessKind == kind || sessKind == "all" {
			return true
		}
	} else {
		// The "kind" key didn't exist in the melody.Session so assume no filter
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

func broadcastNamespaceHealth(name string) {
	if name != "" {
		// Check is cache is synced
		cache.WaitForCacheSync(stopCh, Factory.Core().V1().Namespaces().Informer().HasSynced)

		if ns, err := GetNamespace(name); err == nil {
			report := ns.HealthReport
			update := models.HealthUpdate{
				Action:  "refresh",
				Kind:    "namespace",
				Name:    ns.Name,
				Healthy: report.Healthy,
				Errors:  report.Errors,
			}
			HealthUpdates.BroadcastFilter(update.ToMsg(), filterNamespace)
		}
	}
}

func handleAdd(obj interface{}) {
	var (
		report models.HealthReport
		filter filterFunc
		err    error
	)

	if report, err = ReportFor(obj); err != nil {
		logger.Err(err)
		return
	}

	update := models.HealthUpdate{
		Action:    "add",
		Kind:      report.Kind,
		Namespace: report.Namespace,
		Name:      report.Name,
		Healthy:   report.Healthy,
		Errors:    report.Errors,
	}

	//wsrouter.Broadcast(update)

	//logger.Debug().Interface("object", obj).Msg("Add event occurred.")
	//filter = FilterFor(obj)
	HealthUpdates.BroadcastFilter(update.ToMsg(), filter)
	broadcastNamespaceHealth(report.Namespace)
}

func handleUpdate(prevObj interface{}, obj interface{}) {
	var (
		filter             filterFunc
		report, prevReport models.HealthReport
	)

	prevReport, err := ReportFor(prevObj)
	if err != nil {
		logger.Err(err)
		return
	}

	if report, err = ReportFor(obj); err != nil {
		logger.Err(err)
		return
	}

	if report.Healthy != prevReport.Healthy {
		update := models.HealthUpdate{
			Action:          "update",
			Kind:            report.Kind,
			Namespace:       report.Namespace,
			Name:            report.Name,
			Healthy:         report.Healthy,
			PreviousHealthy: prevReport.Healthy,
			Errors:          report.Errors,
		}

		//wsrouter.Broadcast(update)

		//logger.Debug().Interface("object", obj).Msg("Update event occurred.")
		//filter = FilterFor(obj)
		HealthUpdates.BroadcastFilter(update.ToMsg(), filter)
		broadcastNamespaceHealth(report.Namespace)
	}
}

func handleDelete(obj interface{}) {
	var (
		filter filterFunc
	)

	if report, err := ReportFor(obj); err == nil {
		update := models.HealthUpdate{
			Timestamp:       0,
			Action:          "delete",
			Kind:            report.Kind,
			Namespace:       report.Namespace,
			Name:            report.Name,
			Healthy:         report.Healthy,
			PreviousHealthy: models.StatusUnknown,
			Errors:          report.Errors,
			Warnings:        report.Warnings,
		}

		//wsrouter.Broadcast(update)

		//logger.Debug().Interface("object", obj).Msg("Delete event occurred.")
		//filter = FilterFor(obj)
		HealthUpdates.BroadcastFilter(update.ToMsg(), filter)
		broadcastNamespaceHealth(report.Namespace)
	} else {
		logger.Err(err)
	}
}
