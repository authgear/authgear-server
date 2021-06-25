package main

import (
	"context"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/authgear/authgear-server/cmd/portal/plan"
	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/lib/infra/db"
)

func init() {
	cmdPricing.AddCommand(cmdPricingPlan)
	cmdPricingPlan.AddCommand(cmdPricingPlanCreate)

	for _, cmd := range []*cobra.Command{
		cmdPricingPlanCreate,
	} {
		ArgDatabaseURL.Bind(cmd.Flags(), viper.GetViper())
		ArgDatabaseSchema.Bind(cmd.Flags(), viper.GetViper())
	}
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
