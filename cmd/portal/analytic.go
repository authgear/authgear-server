package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/authgear/authgear-server/cmd/portal/analytic"
	analyticlib "github.com/authgear/authgear-server/pkg/lib/analytic"
	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/lib/infra/db"
	"github.com/spf13/cobra"
)

func init() {
	binder := getBinder()
	cmdAnalytic.AddCommand(cmdAnalyticReport)
	binder.BindString(cmdAnalyticReport.Flags(), ArgDatabaseURL)
	binder.BindString(cmdAnalyticReport.Flags(), ArgDatabaseSchema)

	binder.BindString(cmdAnalyticReport.Flags(), ArgAnalyticPortalAppID)
	binder.BindInt(cmdAnalyticReport.Flags(), ArgAnalyticYear)
	binder.BindInt(cmdAnalyticReport.Flags(), ArgAnalyticISOWeek)

	binder.BindString(cmdAnalyticReport.Flags(), ArgAnalyticOutputType)
	binder.BindString(cmdAnalyticReport.Flags(), ArgAnalyticCSVOutputFilePath)
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
		csvOutputFilePath := ""
		tokenJSONFilePath := ""
		googleSpreadsheetID := ""
		googleSpreadsheetRange := ""
		reportType := args[0]

		switch outputType {
		case analytic.ReportOutputTypeCSV:
			csvOutputFilePath = binder.GetString(cmd, ArgAnalyticCSVOutputFilePath)
			// if the csv output file path is not provided
			// use the report type as the default file name
			if csvOutputFilePath == "" {
				csvOutputFilePath = fmt.Sprintf("%s.csv", reportType)
			}
		case analytic.ReportOutputTypeGoogleSheets:
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
		default:
			log.Fatalf("unknown output type: %s", outputType)
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

		var data *analyticlib.ReportData
		dbPool := db.NewPool()
		switch reportType {
		case analytic.ReportTypeUserWeeklyReport:
			year, week, err := getISOWeek()
			if err != nil {
				return err
			}
			report := analytic.NewUserWeeklyReport(context.Background(), dbPool, dbCredentials)
			data, err = report.Run(&analyticlib.UserWeeklyReportOptions{
				Year:        year,
				Week:        week,
				PortalAppID: portalAppID,
			})
			if err != nil {
				return err
			}
		case analytic.ReportTypeProjectWeeklyReport:
			year, week, err := getISOWeek()
			if err != nil {
				return err
			}
			report := analytic.NewProjectWeeklyReport(context.Background(), dbPool, dbCredentials)
			data, err = report.Run(&analyticlib.ProjectWeeklyReportOptions{
				Year:        year,
				Week:        week,
				PortalAppID: portalAppID,
			})
			if err != nil {
				return err
			}
		case analytic.ReportTypeProjectMonthlyReport:
			log.Printf("TODO")
		default:
			log.Fatalf("unknown report type: %s", reportType)
		}

		return analytic.OutputReport(
			context.Background(),
			&analytic.OutputReportOptions{
				OutputType:                               outputType,
				CSVOutputFilePath:                        csvOutputFilePath,
				GoogleOAuthClientCredentialsJSONFilePath: clientCredentialsJSONFilePath,
				GoogleOAuthTokenFilePath:                 tokenJSONFilePath,
				SpreadsheetID:                            googleSpreadsheetID,
				SpreadsheetRange:                         googleSpreadsheetRange,
			},
			data,
		)
	},
}
