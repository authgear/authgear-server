package main

import (
	"context"
	"fmt"
	"log"

	"github.com/authgear/authgear-server/cmd/portal/analytic"
	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/lib/infra/db"
	"github.com/authgear/authgear-server/pkg/lib/infra/redis"
	"github.com/authgear/authgear-server/pkg/util/periodical"
	"github.com/authgear/authgear-server/pkg/util/timeutil"
	"github.com/spf13/cobra"
)

func init() {
	binder := getBinder()
	cmdAnalytic.AddCommand(cmdAnalyticCollectCount)
	binder.BindString(cmdAnalyticCollectCount.Flags(), ArgDatabaseURL)
	binder.BindString(cmdAnalyticCollectCount.Flags(), ArgDatabaseSchema)
	binder.BindString(cmdAnalyticCollectCount.Flags(), ArgAuditDatabaseURL)
	binder.BindString(cmdAnalyticCollectCount.Flags(), ArgAuditDatabaseSchema)
	binder.BindString(cmdAnalyticCollectCount.Flags(), ArgAnalyticRedisURL)
}

var cmdAnalyticCollectCount = &cobra.Command{
	Use:   "collect-count [period]",
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

		var analyticRedisCredentials *config.AnalyticRedisCredentials
		analyticRedisURL := binder.GetString(cmd, ArgAnalyticRedisURL)
		if analyticRedisURL != "" {
			analyticRedisCredentials = &config.AnalyticRedisCredentials{
				RedisURL: analyticRedisURL,
			}
		}

		dbPool := db.NewPool()
		redisPool := redis.NewPool()
		countCollector := analytic.NewCountCollector(
			context.Background(),
			dbPool,
			dbCredentials,
			auditDBCredentials,
			redisPool,
			analyticRedisCredentials,
		)

		period := args[0]
		parser := analytic.NewPeriodicalArgumentParser()
		periodicalType, date, err := parser.Parse(period)
		if err != nil {
			return err
		}
		switch periodicalType {
		case periodical.Daily:
			log.Println("Start collecting daily analytic count", date.Format(timeutil.LayoutISODate))
			updatedCount, err := countCollector.CollectDaily(date)
			if err != nil {
				return err
			}
			log.Printf("Number of counts have been updated: %d", updatedCount)
		case periodical.Weekly:
			year, week := date.ISOWeek()
			log.Println(
				"Start collecting weekly analytic count",
				date.Format(timeutil.LayoutISODate),
				fmt.Sprintf("%04d-W%02d", year, week),
			)
			updatedCount, err := countCollector.CollectWeekly(date)
			if err != nil {
				return err
			}
			log.Printf("Number of counts have been updated: %d", updatedCount)
		case periodical.Monthly:
			log.Println(
				"Start collecting monthly analytic count",
				date.Format("2006-01"),
			)
			updatedCount, err := countCollector.CollectMonthly(date)
			if err != nil {
				return err
			}
			log.Printf("Number of counts have been updated: %d", updatedCount)
		}

		return nil
	},
}
