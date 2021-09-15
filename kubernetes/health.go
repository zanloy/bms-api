package kubernetes // import "github.com/zanloy/bms-api/kubernetes"

import (
	"fmt"

	"github.com/zanloy/bms-api/models"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	extensionsv1beta1 "k8s.io/api/extensions/v1beta1"
	"k8s.io/apimachinery/pkg/labels"
)

func HealthUpdateFor(obj interface{}, action string) (models.HealthUpdate, error) {
	switch typed := obj.(type) {
	case *extensionsv1beta1.DaemonSet:
		ds := models.NewDaemonSet(typed, true)
		return models.HealthUpdate{
			TypeMeta:             ds.TypeMeta,
			Kind:                 "DaemonSet",
			Name:                 ds.Name,
			Namespace:            ds.Namespace,
			TenantInfo:           ds.TenantInfo,
			Action:               action,
			HealthReport:         ds.HealthReport,
			PreviousHealthReport: nil,
		}, nil
	case *extensionsv1beta1.Deployment:
		deployment := models.NewDeployment(typed, true)
		return models.HealthUpdate{
			TypeMeta:             deployment.TypeMeta,
			Kind:                 "Deployment",
			Name:                 deployment.Name,
			Namespace:            deployment.Namespace,
			TenantInfo:           deployment.TenantInfo,
			Action:               action,
			HealthReport:         deployment.HealthReport,
			PreviousHealthReport: nil,
		}, nil
	case *corev1.Namespace:
		ns := models.NewNamespace(typed)
		ns.DaemonSets, _ = GetDaemonSets(ns.Name)
		ns.Deployments, _ = GetDeployments(ns.Name)
		ns.Pods, _ = GetPods(ns.Name)
		ns.Services, _ = GetServices(ns.Name)
		ns.StatefulSets, _ = GetStatefulSets(ns.Name)
		// TODO: Fix this
		schedules, _ := GetVeleroSchedules(ns.Name)
		backups, _ := GetVeleroBackups(ns.Name)
		ns.Velero = models.NamespaceVeleroInfo{
			Schedules: schedules,
			Backups:   backups,
		}
		ns.CheckHealth()
		return models.HealthUpdate{
			TypeMeta:             ns.TypeMeta,
			Kind:                 "Namespace",
			Name:                 ns.Name,
			Namespace:            "",
			TenantInfo:           ns.TenantInfo,
			Action:               action,
			HealthReport:         ns.HealthReport,
			PreviousHealthReport: nil,
		}, nil
	case *corev1.Node:
		node := models.NewNode(typed, true)
		return models.HealthUpdate{
			TypeMeta:             node.TypeMeta,
			Kind:                 "Node",
			Name:                 node.Name,
			Namespace:            node.Namespace,
			Action:               action,
			HealthReport:         node.HealthReport,
			PreviousHealthReport: nil,
		}, nil
	case *corev1.Pod:
		pod := models.NewPod(typed, true)
		return models.HealthUpdate{
			TypeMeta:             pod.TypeMeta,
			Kind:                 "Pod",
			Name:                 pod.Name,
			Namespace:            pod.Namespace,
			TenantInfo:           pod.TenantInfo,
			Action:               action,
			HealthReport:         pod.HealthReport,
			PreviousHealthReport: nil,
		}, nil
	case *corev1.Service:
		selector := labels.SelectorFromSet(labels.Set(typed.Spec.Selector))
		pods, err := GetPodsBySelector(typed.Namespace, selector)
		if err != nil {
			pods = make([]models.Pod, 0)
		}
		svc := models.NewServiceWithPods(typed, pods, true)
		return models.HealthUpdate{
			TypeMeta:             svc.TypeMeta,
			Kind:                 "Service",
			Name:                 svc.Name,
			Namespace:            svc.Namespace,
			TenantInfo:           svc.TenantInfo,
			Action:               action,
			HealthReport:         svc.HealthReport,
			PreviousHealthReport: nil,
		}, nil
	case *appsv1.StatefulSet:
		ss := models.NewStatefulSet(typed, true)
		return models.HealthUpdate{
			TypeMeta:             ss.TypeMeta,
			Kind:                 "StatefulSet",
			Name:                 ss.Name,
			Namespace:            ss.Namespace,
			TenantInfo:           ss.TenantInfo,
			Action:               action,
			HealthReport:         ss.HealthReport,
			PreviousHealthReport: nil,
		}, nil
	default:
		return models.HealthUpdate{}, fmt.Errorf("can not generate a report for object: %+v", typed)
	}
}
