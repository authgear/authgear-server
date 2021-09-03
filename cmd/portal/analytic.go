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

	binder.BindString(cmdAnalyticReport.Flags(), ArgAnalyticPortalAppID)
	binder.BindInt(cmdAnalyticReport.Flags(), ArgAnalyticYear)
	binder.BindInt(cmdAnalyticReport.Flags(), ArgAnalyticISOWeek)

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

		portalAppID, err := binder.GetRequiredString(cmd, ArgAnalyticPortalAppID)
		if err != nil {
			return err
		}

		getISOWeek := func() (year int, week int, err error) {
			y, err := binder.GetInt(cmd, ArgAnalyticYear)
			if err != nil {
				return
			}
			w, err := binder.GetInt(cmd, ArgAnalyticISOWeek)
			if err != nil {
				return
			}
			if y == nil || w == nil {
				// if year and week are not provided
				// use last week as default
				now := time.Now().UTC()
				lastWeek := now.AddDate(0, 0, -7)
				year, week = lastWeek.ISOWeek()
			} else {
				year, week = *y, *w
			}
			return
		}

		dbPool := db.NewPool()
		reportType := args[0]
		switch reportType {
		case reportTypeUserWeeklyReport:
			year, week, err := getISOWeek()
			if err != nil {
				return err
			}
			report := analytic.NewUserWeeklyReport(context.Background(), dbPool, dbCredentials)
			report.Run(&analytic.UserWeeklyReportOptions{
				Year:        year,
				Week:        week,
				PortalAppID: portalAppID,
			})
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
