package app

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/spf13/cobra"
	"sigs.k8s.io/controller-runtime/pkg/client"

	rbacv1 "k8s.io/api/rbac/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"

	crdutil "github.com/openmcp-project/controller-utils/pkg/crds"
	clustersv1alpha1 "github.com/openmcp-project/openmcp-operator/api/clusters/v1alpha1"
	openmcpconst "github.com/openmcp-project/openmcp-operator/api/constants"

	"github.com/openmcp-project/openmcp-operator/lib/clusteraccess"

	"github.com/openmcp-project/platform-service-template/api/crds"
	"github.com/openmcp-project/platform-service-template/api/providerscheme"
	"github.com/openmcp-project/platform-service-template/api/v1alpha1"
)

func NewInitCommand(so *SharedOptions) *cobra.Command {
	opts := &InitOptions{
		SharedOptions: so,
	}
	cmd := &cobra.Command{
		Use:   "init",
		Short: "Initialize the Platform Service",
		RunE: func(cmd *cobra.Command, args []string) error {
			opts.PrintRawOptions(cmd)
			if err := opts.Complete(cmd.Context()); err != nil {
				return fmt.Errorf("error completing options: %w", err)
			}
			opts.PrintCompletedOptions(cmd)
			if opts.DryRun {
				cmd.Println("=== END OF DRY RUN ===")
				return nil
			}
			if err := opts.Run(cmd.Context()); err != nil {
				return err
			}
			return nil
		},
	}
	opts.AddFlags(cmd)

	return cmd
}

type InitOptions struct {
	*SharedOptions
}

func (o *InitOptions) AddFlags(cmd *cobra.Command) {}

func (o *InitOptions) Complete(ctx context.Context) error {
	if err := o.SharedOptions.Complete(); err != nil {
		return err
	}

	return nil
}

func (o *InitOptions) Run(ctx context.Context) error {
	platformScheme := runtime.NewScheme()
	providerscheme.InstallOperatorAPIsPlatform(platformScheme)
	providerscheme.InstallCRDAPIs(platformScheme)
	if err := o.PlatformCluster.InitializeClient(platformScheme); err != nil {
		return err
	}

	log := o.Log.WithName("main")
	log.Info("Environment", "value", o.Environment)
	log.Info("ProviderName", "value", o.ProviderName)

	log.Info("Getting access to the onboarding cluster")
	onboardingScheme := runtime.NewScheme()
	providerscheme.InstallOperatorAPIsOnboarding(onboardingScheme)
	providerscheme.InstallCRDAPIs(onboardingScheme)

	providerSystemNamespace := os.Getenv(openmcpconst.EnvVariablePodNamespace)
	if providerSystemNamespace == "" {
		return fmt.Errorf("environment variable %s is not set", openmcpconst.EnvVariablePodNamespace)
	}

	clusterAccessManager := clusteraccess.NewClusterAccessManager(o.PlatformCluster.Client(), o.ProviderName, providerSystemNamespace)
	clusterAccessManager.WithLogger(&log).
		WithInterval(10 * time.Second).
		WithTimeout(30 * time.Minute)

	onboardingCluster, err := clusterAccessManager.CreateAndWaitForCluster(ctx, clustersv1alpha1.PURPOSE_ONBOARDING+"-init", clustersv1alpha1.PURPOSE_ONBOARDING,
		onboardingScheme, []clustersv1alpha1.PermissionsRequest{
			{
				Rules: []rbacv1.PolicyRule{
					{
						APIGroups: []string{"apiextensions.k8s.io"},
						Resources: []string{"customresourcedefinitions"},
						Verbs:     []string{"*"},
					},
				},
			},
		})

	if err != nil {
		return fmt.Errorf("error creating/updating onboarding cluster: %w", err)
	}

	// apply CRDs
	log.Info("Creating/updating CRDs")
	crdManager := crdutil.NewCRDManager(openmcpconst.ClusterLabel, crds.CRDs)
	crdManager.AddCRDLabelToClusterMapping(clustersv1alpha1.PURPOSE_PLATFORM, o.PlatformCluster)
	crdManager.AddCRDLabelToClusterMapping(clustersv1alpha1.PURPOSE_ONBOARDING, onboardingCluster)
	if err := crdManager.CreateOrUpdateCRDs(ctx, &log); err != nil {
		return fmt.Errorf("error creating/updating CRDs: %w", err)
	}

	log.Info("Lookup ProviderConfig")
	svcCfg := &v1alpha1.ProviderConfig{}
	svcCfg.Name = o.ProviderName
	if err := o.PlatformCluster.Client().Get(ctx, client.ObjectKeyFromObject(svcCfg), svcCfg); err != nil {
		if apierrors.IsNotFound(err) {
			log.Info("Create default ProviderConfig")
			if err := o.PlatformCluster.Client().Create(ctx, svcCfg); err != nil {
				return fmt.Errorf("error creating default ProviderConfig '%s': %w", svcCfg.Name, err)
			}
		}
		return fmt.Errorf("error getting ProviderConfig '%s': %w", svcCfg.Name, err)
	}

	log.Info("Finished init command")
	return nil
}
