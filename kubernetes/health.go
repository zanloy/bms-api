package kubernetes // import "github.com/zanloy/bms-api/kubernetes"

import (
	"fmt"

	"github.com/zanloy/bms-api/models"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	extensionsv1beta1 "k8s.io/api/extensions/v1beta1"
)

func HealthUpdateFor(obj interface{}, action string) (models.HealthUpdate, error) {
	switch typed := obj.(type) {
	case *extensionsv1beta1.DaemonSet:
		ds := models.NewDaemonSet(typed, true)
		return models.HealthUpdate{
			TypeMeta:             typed.TypeMeta,
			Kind:                 typed.Kind,
			Name:                 typed.Name,
			Namespace:            typed.Namespace,
			HealthReport:         ds.HealthReport,
			Action:               action,
			PreviousHealthReport: nil,
		}, nil
	case *extensionsv1beta1.Deployment:
		deployment := models.NewDeployment(typed, true)
		return models.HealthUpdate{
			TypeMeta:             typed.TypeMeta,
			Kind:                 typed.Kind,
			Name:                 typed.Name,
			Namespace:            typed.Namespace,
			HealthReport:         deployment.HealthReport,
			Action:               action,
			PreviousHealthReport: nil,
		}, nil
	case *corev1.Namespace:
		ns := models.NewNamespace(typed)
		ns.DaemonSets, _ = GetDaemonSets(ns.Name)
		ns.Deployments, _ = GetDeployments(ns.Name)
		ns.Pods, _ = GetPods(ns.Name)
		ns.Services, _ = GetServices(ns.Name)
		ns.StatefulSets, _ = GetStatefulSets(ns.Name)
		schedules, _ := GetVeleroSchedules(ns.Name)
		backups, _ := GetVeleroBackups(ns.Name)
		ns.Velero = models.NamespaceVeleroInfo{
			Schedules: schedules,
			Backups:   backups,
		}
		ns.CheckHealth()
		return models.HealthUpdate{
			TypeMeta:             typed.TypeMeta,
			Kind:                 typed.Kind,
			Name:                 typed.Name,
			Namespace:            typed.Namespace,
			HealthReport:         ns.HealthReport,
			Action:               action,
			PreviousHealthReport: nil,
		}, nil
	case *corev1.Node:
		node := models.NewNode(typed, true)
		return models.HealthUpdate{
			TypeMeta:             typed.TypeMeta,
			Kind:                 typed.Kind,
			Name:                 typed.Name,
			Namespace:            typed.Namespace,
			HealthReport:         node.HealthReport,
			Action:               action,
			PreviousHealthReport: nil,
		}, nil
	case *corev1.Pod:
		pod := models.NewPod(typed, true)
		return models.HealthUpdate{
			TypeMeta:             typed.TypeMeta,
			Kind:                 typed.Kind,
			Name:                 typed.Name,
			Namespace:            typed.Namespace,
			HealthReport:         pod.HealthReport,
			Action:               action,
			PreviousHealthReport: nil,
		}, nil
	case *appsv1.StatefulSet:
		ss := models.NewStatefulSet(typed, true)
		return models.HealthUpdate{
			TypeMeta:             typed.TypeMeta,
			Kind:                 typed.Kind,
			Name:                 typed.Name,
			Namespace:            typed.Namespace,
			HealthReport:         ss.HealthReport,
			Action:               action,
			PreviousHealthReport: nil,
		}, nil
	default:
		return models.HealthUpdate{}, fmt.Errorf("can not generate a report for object: %+v", typed)
	}
}
