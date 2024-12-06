package cmdpricing

import (
	"context"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"strings"

	"github.com/spf13/cobra"
	"sigs.k8s.io/yaml"

	portalcmd "github.com/authgear/authgear-server/cmd/portal/cmd"
	"github.com/authgear/authgear-server/cmd/portal/plan"
	"github.com/authgear/authgear-server/cmd/portal/util/editor"
	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/lib/config/configsource"
	"github.com/authgear/authgear-server/pkg/lib/infra/db"
	portalconfig "github.com/authgear/authgear-server/pkg/portal/config"
	"github.com/authgear/authgear-server/pkg/util/cobrasentry"
)

func init() {
	binder := portalcmd.GetBinder()
	cmdPricing.AddCommand(cmdPricingPlan)
	cmdPricing.AddCommand(cmdPricingApp)

	cmdPricing.AddCommand(cmdPricingUploadUsageToStripe)
	_ = cmdPricingUploadUsageToStripe.Flags().Bool("all", false, "All apps")
	binder.BindString(cmdPricingUploadUsageToStripe.Flags(), portalcmd.ArgDatabaseURL)
	binder.BindString(cmdPricingUploadUsageToStripe.Flags(), portalcmd.ArgDatabaseSchema)
	binder.BindString(cmdPricingUploadUsageToStripe.Flags(), portalcmd.ArgStripeSecretKey)
	binder.BindString(cmdPricingUploadUsageToStripe.Flags(), cobrasentry.ArgSentryDSN)

	cmdPricingPlan.AddCommand(cmdPricingPlanCreate)
	cmdPricingPlan.AddCommand(cmdPricingPlanUpdate)
	cmdPricingApp.AddCommand(cmdPricingAppSetPlan)
	cmdPricingApp.AddCommand(cmdPricingAppUpdate)

	for _, cmd := range []*cobra.Command{
		cmdPricingPlanCreate,
		cmdPricingPlanUpdate,
		cmdPricingAppSetPlan,
		cmdPricingAppUpdate,
	} {
		binder.BindString(cmd.Flags(), portalcmd.ArgDatabaseURL)
		binder.BindString(cmd.Flags(), portalcmd.ArgDatabaseSchema)
	}

	binder.BindString(cmdPricingPlanUpdate.Flags(), portalcmd.ArgFeatureConfigFilePath)

	binder.BindString(cmdPricingAppSetPlan.Flags(), portalcmd.ArgPlanName)

	binder.BindString(cmdPricingAppUpdate.Flags(), portalcmd.ArgFeatureConfigFilePath)
	binder.BindString(cmdPricingAppUpdate.Flags(), portalcmd.ArgPlanNameForAppUpdate)

	cmdPricing.AddCommand(cmdPricingCreateStripePlans2025)
	binder.BindString(cmdPricingCreateStripePlans2025.Flags(), portalcmd.ArgStripeSecretKey)

	portalcmd.Root.AddCommand(cmdPricing)
}

var cmdPricing = &cobra.Command{
	Use:    "pricing",
	Short:  "Pricing management",
	Hidden: true,
}

var cmdPricingPlan = &cobra.Command{
	Use:   "plan",
	Short: "Plan management",
}

var cmdPricingApp = &cobra.Command{
	Use:   "app",
	Short: "App's plan management",
}

var cmdPricingPlanCreate = &cobra.Command{
	Use:   "create [plan name]",
	Short: "Create plan",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) (err error) {
		binder := portalcmd.GetBinder()
		dbURL, err := binder.GetRequiredString(cmd, portalcmd.ArgDatabaseURL)
		if err != nil {
			return err
		}

		dbSchema, err := binder.GetRequiredString(cmd, portalcmd.ArgDatabaseSchema)
		if err != nil {
			return err
		}

		dbCredentials := &config.DatabaseCredentials{
			DatabaseURL:    dbURL,
			DatabaseSchema: dbSchema,
		}

		dbPool := db.NewPool()
		planService := plan.NewService(dbPool, dbCredentials)
		planName := args[0]
		err = planService.CreatePlan(cmd.Context(), planName)
		return
	},
}

