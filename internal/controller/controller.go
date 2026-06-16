//go:generate ocp-gen
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

	// ocp-gen:replace github.com/openmcp-project/platform-service-template=MODULE
	"github.com/openmcp-project/platform-service-template/api/v1alpha1"

	ctrlutils "github.com/openmcp-project/controller-utils/pkg/controller"
	apiconst "github.com/openmcp-project/openmcp-operator/api/constants"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
)

// ocp-gen:replace Foo=KIND
type FooReconciler struct {
	platformCluster   *clusters.Cluster
	onboardingCluster *clusters.Cluster
	providerName      string
}

// ocp-gen:replace Foo=KIND
func NewFooReconciler(platformCluster, onboardingCluster *clusters.Cluster, providerName string) *FooReconciler {
	// ocp-gen:replace Foo=KIND
	return &FooReconciler{
		platformCluster:   platformCluster,
		onboardingCluster: onboardingCluster,
		providerName:      providerName,
	}
}

// ocp-gen:replace Foo=KIND
func (r *FooReconciler) Reconcile(ctx context.Context, req reconcile.Request) (reconcile.Result, error) {
	log := logf.FromContext(ctx)
	// 1. get obj
	// ocp-gen:replace Foo=KIND
	obj := &v1alpha1.Foo{}
	// ocp-gen:if WATCH=onboarding
	if err := r.onboardingCluster.Client().Get(ctx, req.NamespacedName, obj); err != nil {
		return reconcile.Result{}, client.IgnoreNotFound(err)
	}
	// ocp-gen:fi
	// ocp-gen:if WATCH=platform
	if err := r.platformCluster.Client().Get(ctx, req.NamespacedName, obj); err != nil {
		return reconcile.Result{}, client.IgnoreNotFound(err)
	}
	// ocp-gen:fi
	// handle operation annotation
	if obj.GetAnnotations() != nil {
		op, ok := obj.GetAnnotations()[apiconst.OperationAnnotation]
		if ok {
			switch op {
			case apiconst.OperationAnnotationValueIgnore:
				log.Info("Ignoring resource with operation annotation")
				return reconcile.Result{}, nil
			case apiconst.OperationAnnotationValueReconcile:
				// ocp-gen:if WATCH=onboarding
				if err := ctrlutils.EnsureAnnotation(ctx, r.onboardingCluster.Client(), obj, apiconst.OperationAnnotation, "", true, ctrlutils.DELETE); err != nil {
					return reconcile.Result{}, fmt.Errorf("error removing operation annotation: %w", err)
				}
				// ocp-gen:fi
				// ocp-gen:if WATCH=platform
				if err := ctrlutils.EnsureAnnotation(ctx, r.platformCluster.Client(), obj, apiconst.OperationAnnotation, "", true, ctrlutils.DELETE); err != nil {
					return reconcile.Result{}, fmt.Errorf("error removing operation annotation: %w", err)
				}
				// ocp-gen:fi
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

// ocp-gen:replace Foo=KIND
func (r *FooReconciler) SetupWithManager(mgr manager.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		// ocp-gen:replace Foo=KIND
		For(&v1alpha1.Foo{}).
		// ocp-gen:if WATCH=platform
		Owns(&v1alpha1.ProviderConfig{}).
		// ocp-gen:fi
		// ocp-gen:if WATCH=onboarding
		WatchesRawSource(source.Kind(
			r.platformCluster.Cluster().GetCache(),
			&v1alpha1.ProviderConfig{},
			handler.TypedEnqueueRequestsFromMapFunc(r.enqueueAll()),
			ctrlutils.ToTypedPredicate[*v1alpha1.ProviderConfig](ctrlutils.ExactNamePredicate(r.providerName, "")),
		)).
		// ocp-gen:fi
		Named(r.providerName).
		Complete(r)
}

// ocp-gen:if WATCH=onboarding
// ocp-gen:replace Foo=KIND
// create a reconcile.Request for every existing Foo object on provider config changes.
// ocp-gen:replace Foo=KIND
func (r *FooReconciler) enqueueAll() func(ctx context.Context, _ *v1alpha1.ProviderConfig) []reconcile.Request {
	return func(ctx context.Context, _ *v1alpha1.ProviderConfig) []reconcile.Request {
		// ocp-gen:replace Foo=KIND
		list := &v1alpha1.FooList{}
		if err := r.onboardingCluster.Client().List(ctx, list); err != nil {
			// ocp-gen:replace foo=KIND
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

// ocp-gen:fi
