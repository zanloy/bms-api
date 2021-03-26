package kubernetes

import (
	"context"
	"os"

	corev1 "k8s.io/api/core/v1"
	extensionsv1beta1 "k8s.io/api/extensions/v1beta1"
	controllers "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var mylogger = controllers.Log.WithName("bms-manager")

func StartManager(ctx context.Context) {
	manager, err := controllers.NewManager(controllers.GetConfigOrDie(), controllers.Options{})
	if err != nil {
		mylogger.Error(err, "unable to create Kubernetes manager")
		os.Exit(1)
	}
	mylogger.Info("kubernetes manager created", "manager", manager)

	err = controllers.
		NewControllerManagedBy(manager).      // Create the controller
		For(&extensionsv1beta1.Deployment{}). // to watch Deployments
		Complete(&DeploymentReconciler{Client: manager.GetClient()})
	if err != nil {
		mylogger.Error(err, "unable to create Deployments controller")
		os.Exit(1)
	}

	if err := manager.Start(ctx); err != nil {
		mylogger.Error(err, "unable to start Kubernetes manager")
		os.Exit(1)
	}
}

type DeploymentReconciler struct {
	client.Client
}

/* Business Logic */
func (dr *DeploymentReconciler) Reconcile(ctx context.Context, req controllers.Request) (controllers.Result, error) {
	deployment := &extensionsv1beta1.Deployment{}
	err := dr.Get(ctx, req.NamespacedName, deployment)
	if err != nil {
		return controllers.Result{}, err
	}

	pods := &corev1.PodList{}
	err = dr.List(ctx, pods, client.InNamespace(req.Namespace), client.MatchingLabels(deployment.Spec.Selector.MatchLabels))
	if err != nil {
		return controllers.Result{}, err
	}

	// DO HEALTH CHECK
	return controllers.Result{}, nil

}