// Example: go run ./cmd/portal pricing plan update free --file=./var/authgear.features.yaml
var cmdPricingPlanUpdate = &cobra.Command{
	Use:   "update [plan name]",
	Short: "Update plan's feature config",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		binder := portalcmd.GetBinder()
		dbURL, err := binder.GetRequiredString(cmd, portalcmd.ArgDatabaseURL)
		if err != nil {
			return err
		}

		dbSchema, err := binder.GetRequiredString(cmd, portalcmd.ArgDatabaseSchema)
		if err != nil {
			return err
		}

		dbCredentials := &config.DatabaseCredentials{
			DatabaseURL:    dbURL,
			DatabaseSchema: dbSchema,
		}

		dbPool := db.NewPool()
		planService := plan.NewService(dbPool, dbCredentials)

		planName := args[0]
		var featureConfigYAML []byte
		featureConfigPath := binder.GetString(cmd, portalcmd.ArgFeatureConfigFilePath)
		if featureConfigPath != "" {
			// update feature config from file
			featureConfigYAML, err = ioutil.ReadFile(featureConfigPath)
			if err != nil {
				return err
			}
		} else {
			// update feature code through editor
			p, err := planService.GetPlan(cmd.Context(), planName)
			if err != nil {
				return err
			}

			edited, err := yaml.Marshal(p.RawFeatureConfig)
			if err != nil {
				return err
			}

			var editError error
			for {
				edited, err = editor.EditYAML(edited, editError, "authgear.features", "yaml")
				if err != nil {
					if errors.Is(editor.ErrEditorCancelled, err) {
						log.Printf("edit cancelled")
					}
					return nil
				}

				_, err = config.ParseFeatureConfig(edited)
				if err != nil {
					editError = err
					continue
				}

				featureConfigYAML = edited
				break
			}
		}

		// update feature config in plan record
		appIDs, err := planService.UpdatePlan(cmd.Context(), planName, featureConfigYAML)
		if err != nil {
			return err
		}

		log.Printf("updated plan, plan: %s", planName)
		log.Printf("apps have been updated: %s", strings.Join(appIDs, ", "))
		return nil
	},
}

// Example: go run ./cmd/portal pricing app set-plan appID --plan-name=free
var cmdPricingAppSetPlan = &cobra.Command{
	Use:   "set-plan [app id]",
	Short: "Set the app to the plan",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) (err error) {
		binder := portalcmd.GetBinder()
		dbURL, err := binder.GetRequiredString(cmd, portalcmd.ArgDatabaseURL)
		if err != nil {
			return err
		}

		dbSchema, err := binder.GetRequiredString(cmd, portalcmd.ArgDatabaseSchema)
		if err != nil {
			return err
		}

		dbCredentials := &config.DatabaseCredentials{
			DatabaseURL:    dbURL,
			DatabaseSchema: dbSchema,
		}

		dbPool := db.NewPool()
		planService := plan.NewService(dbPool, dbCredentials)

		appID := args[0]
		planName, err := binder.GetRequiredString(cmd, portalcmd.ArgPlanName)
		if err != nil {
			return err
		}

		err = planService.UpdateAppPlan(cmd.Context(), appID, planName)
		if err != nil {
			return err
		}

		log.Printf("updated app plan, app: %s, plan: %s\n", appID, planName)

		return
	},
}

