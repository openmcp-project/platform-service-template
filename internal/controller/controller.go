//go:generate opencontrolplane-gen
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

	// opencontrolplane-gen:replace github.com/openmcp-project/platform-service-template=MODULE
	"github.com/openmcp-project/platform-service-template/api/v1alpha1"

	ctrlutils "github.com/openmcp-project/controller-utils/pkg/controller"
	apiconst "github.com/openmcp-project/openmcp-operator/api/constants"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
)

// opencontrolplane-gen:replace Foo=KIND
type FooReconciler struct {
	platformCluster   *clusters.Cluster
	onboardingCluster *clusters.Cluster
	providerName      string
}

// opencontrolplane-gen:replace Foo=KIND
func NewFooReconciler(platformCluster, onboardingCluster *clusters.Cluster, providerName string) *FooReconciler {
	// opencontrolplane-gen:replace Foo=KIND
	return &FooReconciler{
		platformCluster:   platformCluster,
		onboardingCluster: onboardingCluster,
		providerName:      providerName,
	}
}

// opencontrolplane-gen:replace Foo=KIND
func (r *FooReconciler) Reconcile(ctx context.Context, req reconcile.Request) (reconcile.Result, error) {
	log := logf.FromContext(ctx)
	// 1. get obj
	// opencontrolplane-gen:replace Foo=KIND
	obj := &v1alpha1.Foo{}
	// opencontrolplane-gen:if WATCH=onboarding
	if err := r.onboardingCluster.Client().Get(ctx, req.NamespacedName, obj); err != nil {
		return reconcile.Result{}, client.IgnoreNotFound(err)
	}
	// opencontrolplane-gen:fi
	// opencontrolplane-gen:if WATCH=platform
	if err := r.platformCluster.Client().Get(ctx, req.NamespacedName, obj); err != nil {
		return reconcile.Result{}, client.IgnoreNotFound(err)
	}
	// opencontrolplane-gen:fi
	// handle operation annotation
	if obj.GetAnnotations() != nil {
		op, ok := obj.GetAnnotations()[apiconst.OperationAnnotation]
		if ok {
			switch op {
			case apiconst.OperationAnnotationValueIgnore:
				log.Info("Ignoring resource with operation annotation")
				return reconcile.Result{}, nil
			case apiconst.OperationAnnotationValueReconcile:
				// opencontrolplane-gen:if WATCH=onboarding
				if err := ctrlutils.EnsureAnnotation(ctx, r.onboardingCluster.Client(), obj, apiconst.OperationAnnotation, "", true, ctrlutils.DELETE); err != nil {
					return reconcile.Result{}, fmt.Errorf("error removing operation annotation: %w", err)
				}
				// opencontrolplane-gen:fi
				// opencontrolplane-gen:if WATCH=platform
				if err := ctrlutils.EnsureAnnotation(ctx, r.platformCluster.Client(), obj, apiconst.OperationAnnotation, "", true, ctrlutils.DELETE); err != nil {
					return reconcile.Result{}, fmt.Errorf("error removing operation annotation: %w", err)
				}
				// opencontrolplane-gen:fi
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
		return reconcile.Result{}, client.IgnoreNotFound(err)
	}
	// 3. TODO: reconcile obj
	return reconcile.Result{}, nil
}

// opencontrolplane-gen:replace Foo=KIND
func (r *FooReconciler) SetupWithManager(mgr manager.Manager) error {
	secondaryWatchCache := r.platformCluster.Cluster().GetCache()
	// ocp-gen:if WATCH=onboarding
	secondaryWatchCache = r.onboardingCluster.Cluster().GetCache()
	// ocp-gen:fi
	return ctrl.NewControllerManagedBy(mgr).
		// opencontrolplane-gen:replace Foo=KIND
		For(&v1alpha1.Foo{}).
		Owns(&v1alpha1.ProviderConfig{}).
		WatchesRawSource(source.Kind(
			secondaryWatchCache,
			&v1alpha1.ProviderConfig{},
			handler.TypedEnqueueRequestsFromMapFunc(r.enqueueAll()),
			ctrlutils.ToTypedPredicate[*v1alpha1.ProviderConfig](ctrlutils.ExactNamePredicate(r.providerName, "")),
		)).
		Named(r.providerName).
		Complete(r)
}

// opencontrolplane-gen:replace Foo=KIND
// create a reconcile.Request for every existing Foo object on provider config changes.
// opencontrolplane-gen:replace Foo=KIND
func (r *FooReconciler) enqueueAll() func(ctx context.Context, _ *v1alpha1.ProviderConfig) []reconcile.Request {
	return func(ctx context.Context, _ *v1alpha1.ProviderConfig) []reconcile.Request {
		cl := r.platformCluster.Client()
		// ocp-gen:if WATCH=onboarding
		cl = r.onboardingCluster.Client()
		// ocp-gen:fi
		// opencontrolplane-gen:replace Foo=KIND
		list := &v1alpha1.FooList{}
		if err := cl.List(ctx, list); err != nil {
			// opencontrolplane-gen:replace foo=KIND
			logf.FromContext(ctx).Error(err, "failed to list Foo objects")
			return nil
		}
		reqs := make([]reconcile.Request, 0, len(list.Items))
		for _, obj := range list.Items {
			reqs = append(reqs, reconcile.Request{
				NamespacedName: client.ObjectKeyFromObject(&obj),
			})
		}
		return reqs
	}
}
