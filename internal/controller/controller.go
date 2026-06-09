package controller

import (
	"context"

	"github.com/openmcp-project/controller-utils/pkg/clusters"
	"github.com/openmcp-project/controller-utils/pkg/controller"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/source"

	"github.com/openmcp-project/platform-service-template/api/v1alpha1"
)

type FooReconciler struct {
	platformCluster   *clusters.Cluster
	onboardingCluster *clusters.Cluster
	providerName      string
}

func NewFooReconciler(platformCluster, onboardingCluster *clusters.Cluster, providerName string) *FooReconciler {
	return &FooReconciler{
		platformCluster:   platformCluster,
		onboardingCluster: onboardingCluster,
		providerName:      providerName,
	}
}

func (r *FooReconciler) Reconcile(ctx context.Context, req reconcile.Request) (reconcile.Result, error) {
	// 1. get obj
	// 2. get config
	// 3. do something useful
	return reconcile.Result{}, nil
}

func (r *FooReconciler) SetupWithManager(mgr manager.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&v1alpha1.FooService{}).
		WatchesRawSource(source.Kind(
			r.platformCluster.Cluster().GetCache(),
			&v1alpha1.ProviderConfig{},
			&handler.TypedEnqueueRequestForObject[*v1alpha1.ProviderConfig]{},
			controller.ToTypedPredicate[*v1alpha1.ProviderConfig](controller.ExactNamePredicate(r.providerName, "")),
		)).
		Named(r.providerName).
		Complete(r)
}
