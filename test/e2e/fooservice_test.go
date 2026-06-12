package e2e

import (
	"context"
	"testing"
	"time"

	"sigs.k8s.io/e2e-framework/klient/wait"
	"sigs.k8s.io/e2e-framework/pkg/envconf"
	"sigs.k8s.io/e2e-framework/pkg/features"

	"github.com/openmcp-project/openmcp-testing/pkg/clusterutils"
	"github.com/openmcp-project/openmcp-testing/pkg/resources"
	"github.com/openmcp-project/platform-service-template/api/v1alpha1"
)

func TestPlatformService(t *testing.T) {
	basicPlatformServiceTest := features.New("provider test").
		Setup(func(ctx context.Context, t *testing.T, c *envconf.Config) context.Context {
			// create config on platform cluster
			v1alpha1.AddToScheme(c.GetClient().Resources().GetScheme())
			config := &v1alpha1.ProviderConfig{}
			if err := c.GetClient().Resources().Create(ctx, config); err != nil {
				t.Errorf("(platform cluster) failed to create provider config: %v", err)
			}
			return ctx
		}).
		Assess("verify service can be successfully consumed",
			func(ctx context.Context, t *testing.T, c *envconf.Config) context.Context {
				onboardingConfig, err := clusterutils.OnboardingConfig()
				if err != nil {
					t.Error(err)
					return ctx
				}
				v1alpha1.AddToScheme(onboardingConfig.GetClient().Resources().GetScheme())
				api := &v1alpha1.Foo{}
				if err := c.GetClient().Resources().Create(ctx, api); err != nil {
					t.Errorf("(onboarding cluster) failed to create foo object: %v", err)
				}
				return ctx
			}).
		Assess("verify service can be successfully deleted",
			func(ctx context.Context, t *testing.T, c *envconf.Config) context.Context {
				onboardingConfig, err := clusterutils.OnboardingConfig()
				if err != nil {
					t.Error(err)
					return ctx
				}
				v1alpha1.AddToScheme(onboardingConfig.GetClient().Resources().GetScheme())
				apiList := &v1alpha1.FooList{}
				if err := onboardingConfig.GetClient().Resources().List(ctx, apiList); err != nil {
					t.Error(err)
					return ctx
				}
				for _, obj := range apiList.Items {
					if err := resources.DeleteObject(ctx, onboardingConfig, &obj, wait.WithTimeout(time.Minute)); err != nil {
						t.Errorf("(onboarding cluster) failed to delete foo object: %v", err)
					}
				}
				return ctx
			})
	testenv.Test(t, basicPlatformServiceTest.Feature())
}
