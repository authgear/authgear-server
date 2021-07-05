package main

import (
	"context"
	"errors"
	"io/ioutil"
	"log"
	"strings"

	"github.com/spf13/cobra"
	"sigs.k8s.io/yaml"

	"github.com/authgear/authgear-server/cmd/portal/plan"
	"github.com/authgear/authgear-server/cmd/portal/util/editor"
	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/lib/config/configsource"
	"github.com/authgear/authgear-server/pkg/lib/infra/db"
)

func init() {
	binder := getBinder()
	cmdPricing.AddCommand(cmdPricingPlan)
	cmdPricing.AddCommand(cmdPricingApp)
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
		binder.BindString(cmd.Flags(), ArgDatabaseURL)
		binder.BindString(cmd.Flags(), ArgDatabaseSchema)
	}

	binder.BindString(cmdPricingPlanUpdate.Flags(), ArgFeatureConfigFilePath)

	binder.BindString(cmdPricingAppSetPlan.Flags(), ArgPlanName)

	binder.BindString(cmdPricingAppUpdate.Flags(), ArgFeatureConfigFilePath)
	binder.BindString(cmdPricingAppUpdate.Flags(), ArgPlanNameForAppUpdate)
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
		binder := getBinder()
		dbURL, err := binder.GetRequiredString(cmd, ArgDatabaseURL)
		if err != nil {
			return err
		}

		dbSchema, err := binder.GetRequiredString(cmd, ArgDatabaseSchema)
		if err != nil {
			return err
		}

		dbCredentials := &config.DatabaseCredentials{
			DatabaseURL:    dbURL,
			DatabaseSchema: dbSchema,
		}

		dbPool := db.NewPool()
		planService := plan.NewService(context.Background(), dbPool, dbCredentials)
		planName := args[0]
		err = planService.CreatePlan(planName)
		return
	},
}

// Example: go run ./cmd/portal pricing plan update free --file=./var/authgear.features.yaml
var cmdPricingPlanUpdate = &cobra.Command{
	Use:   "update [plan name]",
	Short: "Update plan's feature config",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		binder := getBinder()
		dbURL, err := binder.GetRequiredString(cmd, ArgDatabaseURL)
		if err != nil {
			return err
		}

		dbSchema, err := binder.GetRequiredString(cmd, ArgDatabaseSchema)
		if err != nil {
			return err
		}

		dbCredentials := &config.DatabaseCredentials{
			DatabaseURL:    dbURL,
			DatabaseSchema: dbSchema,
		}

		dbPool := db.NewPool()
		planService := plan.NewService(context.Background(), dbPool, dbCredentials)

		planName := args[0]
		var featureConfigYAML []byte
		featureConfigPath := binder.GetString(cmd, ArgFeatureConfigFilePath)
		if featureConfigPath != "" {
			// update feature config from file
			featureConfigYAML, err = ioutil.ReadFile(featureConfigPath)
			if err != nil {
				return err
			}
		} else {
			// update feature code through editor
			p, err := planService.GetPlan(planName)
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
		appIDs, err := planService.UpdatePlan(planName, featureConfigYAML)
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
		binder := getBinder()
		dbURL, err := binder.GetRequiredString(cmd, ArgDatabaseURL)
		if err != nil {
			return err
		}

		dbSchema, err := binder.GetRequiredString(cmd, ArgDatabaseSchema)
		if err != nil {
			return err
		}

		dbCredentials := &config.DatabaseCredentials{
			DatabaseURL:    dbURL,
			DatabaseSchema: dbSchema,
		}

		dbPool := db.NewPool()
		planService := plan.NewService(context.Background(), dbPool, dbCredentials)

		appID := args[0]
		planName, err := binder.GetRequiredString(cmd, ArgPlanName)
		if err != nil {
			return err
		}

		err = planService.UpdateAppPlan(appID, planName)
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
		binder := getBinder()
		dbURL, err := binder.GetRequiredString(cmd, ArgDatabaseURL)
		if err != nil {
			return err
		}

		dbSchema, err := binder.GetRequiredString(cmd, ArgDatabaseSchema)
		if err != nil {
			return err
		}

		dbCredentials := &config.DatabaseCredentials{
			DatabaseURL:    dbURL,
			DatabaseSchema: dbSchema,
		}

		dbPool := db.NewPool()
		planService := plan.NewService(context.Background(), dbPool, dbCredentials)

		appID := args[0]
		planName, err := binder.GetRequiredString(cmd, ArgPlanNameForAppUpdate)
		if err != nil {
			return err
		}

		var featureConfigYAML []byte
		featureConfigPath := binder.GetString(cmd, ArgFeatureConfigFilePath)
		if featureConfigPath != "" {
			// update feature config from file
			featureConfigYAML, err = ioutil.ReadFile(featureConfigPath)
			if err != nil {
				return err
			}
		} else {
			// update feature code through editor
			consrc, err := planService.GetDatabaseSourceByAppID(appID)
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

		err = planService.UpdateAppFeatureConfig(appID, featureConfigYAML, planName)
		if err != nil {
			return err
		}

		log.Printf("updated app's feature config, app: %s, plan name: %s\n", appID, planName)
		return nil
	},
}