// Example: go run ./cmd/portal pricing app update appID --file=./var/authgear.features.yaml
var cmdPricingAppUpdate = &cobra.Command{
	Use:   "update [app id]",
	Short: "Update app's feature config and plan name",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) (err error) {
		binder := portalcmd.GetBinder()
		dbURL, err := binder.GetRequiredString(cmd, portalcmd.ArgDatabaseURL)
		if err != nil {
			return err
		}

		dbSchema, err := binder.GetRequiredString(cmd, portalcmd.ArgDatabaseSchema)
		if err != nil {
			return err
		}

		dbCredentials := &config.DatabaseCredentials{
			DatabaseURL:    dbURL,
			DatabaseSchema: dbSchema,
		}

		dbPool := db.NewPool()
		planService := plan.NewService(dbPool, dbCredentials)

		appID := args[0]
		planName, err := binder.GetRequiredString(cmd, portalcmd.ArgPlanNameForAppUpdate)
		if err != nil {
			return err
		}

		var featureConfigYAML []byte
		featureConfigPath := binder.GetString(cmd, portalcmd.ArgFeatureConfigFilePath)
		if featureConfigPath != "" {
			// update feature config from file
			featureConfigYAML, err = ioutil.ReadFile(featureConfigPath)
			if err != nil {
				return err
			}
		} else {
			// update feature code through editor
			consrc, err := planService.GetDatabaseSourceByAppID(cmd.Context(), appID)
			if err != nil {
				return err
			}

			edited := consrc.Data[configsource.AuthgearFeatureYAML]
			var editError error
			for {
				edited, err = editor.EditYAML(edited, editError, "authgear.features", "yaml")
				if err != nil {
					if errors.Is(editor.ErrEditorCancelled, err) {
						log.Printf("edit cancelled")
					}
					return nil
				}

				_, err = config.ParseFeatureConfig(edited)
				if err != nil {
					editError = err
					continue
				}

				// finish editing
				featureConfigYAML = edited
				break
			}
		}

		err = planService.UpdateAppFeatureConfig(cmd.Context(), appID, featureConfigYAML, planName)
		if err != nil {
			return err
		}

		log.Printf("updated app's feature config, app: %s, plan name: %s\n", appID, planName)
		return nil
	},
}

var cmdPricingUploadUsageToStripe = &cobra.Command{
	Use:   "upload-usage-to-stripe {--all | app-id}",
	Short: "Upload usage to Stripe",
	Args:  cobra.MaximumNArgs(1),
	PreRunE: func(cmd *cobra.Command, args []string) (err error) {
		all, err := cmd.Flags().GetBool("all")
		if err == nil && all {
			if len(args) != 0 {
				err = fmt.Errorf("no app ID is expected when --all is specified")
				return
			}
		} else {
			if len(args) != 1 {
				return fmt.Errorf("expected exactly 1 argument of app ID")
			}
		}
		return
	},
	RunE: cobrasentry.RunEWrap(portalcmd.GetBinder, func(ctx context.Context, cmd *cobra.Command, args []string) (err error) {
		binder := portalcmd.GetBinder()
		dbURL, err := binder.GetRequiredString(cmd, portalcmd.ArgDatabaseURL)
		if err != nil {
			return
		}

		dbSchema, err := binder.GetRequiredString(cmd, portalcmd.ArgDatabaseSchema)
		if err != nil {
			return
		}

		stripeSecretKey, err := binder.GetRequiredString(cmd, portalcmd.ArgStripeSecretKey)
		if err != nil {
			return
		}

		dbCredentials := &config.DatabaseCredentials{
			DatabaseURL:    dbURL,
			DatabaseSchema: dbSchema,
		}

		stripeConfig := &portalconfig.StripeConfig{
			SecretKey: stripeSecretKey,
		}

		hub := cobrasentry.GetHub(ctx)
		dbPool := db.NewPool()
		stripeService := NewStripeService(dbPool, dbCredentials, stripeConfig, hub)

		if len(args) == 0 {
			var errorAppIDs []string
			var appIDs []string
			appIDs, err = stripeService.ListAppIDs(ctx)
			if err != nil {
				return
			}
			for _, appID := range appIDs {
				e := stripeService.UploadUsage(ctx, appID)
				if e != nil {
					errorAppIDs = append(errorAppIDs, appID)
				}
			}
			if len(errorAppIDs) > 0 {
				err = fmt.Errorf("failed to upload usage for %v", errorAppIDs)
			}
		} else {
			appID := args[0]
			err = stripeService.UploadUsage(ctx, appID)
		}

		return
	}),
}
