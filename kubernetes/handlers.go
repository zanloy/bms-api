package kubernetes

import (
	"fmt"
	"strings"

	"github.com/zanloy/bms-api/models"
	"github.com/zanloy/bms-api/wsrouter"

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

	Factory.Core().V1().
		Services().Informer().AddEventHandler(handlers)

	Factory.Apps().V1().
		StatefulSets().Informer().AddEventHandler(handlers)

	VeleroFactory.Velero().V1().
		Backups().Informer().AddEventHandler(handlers)

	VeleroFactory.Velero().V1().
		Schedules().Informer().AddEventHandler(handlers)
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

func filterService(s *melody.Session) bool {
	return filterKind(s, "service")
}

func filterStatefulSet(s *melody.Session) bool {
	return filterKind(s, "statefulset")
}

func filterURL(s *melody.Session) bool {
	return filterKind(s, "url")
}

func FilterFor(obj interface{}) filterFunc {
	switch obj.(type) {
	case *extensionsv1beta1.DaemonSet:
		return filterDaemonSet
	case *extensionsv1beta1.Deployment:
		return filterDeployment
	case *corev1.Namespace:
		return filterNamespace
	case *corev1.Node:
		return filterNode
	case *corev1.Pod:
		return filterPod
	case *corev1.Service:
		return filterService
	case *appsv1.StatefulSet:
		return filterStatefulSet
	case models.URLCheck:
		return filterURL
	default:
		return func(s *melody.Session) bool { return false } //filterAllowAll
	}
}

func broadcastNamespaceHealth(name string) {
	if name != "" {
		// TODO: Implement this filter (can't now because circle dependency)
		/*
			// See if we should even alert on this ns
			if !config.Config.ShouldNotify("Namespace", name, "") {
				return
			}
		*/

		//cache.WaitForCacheSync(stopCh, Factory.Core().V1().ComponentStatuses().Informer().HasSynced)

		// TODO: Get this value from Config
		//waitduration, _ := time.ParseDuration("5s")
		//time.Sleep(waitduration)

		ns, err := Namespaces().Get(name)
		if err != nil {
			return
		}

		if update, err := HealthUpdateFor(ns, "refresh"); err == nil {
			HealthUpdates.Broadcast(update.ToMsg())
		} else {
			logger.Debug().Msg(fmt.Sprintf("could not find namespace: %s: %v", name, err))
		}
	}
}

// TODO: Find a home
func IsCacheSynced(obj interface{}) bool {
	switch obj.(type) {
	case *extensionsv1beta1.DaemonSet:
		return Factory.Extensions().V1beta1().DaemonSets().Informer().HasSynced()
	case *extensionsv1beta1.Deployment:
		return Factory.Extensions().V1beta1().Deployments().Informer().HasSynced()
	case *corev1.Namespace:
		return Factory.Core().V1().Namespaces().Informer().HasSynced()
	case *corev1.Node:
		return Factory.Core().V1().Nodes().Informer().HasSynced()
	case *corev1.Pod:
		return Factory.Core().V1().Pods().Informer().HasSynced()
	case *corev1.Service:
		return Factory.Core().V1().Services().Informer().HasSynced()
	case *appsv1.StatefulSet:
		return Factory.Apps().V1().StatefulSets().Informer().HasSynced()
	default:
		return false
	}
}

func handleAdd(obj interface{}) {
	var update models.HealthUpdate

	if !IsCacheSynced(obj) {
		return
	}

	update, err := HealthUpdateFor(obj, "add")
	if err != nil {
		logger.Err(err)
		return
	}

	wsrouter.Broadcast(update)

	HealthUpdates.BroadcastFilter(update.ToMsg(), FilterFor(obj))
	broadcastNamespaceHealth(update.Namespace)
}

func handleUpdate(prevObj interface{}, obj interface{}) {
	var (
		prevReport *models.HealthReport
		update     models.HealthUpdate
	)

	if !IsCacheSynced(obj) {
		return
	}

	if prevUpdate, err := HealthUpdateFor(prevObj, "update"); err == nil {
		prevReport = &prevUpdate.HealthReport
	}

	update, err := HealthUpdateFor(obj, "update")
	if err != nil {
		logger.Err(err)
		return
	}

	update.PreviousHealthReport = prevReport

	//logger.Debug().Interface("object", obj).Msg("Update event occurred.")
	if update.Healthy != prevReport.Healthy {
		wsrouter.Broadcast(update)
		HealthUpdates.BroadcastFilter(update.ToMsg(), FilterFor(obj))
		broadcastNamespaceHealth(update.Namespace)
	}
}

func handleDelete(obj interface{}) {
	var update models.HealthUpdate

	if !IsCacheSynced(obj) {
		return
	}

	update, err := HealthUpdateFor(obj, "delete")
	if err != nil {
		logger.Err(err)
		return
	}

	wsrouter.Broadcast(update)

	//logger.Debug().Interface("object", obj).Msg("Delete event occurred.")
	HealthUpdates.BroadcastFilter(update.ToMsg(), FilterFor(obj))
	broadcastNamespaceHealth(update.Namespace)
}
