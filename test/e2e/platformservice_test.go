//go:generate opencontrolplane-gen
package e2e

import (
	"context"
	"testing"
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/e2e-framework/klient/wait"
	"sigs.k8s.io/e2e-framework/pkg/envconf"
	"sigs.k8s.io/e2e-framework/pkg/features"

	// opencontrolplane-gen:if WATCH=onboarding
	"github.com/openmcp-project/openmcp-testing/pkg/clusterutils"
	// opencontrolplane-gen:fi
	"github.com/openmcp-project/openmcp-testing/pkg/resources"

	// opencontrolplane-gen:replace github.com/openmcp-project/platform-service-template=MODULE
	"github.com/openmcp-project/platform-service-template/api/v1alpha1"
)

func TestPlatformService(t *testing.T) {
	basicPlatformServiceTest := features.New("provider test").
		Assess("verify service can be successfully consumed",
			func(ctx context.Context, t *testing.T, c *envconf.Config) context.Context {
				config := c
				// opencontrolplane-gen:if WATCH=onboarding
				config, err := clusterutils.OnboardingConfig()
				if err != nil {
					t.Error(err)
					return ctx
				}
				// opencontrolplane-gen:fi
				v1alpha1.AddToScheme(config.Client().Resources().GetScheme())
				// opencontrolplane-gen:replace Foo=KIND
				api := &v1alpha1.Foo{}
				api.SetName("test")
				api.SetNamespace(metav1.NamespaceDefault)
				if err := config.Client().Resources().Create(ctx, api); err != nil {
					// opencontrolplane-gen:replace Foo=KIND
					t.Errorf("failed to create Foo object: %v", err)
				}
				return ctx
			}).
		Assess("verify service can be successfully deleted",
			func(ctx context.Context, t *testing.T, c *envconf.Config) context.Context {
				config := c
				// opencontrolplane-gen:if WATCH=onboarding
				config, err := clusterutils.OnboardingConfig()
				if err != nil {
					t.Error(err)
					return ctx
				}
				// opencontrolplane-gen:fi
				v1alpha1.AddToScheme(config.Client().Resources().GetScheme())
				// opencontrolplane-gen:replace Foo=KIND
				apiList := &v1alpha1.FooList{}
				if err := config.Client().Resources().List(ctx, apiList); err != nil {
					t.Error(err)
					return ctx
				}
				for _, obj := range apiList.Items {
					if err := resources.DeleteObject(ctx, config, &obj, wait.WithTimeout(time.Minute)); err != nil {
						// opencontrolplane-gen:replace Foo=KIND
						t.Errorf("failed to delete Foo object: %v", err)
					}
				}
				return ctx
			})
	testenv.Test(t, basicPlatformServiceTest.Feature())
}
