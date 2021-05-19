package models

import (
	"fmt"
	"strings"
	"time"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	extensionsv1beta1 "k8s.io/api/extensions/v1beta1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/client-go/informers"
)

type HealthyStatus string

const (
	StatusHealthy   HealthyStatus = "True"
	StatusUnhealthy HealthyStatus = "False"
	StatusUnknown   HealthyStatus = "Unknown"
	StatusWarn      HealthyStatus = "Warn"
)

type HealthReport struct {
	Timestamp   int64         `json:"timestamp"`
	Action      string        `json:"action,omitempty"`
	Kind        string        `json:"kind"`
	Namespace   string        `json:"namespace,omitempty"`
	Name        string        `json:"name"`
	Tenant      string        `json:"tenant,omitempty"`
	Environment string        `json:"environment,omitempty"`
	Healthy     HealthyStatus `json:"healthy"`
	Text        string        `json:"text,omitempty"`
	Errors      []string      `json:"errors,omitempty"`
}

func NewHealthReport() HealthReport {
	return HealthReport{
		Timestamp: time.Now().Unix(),
		Healthy:   StatusUnknown,
	}
}

func HealthReportFor(obj interface{}, factory informers.SharedInformerFactory) (HealthReport, error) {
	switch typed := obj.(type) {
	case *extensionsv1beta1.DaemonSet:
		return HealthReportForDaemonSet(*typed), nil
	case *extensionsv1beta1.Deployment:
		return HealthReportForDeployment(*typed), nil
	case *corev1.Namespace:
		return HealthReportForNamespace(*typed, factory), nil
	case *corev1.Node:
		return HealthReportForNode(*typed), nil
	case *corev1.Pod:
		return HealthReportForPod(*typed), nil
	case *appsv1.StatefulSet:
		return HealthReportForStatefulSet(*typed), nil
	default:
		return NewHealthReport(), fmt.Errorf("Can not generate a report for object: %+v", typed)
	}
}

func HealthReportForDaemonSet(daemonset extensionsv1beta1.DaemonSet) HealthReport {
	report := NewHealthReport()
	report.Kind = daemonset.Kind
	report.Namespace = daemonset.Namespace
	report.Name = daemonset.Name
	report.Tenant, report.Environment = parseTenantAndEnv(daemonset.Namespace)

	if daemonset.Status.DesiredNumberScheduled != daemonset.Status.NumberReady {
		report.Healthy = StatusUnhealthy
		report.Errors = append(report.Errors, fmt.Sprintf("The number of desired pods [%d] does not match the number of ready pods [%d].", daemonset.Status.DesiredNumberScheduled, daemonset.Status.NumberReady))
	}

	if report.Healthy == StatusUnknown {
		report.Healthy = StatusHealthy
	}

	return report
}

func HealthReportForDeployment(deployment extensionsv1beta1.Deployment) HealthReport {
	report := NewHealthReport()
	report.Kind = deployment.Kind
	report.Namespace = deployment.Namespace
	report.Name = deployment.Name
	report.Tenant, report.Environment = parseTenantAndEnv(deployment.Namespace)

	for _, condition := range deployment.Status.Conditions {
		if condition.Type == extensionsv1beta1.DeploymentAvailable && condition.Status == corev1.ConditionFalse {
			report.Healthy = StatusUnhealthy
			report.Errors = append(report.Errors, condition.Message)
		}
	}

	if report.Healthy == StatusUnknown {
		report.Healthy = StatusHealthy
	}

	return report
}

