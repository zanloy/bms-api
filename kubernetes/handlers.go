package kubernetes

import (
	"strings"

	"github.com/zanloy/bms-api/models"
	"gopkg.in/olahol/melody.v1"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	extensionsv1beta1 "k8s.io/api/extensions/v1beta1"
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

	Factory.Apps().V1().
		StatefulSets().Informer().AddEventHandler(handlers)
}

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

func broadcastNamespaceHealth(name string) {
	if name != "" {
		// Check is cache is synced
		cache.WaitForCacheSync(stopCh, Factory.Core().V1().Namespaces().Informer().HasSynced)

		ns, err := Factory.Core().V1().Namespaces().Lister().Get(name)
		if err == nil {
			report := models.HealthReportForNamespace(*ns, Factory)
			update := models.HealthUpdate{
				Action:  "refresh",
				Kind:    "namespace",
				Name:    ns.Name,
				Healthy: report.Healthy,
				Errors:  report.Errors,
			}

			HealthUpdates.BroadcastFilter(update.ToMsg(), filterNamespace)
		} else {
			logger.Err(err).Str("namespace", name).Msg("Failed to fetch Namespace from Kubernetes.")
		}
	}
}

func handleAdd(obj interface{}) {
	var (
		//kind, namespace, name string
		report models.HealthReport
		filter filterFunc
		err    error
	)

	/*
		switch typed := obj.(type) {
		case *corev1.Namespace:
			kind = "namespace"
			namespace = ""
			name = typed.Name
			report = models.HealthReportForNamespace(*typed, Factory)
			filter = filterNamespace
		case *corev1.Node:
			kind = "node"
			namespace = ""
			name = typed.Name
			report = models.HealthReportForNode(*typed)
			filter = filterNode
		case *corev1.Pod:
			kind = "pod"
			namespace = typed.Namespace
			name = typed.Name
			report = models.HealthReportForPod(*typed)
			filter = filterPod
		case *appsv1.StatefulSet:
			kind = "statefulset"
			namespace = typed.Namespace
			name = typed.Name
			report = models.HealthReportForStatefulSet(*typed)
			filter = filterStatefulSet
		case *extensionsv1beta1.DaemonSet:
			kind = "daemonset"
			namespace = typed.Namespace
			name = typed.Name
			report = models.HealthReportForDaemonSet(*typed)
			filter = filterDaemonSet
		case *extensionsv1beta1.Deployment:
			kind = "deployment"
			namespace = typed.Namespace
			name = typed.Name
			report = models.HealthReportForDeployment(*typed)
			filter = filterDeployment
		default:
			logger.Debug().Interface("object", typed).Msg("Failed to assert type of object.")
			return
		}
	*/
	if report, err = models.HealthReportFor(obj, Factory); err != nil {
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

	//logger.Debug().Interface("object", obj).Msg("Add event occurred.")
	HealthUpdates.BroadcastFilter(update.ToMsg(), filter)
	broadcastNamespaceHealth(report.Namespace)
}

func handleUpdate(prevObj interface{}, obj interface{}) {
	var (
		//kind, namespace, name string
		filter             filterFunc
		report, prevReport models.HealthReport
	)

	prevReport, err := models.HealthReportFor(prevObj, Factory)
	if err != nil {
		logger.Err(err)
		return
	}

	/*
		switch typed := obj.(type) {
		case *corev1.Namespace:
			kind = "namespace"
			namespace = ""
			name = typed.Name
			report = models.HealthReportForNamespace(*typed, Factory)
			filter = filterNamespace
		case *corev1.Node:
			kind = "node"
			namespace = ""
			name = typed.Name
			report = models.HealthReportForNode(*typed)
			filter = filterNode
		case *corev1.Pod:
			kind = "pod"
			namespace = typed.Namespace
			name = typed.Name
			report = models.HealthReportForPod(*typed)
			filter = filterPod
		case *appsv1.StatefulSet:
			kind = "statefulset"
			namespace = typed.Namespace
			name = typed.Name
			report = models.HealthReportForStatefulSet(*typed)
			filter = filterStatefulSet
		case *extensionsv1beta1.DaemonSet:
			kind = "daemonset"
			namespace = typed.Namespace
			name = typed.Name
			report = models.HealthReportForDaemonSet(*typed)
			filter = filterDaemonSet
		case *extensionsv1beta1.Deployment:
			kind = "deployment"
			namespace = typed.Namespace
			name = typed.Name
			report = models.HealthReportForDeployment(*typed)
			filter = filterDeployment
		default:
			logger.Debug().Interface("object", typed).Msg("Failed to assert type of object.")
			return
		}
	*/
	if report, err = models.HealthReportFor(obj, Factory); err != nil {
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

		//logger.Debug().Interface("object", obj).Msg("Update event occurred.")
		HealthUpdates.BroadcastFilter(update.ToMsg(), filter)
		broadcastNamespaceHealth(report.Namespace)
	}
}

func handleDelete(obj interface{}) {
	var (
		kind, namespace, name string
		report                models.HealthReport
		filter                filterFunc
	)

	switch typed := obj.(type) {
	case *corev1.Namespace:
		kind = "namespace"
		namespace = ""
		name = typed.Name
		report = models.HealthReportForNamespace(*typed, Factory)
		filter = filterNamespace
	case *corev1.Node:
		kind = "node"
		namespace = ""
		name = typed.Name
		report = models.HealthReportForNode(*typed)
		filter = filterNode
	case *corev1.Pod:
		kind = "pod"
		namespace = typed.Namespace
		name = typed.Name
		report = models.HealthReportForPod(*typed)
		filter = filterPod
	case *appsv1.StatefulSet:
		kind = "statefulset"
		namespace = typed.Namespace
		name = typed.Name
		report = models.HealthReportForStatefulSet(*typed)
		filter = filterStatefulSet
	case *extensionsv1beta1.DaemonSet:
		kind = "daemonset"
		namespace = typed.Namespace
		name = typed.Name
		report = models.HealthReportForDaemonSet(*typed)
		filter = filterDaemonSet
	case *extensionsv1beta1.Deployment:
		kind = "deployment"
		namespace = typed.Namespace
		name = typed.Name
		report = models.HealthReportForDeployment(*typed)
		filter = filterDeployment
	//case cache.DeletedFinalStateUnknown: // This is a placeholder until I figure out something better to do in this case.
	default:
		logger.Debug().Interface("object", typed).Msg("Failed to assert type of object.")
		return
	}

	update := models.HealthUpdate{
		Action:    "delete",
		Kind:      kind,
		Namespace: namespace,
		Name:      name,
		Healthy:   report.Healthy,
		Errors:    report.Errors,
	}

	//logger.Debug().Interface("object", obj).Msg("Delete event occurred.")
	HealthUpdates.BroadcastFilter(update.ToMsg(), filter)
	broadcastNamespaceHealth(namespace)
}
