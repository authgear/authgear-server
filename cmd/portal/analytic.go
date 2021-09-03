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

	binder.BindString(cmdAnalyticReport.Flags(), ArgAnalyticOutputType)
	binder.BindString(cmdAnalyticReport.Flags(), ArgAnalyticGoogleOAuthClientCredentialsJSONFilePath)
	binder.BindString(cmdAnalyticReport.Flags(), ArgAnalyticGoogleOAuthTokenFilePath)
	binder.BindString(cmdAnalyticReport.Flags(), ArgAnalyticGoogleSpreadsheetID)
	binder.BindString(cmdAnalyticReport.Flags(), ArgAnalyticGoogleSpreadsheetRange)
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

		outputType, err := binder.GetRequiredString(cmd, ArgAnalyticOutputType)
		if err != nil {
			return err
		}

		clientCredentialsJSONFilePath := ""
		tokenJSONFilePath := ""
		googleSpreadsheetID := ""
		googleSpreadsheetRange := ""
		if outputType == analytic.ReportOutputTypeGoogleSheets {
			clientCredentialsJSONFilePath, err = binder.GetRequiredString(cmd, ArgAnalyticGoogleOAuthClientCredentialsJSONFilePath)
			if err != nil {
				return err
			}

			tokenJSONFilePath, err = binder.GetRequiredString(cmd, ArgAnalyticGoogleOAuthTokenFilePath)
			if err != nil {
				return err
			}

			googleSpreadsheetID, err = binder.GetRequiredString(cmd, ArgAnalyticGoogleSpreadsheetID)
			if err != nil {
				return err
			}

			googleSpreadsheetRange, err = binder.GetRequiredString(cmd, ArgAnalyticGoogleSpreadsheetRange)
			if err != nil {
				return err
			}
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

		var data *analytic.ReportData
		dbPool := db.NewPool()
		reportType := args[0]
		switch reportType {
		case reportTypeUserWeeklyReport:
			year, week, err := getISOWeek()
			if err != nil {
				return err
			}
			report := analytic.NewUserWeeklyReport(context.Background(), dbPool, dbCredentials)
			data, err = report.Run(&analytic.UserWeeklyReportOptions{
				Year:        year,
				Week:        week,
				PortalAppID: portalAppID,
			})
			if err != nil {
				return err
			}
		case reportTypeProjectWeeklyReport:
			log.Printf("TODO")
		case reportTypeProjectMonthlyReport:
			log.Printf("TODO")
		default:
			log.Fatalf("unknown report type: %s", reportType)
		}

		return analytic.OutputReport(
			context.Background(),
			&analytic.OutputReportOptions{
				OutputType:                               outputType,
				GoogleOAuthClientCredentialsJSONFilePath: clientCredentialsJSONFilePath,
				GoogleOAuthTokenFilePath:                 tokenJSONFilePath,
				SpreadsheetID:                            googleSpreadsheetID,
				SpreadsheetRange:                         googleSpreadsheetRange,
			},
			data,
		)
	},
}
