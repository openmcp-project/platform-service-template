package controller

import (
	"context"
	"fmt"

	"github.com/openmcp-project/controller-utils/pkg/clusters"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/source"

	"github.com/openmcp-project/platform-service-template/api/v1alpha1"

	ctrlutils "github.com/openmcp-project/controller-utils/pkg/controller"
	apiconst "github.com/openmcp-project/openmcp-operator/api/constants"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
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
	log := logf.FromContext(ctx)
	// 1. get obj
	obj := &v1alpha1.FooService{}
	if err := r.onboardingCluster.Client().Get(ctx, req.NamespacedName, obj); err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}
	// handle operation annotation
	if obj.GetAnnotations() != nil {
		op, ok := obj.GetAnnotations()[apiconst.OperationAnnotation]
		if ok {
			switch op {
			case apiconst.OperationAnnotationValueIgnore:
				log.Info("Ignoring resource with operation annotation")
				return reconcile.Result{}, nil
			case apiconst.OperationAnnotationValueReconcile:
				if err := ctrlutils.EnsureAnnotation(ctx, r.onboardingCluster.Client(), obj, apiconst.OperationAnnotation, "", true, ctrlutils.DELETE); err != nil {
					return reconcile.Result{}, fmt.Errorf("error removing operation annotation: %w", err)
				}
				log.Info("Manual reconciliation triggered with operation annotation")
			}
		}
	}
	// 2. get config
	config := &v1alpha1.ProviderConfig{}
	if err := r.platformCluster.Client().Get(ctx, types.NamespacedName{Name: r.providerName}, config); err != nil {
		if apierrors.IsNotFound(err) {
			log.Info("No config found", "name", r.providerName)
		}
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}
	// 3. TODO: reconcile obj
	return reconcile.Result{}, nil
}

func (r *FooReconciler) SetupWithManager(mgr manager.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&v1alpha1.FooService{}).
		WatchesRawSource(source.Kind(
			r.platformCluster.Cluster().GetCache(),
			&v1alpha1.ProviderConfig{},
			&handler.TypedEnqueueRequestForObject[*v1alpha1.ProviderConfig]{},
			ctrlutils.ToTypedPredicate[*v1alpha1.ProviderConfig](ctrlutils.ExactNamePredicate(r.providerName, "")),
		)).
		Named(r.providerName).
		Complete(r)
}