func HealthReportForNamespace(namespace corev1.Namespace, f informers.SharedInformerFactory) HealthReport {
	nsreport := NewHealthReport()
	nsreport.Kind = namespace.Kind
	nsreport.Name = namespace.Name
	nsreport.Tenant, nsreport.Environment = parseTenantAndEnv(namespace.Name)

	// Check DaemonSets
	if daemonsets, err := f.Extensions().V1beta1().DaemonSets().Lister().DaemonSets(namespace.Name).List(labels.Everything()); err == nil {
		unhealthyDaemonSets := make([]string, 0, len(daemonsets))
		for _, daemonset := range daemonsets {
			report := HealthReportForDaemonSet(*daemonset)
			if report.Healthy != StatusHealthy {
				unhealthyDaemonSets = append(unhealthyDaemonSets, daemonset.Name)
			}
		}
		if len(unhealthyDaemonSets) > 0 {
			nsreport.Healthy = StatusUnhealthy
			nsreport.Errors = append(nsreport.Errors, fmt.Sprintf("DaemonSets with unhealthy status: [%s].", strings.Join(unhealthyDaemonSets, ",")))
		}
	} else {
		nsreport.Errors = append(nsreport.Errors, "Failed to fetch DaemonSets from Kubernetes.")
	}

	// Check Deployments
	if deployments, err := f.Extensions().V1beta1().Deployments().Lister().Deployments(namespace.Name).List(labels.Everything()); err == nil {
		unhealthyDeployments := make([]string, 0, len(deployments))
		for _, deployment := range deployments {
			report := HealthReportForDeployment(*deployment)
			if report.Healthy != StatusHealthy {
				unhealthyDeployments = append(unhealthyDeployments, deployment.Name)
			}
		}
		if len(unhealthyDeployments) > 0 {
			nsreport.Healthy = StatusUnhealthy
			nsreport.Errors = append(nsreport.Errors, fmt.Sprintf("Deployments with unhealthy status: [%s].", strings.Join(unhealthyDeployments, ",")))
		}
	} else {
		nsreport.Errors = append(nsreport.Errors, "Failed to fetch Deployments from Kubernetes.")
	}

	// Check Pods
	if pods, err := f.Core().V1().Pods().Lister().Pods(namespace.Name).List(labels.Everything()); err == nil {
		unhealthyPods := make([]string, 0, len(pods))
		for _, pod := range pods {
			report := HealthReportForPod(*pod)
			if report.Healthy != StatusHealthy {
				unhealthyPods = append(unhealthyPods, pod.Name)
			}
		}
		if len(unhealthyPods) > 0 {
			nsreport.Healthy = StatusUnhealthy
			nsreport.Errors = append(nsreport.Errors, fmt.Sprintf("Pods with unhealthy status: [%s].", strings.Join(unhealthyPods, ",")))
		}
	} else {
		nsreport.Errors = append(nsreport.Errors, "Failed to fetch Pods from Kubernetes.")
	}

	// Check Services
	if services, err := f.Core().V1().Services().Lister().Services(namespace.Name).List(labels.Everything()); err == nil {
		unhealthyServices := make([]string, 0, len(services))
		for _, service := range services {
			report := HealthReportForService(*service, f)
			if report.Healthy != StatusHealthy {
				unhealthyServices = append(unhealthyServices, service.Name)
			}
		}
		if len(unhealthyServices) > 0 {
			nsreport.Healthy = StatusUnhealthy
			nsreport.Errors = append(nsreport.Errors, fmt.Sprintf("Services with unhealthy status: [%s].", strings.Join(unhealthyServices, ",")))
		}
	} else {
		nsreport.Errors = append(nsreport.Errors, "Failed to fetch Services from Kubernetes.")
	}

	// Check StatefulSets
	if statefulsets, err := f.Apps().V1().StatefulSets().Lister().StatefulSets(namespace.Name).List(labels.Everything()); err == nil {
		unhealthyStatefulSets := make([]string, 0, len(statefulsets))
		for _, statefulset := range statefulsets {
			report := HealthReportForStatefulSet(*statefulset)
			if report.Healthy != StatusHealthy {
				unhealthyStatefulSets = append(unhealthyStatefulSets, statefulset.Name)
			}
		}
		if len(unhealthyStatefulSets) > 0 {
			nsreport.Healthy = StatusUnhealthy
			nsreport.Errors = append(nsreport.Errors, fmt.Sprintf("StatefulSets with unhealthy status: [%s].", strings.Join(unhealthyStatefulSets, ",")))
		}
	} else {
		nsreport.Errors = append(nsreport.Errors, "Failed to fetch StatefulSets from Kubernetes.")
	}

	// If nobody said we're unhealthy, that must mean we are health, right?
	if nsreport.Healthy == StatusUnknown {
		nsreport.Healthy = StatusHealthy
	}

	return nsreport
}

