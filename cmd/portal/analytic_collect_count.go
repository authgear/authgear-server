package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/authgear/authgear-server/cmd/portal/analytic"
	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/lib/infra/db"
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

		dbPool := db.NewPool()
		countCollector := analytic.NewCountCollector(context.Background(), dbPool, dbCredentials, auditDBCredentials)

		interval := args[0]
		switch interval {
		case analytic.CollectIntervalTypeDaily:
			log.Println("Start collecting daily analytic count")
			yesterday := time.Now().UTC().AddDate(0, 0, -1)
			updatedCount, err := countCollector.CollectDaily(&yesterday)
			if err != nil {
				return err
			}
			log.Printf("Number of counts have been updated: %d", updatedCount)
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
