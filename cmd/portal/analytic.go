package main

import (
	"context"
	"log"
	"time"

	"github.com/authgear/authgear-server/cmd/portal/analytic"
	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/lib/infra/db"
	"github.com/spf13/cobra"
)

const reportTypeUserWeeklyReport = "user-weekly-report"
const reportTypeProjectWeeklyReport = "project-weekly-report"
const reportTypeProjectMonthlyReport = "project-monthly-report"

func init() {
	binder := getBinder()
	cmdAnalytic.AddCommand(cmdAnalyticReport)
	binder.BindString(cmdAnalyticReport.Flags(), ArgDatabaseURL)
	binder.BindString(cmdAnalyticReport.Flags(), ArgDatabaseSchema)
}

var cmdAnalytic = &cobra.Command{
	Use:   "analytic",
	Short: "Analytic report",
}

var cmdAnalyticReport = &cobra.Command{
	Use:   "report [report-type]",
	Short: "Analytic report",
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
		reportType := args[0]
		switch reportType {
		case reportTypeUserWeeklyReport:
			now := time.Now().UTC()
			report := analytic.NewUserWeeklyReport(context.Background(), dbPool, dbCredentials)
			year, week := now.ISOWeek()
			report.Run(year, week)
		case reportTypeProjectWeeklyReport:
			log.Printf("TODO")
		case reportTypeProjectMonthlyReport:
			log.Printf("TODO")
		default:
			log.Fatalf("unknown report type: %s", reportType)
		}

		return
	},
}
