package kubernetes // import github.com/zanloy/bms-api/kubernetes

//This file contains all the syntactical sugars for the kubernetes package.

import (
	informerv1 "k8s.io/client-go/informers/core/v1"
	informersv1beta1 "k8s.io/client-go/informers/extensions/v1beta1"
	listerv1 "k8s.io/client-go/listers/core/v1"
	listersv1beta1 "k8s.io/client-go/listers/extensions/v1beta1"
)

// Returns a base core interface.
func Core() informerv1.Interface {
	mustBeInitialized()
	return Factory.Core().V1()
}

// Returns a lister interface for deployments.
func Deployments(namespace string) listersv1beta1.DeploymentNamespaceLister {
	mustBeInitialized()
	return Factory.Extensions().V1beta1().Deployments().Lister().Deployments(namespace)
}

// Returns a base extensions interface.
func Extensions() informersv1beta1.Interface {
	mustBeInitialized()
	return Factory.Extensions().V1beta1()
}

// Return a lister interface for namespaces.
func Namespaces() listerv1.NamespaceLister {
	mustBeInitialized()
	return Factory.Core().V1().Namespaces().Lister()
}

// Return a lister interface for nodes.
func Nodes() listerv1.NodeLister {
	mustBeInitialized()
	return Factory.Core().V1().Nodes().Lister()
}

// Returns a lister interface for pods.
func Pods(namespace string) listerv1.PodNamespaceLister {
	mustBeInitialized()
	return Factory.Core().V1().Pods().Lister().Pods(namespace)
}

// Return a lister interface for services.
func Services(namespace string) listerv1.ServiceNamespaceLister {
	mustBeInitialized()
	return Factory.Core().V1().Services().Lister().Services(namespace)
}
