package main

import (
	"context"
	"io/ioutil"
	"log"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/authgear/authgear-server/cmd/portal/plan"
	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/lib/infra/db"
)

func init() {
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
		ArgDatabaseURL.Bind(cmd.Flags(), viper.GetViper())
		ArgDatabaseSchema.Bind(cmd.Flags(), viper.GetViper())
	}

	ArgFeatureConfigFilePath.Bind(cmdPricingPlanUpdate.Flags(), viper.GetViper())
	ArgPlanName.Bind(cmdPricingAppSetPlan.Flags(), viper.GetViper())

	ArgFeatureConfigFilePath.Bind(cmdPricingAppUpdate.Flags(), viper.GetViper())
	ArgPlanNameForAppUpdate.Bind(cmdPricingAppUpdate.Flags(), viper.GetViper())
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
		dbURL, err := ArgDatabaseURL.GetRequired(viper.GetViper())
		if err != nil {
			return err
		}

		dbSchema, err := ArgDatabaseSchema.GetRequired(viper.GetViper())
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
	RunE: func(cmd *cobra.Command, args []string) (err error) {
		dbURL, err := ArgDatabaseURL.GetRequired(viper.GetViper())
		if err != nil {
			return err
		}

		dbSchema, err := ArgDatabaseSchema.GetRequired(viper.GetViper())
		if err != nil {
			return err
		}

		dbCredentials := &config.DatabaseCredentials{
			DatabaseURL:    dbURL,
			DatabaseSchema: dbSchema,
		}

		dbPool := db.NewPool()
		planService := plan.NewService(context.Background(), dbPool, dbCredentials)

		// read the feature config file
		featureConfigPath, err := ArgFeatureConfigFilePath.GetRequired(viper.GetViper())
		if err != nil {
			return err
		}

		featureConfigYAML, err := ioutil.ReadFile(featureConfigPath)
		if err != nil {
			return err
		}

		featureConfig, err := config.ParseFeatureConfig(featureConfigYAML)
		if err != nil {
			return err
		}

		// update feature config in plan record
		planName := args[0]
		appCount, err := planService.UpdatePlan(planName, featureConfig)
		if err != nil {
			return err
		}

		log.Printf("number of apps have been updated: %d", appCount)
		return
	},
}

// Example: go run ./cmd/portal pricing app set-plan appID --plan-name=free
var cmdPricingAppSetPlan = &cobra.Command{
	Use:   "set-plan [app id]",
	Short: "Set the app to the plan",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) (err error) {
		dbURL, err := ArgDatabaseURL.GetRequired(viper.GetViper())
		if err != nil {
			return err
		}

		dbSchema, err := ArgDatabaseSchema.GetRequired(viper.GetViper())
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
		planName, err := ArgPlanName.GetRequired(viper.GetViper())
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
		dbURL, err := ArgDatabaseURL.GetRequired(viper.GetViper())
		if err != nil {
			return err
		}

		dbSchema, err := ArgDatabaseSchema.GetRequired(viper.GetViper())
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
		planName, err := ArgPlanName.GetRequired(viper.GetViper())
		if err != nil {
			return err
		}

		// read the feature config file
		featureConfigPath, err := ArgFeatureConfigFilePath.GetRequired(viper.GetViper())
		if err != nil {
			return err
		}

		featureConfigYAML, err := ioutil.ReadFile(featureConfigPath)
		if err != nil {
			return err
		}

		featureConfig, err := config.ParseFeatureConfig(featureConfigYAML)
		if err != nil {
			return err
		}

		err = planService.UpdateAppFeatureConfig(appID, featureConfig, planName)
		if err != nil {
			return err
		}

		log.Printf("updated app's feature config, app: %s, plan name: %s\n", appID, planName)

		return
	},
}
