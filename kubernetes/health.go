package kubernetes // import "github.com/zanloy/bms-api/kubernetes"

import (
	"fmt"

	"github.com/zanloy/bms-api/models"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	extensionsv1beta1 "k8s.io/api/extensions/v1beta1"
)

func ReportFor(obj interface{}) (models.HealthReport, error) {
	switch typed := obj.(type) {
	case *extensionsv1beta1.DaemonSet:
		ds := models.NewDaemonSet(typed, true)
		return ds.HealthReport, nil
	case *extensionsv1beta1.Deployment:
		deployment := models.NewDeployment(typed, true)
		return deployment.HealthReport, nil
	case *corev1.Namespace:
		ns := models.NewNamespace(typed)
		ns.DaemonSets, _ = GetDaemonSets(ns.Name)
		ns.Deployments, _ = GetDeployments(ns.Name)
		ns.Pods, _ = GetPods(ns.Name)
		ns.Services, _ = GetServices(ns.Name)
		ns.StatefulSets, _ = GetStatefulSets(ns.Name)

		schedules, _ := GetVeleroSchedules(ns.Name)
		backups, _ := GetVeleroBackups(ns.Name)
		ns.Velero = models.NSVelero{
			Schedules: schedules,
			Backups:   backups,
		}
		ns.CheckHealth()
		return ns.HealthReport, nil
	case *corev1.Node:
		node := models.NewNode(typed, true)
		return node.HealthReport, nil
	case *corev1.Pod:
		pod := models.NewPod(typed, true)
		return pod.HealthReport, nil
	case *appsv1.StatefulSet:
		ss := models.NewStatefulSet(typed, true)
		return ss.HealthReport, nil
	default:
		return models.NewHealthReport(), fmt.Errorf("can not generate a report for object: %+v", typed)
	}
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
	case *appsv1.StatefulSet:
		return filterStatefulSet
	default:
		return filterAllowAll
	}
}
