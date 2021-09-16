package kubernetes // import "github.com/zanloy/bms-api/kubernetes"

import (
	"context"
	"fmt"
	"regexp"

	"github.com/zanloy/bms-api/models"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	labels "k8s.io/apimachinery/pkg/labels"
)

// This file contains the funcs to convert k8 resources to bms-api models.

func GetDaemonSet(namespace string, name string) (models.DaemonSet, error) {
	daemonset, err := DaemonSets(namespace).Get(name)
	if err != nil {
		return models.DaemonSet{}, err
	}

	return models.NewDaemonSet(daemonset, true), nil
}

func GetDaemonSets(namespace string) ([]models.DaemonSet, error) {
	results := make([]models.DaemonSet, 0)
	daemonsets, err := DaemonSets(namespace).List(labels.Everything())
	if err != nil {
		return results, err
	}
	for _, daemonset := range daemonsets {
		results = append(results, models.NewDaemonSet(daemonset, true))
	}
	return results, nil
}

func GetAllDaemonSets() ([]models.DaemonSet, error) {
	results := make([]models.DaemonSet, 0)
	daemonsets, err := Extensions().DaemonSets().Lister().List(labels.Everything())
	if err != nil {
		return results, err
	}
	for _, daemonset := range daemonsets {
		results = append(results, models.NewDaemonSet(daemonset, true))
	}
	return results, nil
}

func GetDeployments(namespace string) ([]models.Deployment, error) {
	results := make([]models.Deployment, 0)
	deployments, err := Deployments(namespace).List(labels.Everything())
	if err != nil {
		return results, err
	}
	for _, deployment := range deployments {
		results = append(results, models.NewDeployment(deployment, true))
	}
	return results, nil
}

func GetAllDeployments() ([]models.Deployment, error) {
	results := make([]models.Deployment, 0)
	deployments, err := Extensions().Deployments().Lister().List(labels.Everything())
	if err != nil {
		return results, err
	}
	for _, deployment := range deployments {
		results = append(results, models.NewDeployment(deployment, true))
	}
	return results, nil
}

func GetNamespace(name string) (models.Namespace, error) {
	namespace, err := Namespaces().Get(name)
	if err != nil {
		return models.Namespace{}, err
	}

	ns := models.NewNamespace(namespace)

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

	// TODO: Get bms configmap
	// Setup values from config

	ns.CheckHealth()
	return ns, nil
}

func GetNamespaceWithEvents(name string) (ns models.Namespace, err error) {
	ns, err = GetNamespace(name)
	if err != nil {
		return
	}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	events, err := Clientset.CoreV1().Events(name).List(ctx, metav1.ListOptions{})
	if err != nil {
		return
	}
	ns.Events = events.Items
	return
}

func GetNamespaces() ([]models.Namespace, []error) {
	results := make([]models.Namespace, 0)
	errs := make([]error, 0)
	if namespaces, err := Namespaces().List(labels.Everything()); err == nil {
		for _, namespace := range namespaces {
			if ns, err := GetNamespace(namespace.Name); err == nil {
				results = append(results, ns)
			} else {
				errs = append(errs, err)
			}
		}
	} else {
		errs = append(errs, err)
	}
	return results, errs
}

func GetNode(name string) (models.Node, error) {
	if result, err := Nodes().Get(name); err == nil {
		return models.NewNode(result, true), nil
	} else {
		return models.Node{}, err
	}
}

func GetNodes() ([]models.Node, error) {
	results := make([]models.Node, 0)
	k8nodes, err := Nodes().List(labels.Everything())
	if err != nil {
		return results, err
	}
	for _, k8node := range k8nodes {
		node := models.NewNode(k8node, true)
		if metrics, err := GetNodeMetrics(&node); err == nil {
			node.AddMetrics(metrics)
		}
		results = append(results, node)
	}
	return results, nil
}

func GetPod(namespace string, name string) (models.Pod, error) {
	if pod, err := Pods(namespace).Get(name); err == nil {
		return models.NewPod(pod, true), nil
	} else {
		return models.Pod{}, err
	}
}

func GetPods(namespace string) ([]models.Pod, error) {
	return GetPodsBySelector(namespace, labels.Everything())
}

func GetAllPods() ([]models.Pod, error) {
	results := make([]models.Pod, 0)
	pods, err := Core().Pods().Lister().List(labels.Everything())
	if err != nil {
		return results, err
	}
	for _, pod := range pods {
		results = append(results, models.NewPod(pod, true))
	}
	return results, nil
}

func GetPodsBySelector(namespace string, selector labels.Selector) ([]models.Pod, error) {
	results := make([]models.Pod, 0)
	var pods []*corev1.Pod
	var err error
	if namespace == "" {
		pods, err = Core().Pods().Lister().List(selector)
	} else {
		pods, err = Pods(namespace).List(selector)
	}
	if err != nil {
		return results, err
	}
	for _, pod := range pods {
		results = append(results, models.NewPod(pod, true))
	}
	return results, nil
}

func GetServices(namespace string) ([]models.Service, error) {
	results := make([]models.Service, 0)
	services, err := Services(namespace).List(labels.Everything())
	if err != nil {
		return results, err
	}
	for _, service := range services {
		selector := labels.SelectorFromSet(labels.Set(service.Spec.Selector))
		pods, err := GetPodsBySelector(namespace, selector)
		if err != nil {
			return results, err
		}
		svc := models.NewServiceWithPods(service, pods, true)
		results = append(results, svc)
	}
	return results, nil
}

func GetStatefulSets(namespace string) ([]models.StatefulSet, error) {
	results := make([]models.StatefulSet, 0)
	statefulsets, err := StatefulSets(namespace).List(labels.Everything())
	if err != nil {
		return results, err
	}
	for _, ss := range statefulsets {
		results = append(results, models.NewStatefulSet(ss, true))
	}
	return results, nil
}

func GetAllStatefulSets() ([]models.StatefulSet, error) {
	results := make([]models.StatefulSet, 0)
	statefulsets, err := Apps().StatefulSets().Lister().List(labels.Everything())
	if err != nil {
		return results, err
	}
	for _, statefulset := range statefulsets {
		results = append(results, models.NewStatefulSet(statefulset, true))
	}
	return results, nil
}

func GetVeleroBackups(filter string) ([]models.VeleroBackup, error) {
	results := make([]models.VeleroBackup, 0)
	if filter == "" {
		filter = ".*"
	}

	// Get backups
	if backups, err := VeleroBackups().List(labels.Everything()); err == nil {
		for _, backup := range backups {
			for _, ns := range backup.Spec.IncludedNamespaces {
				// TODO: Actually handle or log the error
				if ok, _ := regexp.Match(filter, []byte(ns)); ok {
					results = append(results, models.NewVeleroBackup(backup, true))
				}
			}
		}
	} else {
		logger.Err(err).Msg(fmt.Sprintf("failed to get velero backups for namespace [%s]", filter))
		return results, err
	}
	return results, nil
}

func GetVeleroSchedules(filter string) ([]models.VeleroSchedule, error) {
	results := make([]models.VeleroSchedule, 0)
	if filter == "" {
		filter = ".*"
	}

	// Get schedules
	if schedules, err := VeleroSchedules().List(labels.Everything()); err == nil {
		for _, schedule := range schedules {
			for _, ns := range schedule.Spec.Template.IncludedNamespaces {
				// TODO: Actually handle or log the error
				if ok, _ := regexp.Match(filter, []byte(ns)); ok {
					results = append(results, models.NewVeleroSchedule(schedule, true))
				}
			}
		}
	} else {
		return results, err
	}

	return results, nil
}