func HealthReportForNode(node corev1.Node) HealthReport {
	report := NewHealthReport()
	report.Kind = "Node"
	report.Name = node.Name

	for _, condition := range node.Status.Conditions {
		switch condition.Type {
		case corev1.NodeReady:
			if condition.Status != corev1.ConditionTrue {
				report.Healthy = StatusUnhealthy
				report.Errors = append(report.Errors, condition.Message)
			}
		default:
			if condition.Status != corev1.ConditionFalse {
				report.Healthy = StatusUnhealthy
				report.Errors = append(report.Errors, condition.Message)
			}
		}
	}

	if report.Healthy == StatusUnknown {
		report.Healthy = StatusHealthy
	}

	return report
}

func HealthReportForPod(pod corev1.Pod) HealthReport {
	report := NewHealthReport()
	report.Kind = "Pod"
	report.Namespace = pod.Namespace
	report.Name = pod.Name
	report.Tenant, report.Environment = parseTenantAndEnv(pod.Namespace)

	// First check if this pod should be ignored...
	for _, owner := range pod.OwnerReferences {
		if owner.Kind == "Job" {
			// We mark it Healthy so it doesn't mark a NS unhealthy.
			report.Healthy = StatusHealthy
			return report
		}
	}
	for name, value := range pod.Labels {
		if name == "jenkins" && value == "slave" {
			// This pod is part of a jenkins job
			report.Healthy = StatusHealthy
			return report
		}
	}

	if pod.Status.Phase != corev1.PodSucceeded {
		for _, condition := range pod.Status.Conditions {
			if condition.Type == corev1.PodReady && condition.Status == corev1.ConditionFalse {
				report.Healthy = StatusUnhealthy
				report.Errors = append(report.Errors, condition.Message)
			}
		}
	}

	// If nobody said so, we assume we are healthy.
	if report.Healthy == StatusUnknown {
		report.Healthy = StatusHealthy
	}

	return report
}

func HealthReportForService(service corev1.Service, f informers.SharedInformerFactory) HealthReport {
	//   - A service is considered unhealthy if no pods are handling requests
	report := NewHealthReport()
	report.Kind = service.Kind
	report.Namespace = service.Namespace
	report.Name = service.Name
	report.Tenant, report.Environment = parseTenantAndEnv(service.Namespace)

	selector := labels.SelectorFromSet(labels.Set(service.Spec.Selector))
	pods, err := f.Core().V1().Pods().Lister().Pods(service.Namespace).List(selector)
	if err == nil {
		var podReport HealthReport
		for _, pod := range pods {
			podReport = HealthReportForPod(*pod)
			if podReport.Healthy == StatusHealthy {
				// We are good if even a single pod is Ready
				report.Healthy = StatusHealthy
				break
			}
		}
	}

	if report.Healthy == StatusUnknown {
		report.Healthy = StatusHealthy
	}

	return report
}

func HealthReportForStatefulSet(statefulset appsv1.StatefulSet) HealthReport {
	report := NewHealthReport()
	report.Kind = statefulset.Kind
	report.Namespace = statefulset.Namespace
	report.Name = statefulset.Name
	report.Tenant, report.Environment = parseTenantAndEnv(statefulset.Namespace)

	if int32(*statefulset.Spec.Replicas) != int32(statefulset.Status.ReadyReplicas) {
		report.Healthy = StatusUnhealthy
		report.Errors = append(report.Errors, fmt.Sprintf("The number of desired replicas [%d] does not match the number of ready replicas [%d].", statefulset.Spec.Replicas, statefulset.Status.ReadyReplicas))
	}

	if report.Healthy == StatusUnknown {
		report.Healthy = StatusHealthy
	}

	return report
}
