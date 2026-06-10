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
			if _, err := resources.CreateObjectsFromDir(ctx, c, "config"); err != nil {
				t.Errorf("failed to create platform cluster objects: %v", err)
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
				_, err = resources.CreateObjectsFromDir(ctx, onboardingConfig, "api")
				if err != nil {
					t.Errorf("failed to create platform cluster objects: %v", err)
				}
				return ctx
			}).
		Assess("verify service can be successfully deleted",
			func(ctx context.Context, t *testing.T, c *envconf.Config) context.Context {
				onboardingConfig, err := clusterutils.OnboardingConfig()
				v1alpha1.AddToScheme(onboardingConfig.GetClient().Resources().GetScheme())
				if err != nil {
					t.Error(err)
					return ctx
				}
				fooList := &v1alpha1.FooServiceList{}
				if err := onboardingConfig.GetClient().Resources().List(ctx, fooList); err != nil {
					t.Error(err)
					return ctx
				}
				for _, obj := range fooList.Items {
					if err := resources.DeleteObject(ctx, onboardingConfig, &obj, wait.WithTimeout(time.Minute)); err != nil {
						t.Errorf("failed to delete fooservice object: %v", err)
					}
				}
				return ctx
			})
	testenv.Test(t, basicPlatformServiceTest.Feature())
}
