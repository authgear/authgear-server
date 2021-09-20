package main

import (
	"context"
	"fmt"
	"log"

	"github.com/authgear/authgear-server/cmd/portal/analytic"
	analyticlib "github.com/authgear/authgear-server/pkg/lib/analytic"
	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/lib/infra/db"
	"github.com/authgear/authgear-server/pkg/util/periodical"
	"github.com/spf13/cobra"
)

func init() {
	binder := getBinder()
	cmdAnalytic.AddCommand(cmdAnalyticReport)
	binder.BindString(cmdAnalyticReport.Flags(), ArgDatabaseURL)
	binder.BindString(cmdAnalyticReport.Flags(), ArgDatabaseSchema)
	binder.BindString(cmdAnalyticReport.Flags(), ArgAuditDatabaseURL)
	binder.BindString(cmdAnalyticReport.Flags(), ArgAuditDatabaseSchema)

	binder.BindString(cmdAnalyticReport.Flags(), ArgAnalyticPortalAppID)
	binder.BindString(cmdAnalyticReport.Flags(), ArgAnalyticPeriod)

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

		getAuditDBCredentials := func() (*config.AuditDatabaseCredentials, error) {
			dbURL, err := binder.GetRequiredString(cmd, ArgAuditDatabaseURL)
			if err != nil {
				return nil, err
			}

			dbSchema, err := binder.GetRequiredString(cmd, ArgAuditDatabaseSchema)
			if err != nil {
				return nil, err
			}

			return &config.AuditDatabaseCredentials{
				DatabaseURL:    dbURL,
				DatabaseSchema: dbSchema,
			}, nil
		}

		period, err := binder.GetRequiredString(cmd, ArgAnalyticPeriod)
		if err != nil {
			return err
		}
		parser := analytic.NewPeriodicalArgumentParser()
		periodicalType, date, err := parser.Parse(period)
		if err != nil {
			return err
		}

		var data *analyticlib.ReportData
		dbPool := db.NewPool()
		switch reportType {
		case analytic.ReportTypeUserWeeklyReport:
			if periodicalType != periodical.Weekly {
				return fmt.Errorf("invalid period, it should be last-week or in the format YYYY-Www")
			}
			year, week := date.ISOWeek()
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
			auditDBCredentials, err := getAuditDBCredentials()
			if err != nil {
				return err
			}
			if periodicalType != periodical.Weekly {
				return fmt.Errorf("invalid period, it should be last-week or in the format YYYY-Www")
			}
			year, week := date.ISOWeek()
			report := analytic.NewProjectWeeklyReport(
				context.Background(),
				dbPool,
				dbCredentials,
				auditDBCredentials,
			)
			data, err = report.Run(&analyticlib.ProjectWeeklyReportOptions{
				Year:        year,
				Week:        week,
				PortalAppID: portalAppID,
			})
			if err != nil {
				return err
			}
		case analytic.ReportTypeProjectMonthlyReport:
			auditDBCredentials, err := getAuditDBCredentials()
			if err != nil {
				return err
			}
			if periodicalType != periodical.Monthly {
				return fmt.Errorf("invalid period, it should be last-month or in the format YYYY-MM")
			}
			report := analytic.NewProjectMonthlyReport(
				context.Background(),
				dbPool,
				dbCredentials,
				auditDBCredentials,
			)
			data, err = report.Run(&analyticlib.ProjectMonthlyReportOptions{
				Year:  date.Year(),
				Month: int(date.Month()),
			})
			if err != nil {
				return err
			}
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
