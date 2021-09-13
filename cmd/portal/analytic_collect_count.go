package main

import (
	"fmt"
	"log"

	"github.com/authgear/authgear-server/cmd/portal/analytic"
	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/spf13/cobra"
)

func init() {
	binder := getBinder()
	cmdAnalytic.AddCommand(cmdAnalyticCollectCount)
	binder.BindString(cmdAnalyticCollectCount.Flags(), ArgDatabaseURL)
	binder.BindString(cmdAnalyticCollectCount.Flags(), ArgDatabaseSchema)
	binder.BindString(cmdAnalyticCollectCount.Flags(), ArgAuditDatabaseURL)
	binder.BindString(cmdAnalyticCollectCount.Flags(), ArgAuditDatabaseSchema)
}

var cmdAnalyticCollectCount = &cobra.Command{
	Use:   "collect-count [interval]",
	Short: "Collect analytic count to the audit db",
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

		auditDBURL, err := binder.GetRequiredString(cmd, ArgAuditDatabaseURL)
		if err != nil {
			return err
		}

		auditDBSchema, err := binder.GetRequiredString(cmd, ArgAuditDatabaseSchema)
		if err != nil {
			return err
		}

		auditDBCredentials := &config.AuditDatabaseCredentials{
			DatabaseURL:    auditDBURL,
			DatabaseSchema: auditDBSchema,
		}

		interval := args[0]
		switch interval {
		case analytic.CollectIntervalTypeDaily:
			fmt.Println(dbCredentials, auditDBCredentials)
			return fmt.Errorf("TODO: collect analytic count daily")
		case analytic.CollectIntervalTypeWeekly:
			return fmt.Errorf("TODO: collect analytic count weekly")
		case analytic.CollectIntervalTypeMonthly:
			return fmt.Errorf("TODO: collect analytic count monthly")
		default:
			log.Fatalf("unknown interval: %s", interval)
		}

		return nil
	},
}
