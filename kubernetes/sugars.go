package kubernetes // import github.com/zanloy/bms-api/kubernetes

//This file contains all the syntactical sugars for the kubernetes package.

import (
	informersappsv1 "k8s.io/client-go/informers/apps/v1"
	informersv1 "k8s.io/client-go/informers/core/v1"
	informersv1beta1 "k8s.io/client-go/informers/extensions/v1beta1"
	listersappsv1 "k8s.io/client-go/listers/apps/v1"
	listersv1 "k8s.io/client-go/listers/core/v1"
	listersv1beta1 "k8s.io/client-go/listers/extensions/v1beta1"
)

// Returns a base appsv1 interface.
func Apps() informersappsv1.Interface {
	mustBeInitialized()
	return Factory.Apps().V1()
}

// Returns a base core interface.
func Core() informersv1.Interface {
	mustBeInitialized()
	return Factory.Core().V1()
}

// Returns a lister interface for daemonsets.
func DaemonSets(namespace string) listersv1beta1.DaemonSetNamespaceLister {
	mustBeInitialized()
	return Extensions().DaemonSets().Lister().DaemonSets(namespace)
}

// Returns a lister interface for deployments.
func Deployments(namespace string) listersv1beta1.DeploymentNamespaceLister {
	mustBeInitialized()
	return Extensions().Deployments().Lister().Deployments(namespace)
}

// Returns a base extensions interface.
func Extensions() informersv1beta1.Interface {
	mustBeInitialized()
	return Factory.Extensions().V1beta1()
}

// Return a lister interface for namespaces.
func Namespaces() listersv1.NamespaceLister {
	mustBeInitialized()
	return Core().Namespaces().Lister()
}

// Return a lister interface for nodes.
func Nodes() listersv1.NodeLister {
	mustBeInitialized()
	return Core().Nodes().Lister()
}

// Returns a lister interface for pods.
func Pods(namespace string) listersv1.PodNamespaceLister {
	mustBeInitialized()
	return Core().Pods().Lister().Pods(namespace)
}

// Return a lister interface for services.
func Services(namespace string) listersv1.ServiceNamespaceLister {
	mustBeInitialized()
	return Core().Services().Lister().Services(namespace)
}

// Returns a lister interface for statefulsets.
func StatefulSets(namespace string) listersappsv1.StatefulSetNamespaceLister {
	mustBeInitialized()
	return Apps().StatefulSets().Lister().StatefulSets(namespace)
}
