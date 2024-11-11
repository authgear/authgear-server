package cmdanalytic

import (
	"fmt"
	"log"

	"github.com/spf13/cobra"

	"github.com/authgear/authgear-server/cmd/portal/analytic"
	portalcmd "github.com/authgear/authgear-server/cmd/portal/cmd"
	analyticlib "github.com/authgear/authgear-server/pkg/lib/analytic"
	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/lib/infra/db"
	"github.com/authgear/authgear-server/pkg/util/periodical"
)

var cmdAnalyticReport = &cobra.Command{
	Use:   "report [report-type]",
	Short: "Generate analytics report",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) (err error) {
		binder := portalcmd.GetBinder()
		dbURL, err := binder.GetRequiredString(cmd, portalcmd.ArgDatabaseURL)
		if err != nil {
			return err
		}

		dbSchema, err := binder.GetRequiredString(cmd, portalcmd.ArgDatabaseSchema)
		if err != nil {
			return err
		}

		dbCredentials := &config.DatabaseCredentials{
			DatabaseURL:    dbURL,
			DatabaseSchema: dbSchema,
		}

		portalAppID, err := binder.GetRequiredString(cmd, portalcmd.ArgAnalyticPortalAppID)
		if err != nil {
			return err
		}

		outputType, err := binder.GetRequiredString(cmd, portalcmd.ArgAnalyticOutputType)
		if err != nil {
			return err
		}

		period, err := binder.GetRequiredString(cmd, portalcmd.ArgAnalyticPeriod)
		if err != nil {
			return err
		}
		parser := analytic.NewPeriodicalArgumentParser()
		periodicalType, date, err := parser.Parse(period)
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
			csvOutputFilePath = binder.GetString(cmd, portalcmd.ArgAnalyticCSVOutputFilePath)
			// if the csv output file path is not provided
			// use the report type as the default file name
			if csvOutputFilePath == "" {
				csvOutputFilePath = fmt.Sprintf("%s-%s-report.csv", reportType, periodicalType)
			}
		case analytic.ReportOutputTypeGoogleSheets:
			clientCredentialsJSONFilePath, err = binder.GetRequiredString(cmd, portalcmd.ArgAnalyticGoogleOAuthClientCredentialsJSONFilePath)
			if err != nil {
				return err
			}

			tokenJSONFilePath, err = binder.GetRequiredString(cmd, portalcmd.ArgAnalyticGoogleOAuthTokenFilePath)
			if err != nil {
				return err
			}

			googleSpreadsheetID, err = binder.GetRequiredString(cmd, portalcmd.ArgAnalyticGoogleSpreadsheetID)
			if err != nil {
				return err
			}

			googleSpreadsheetRange, err = binder.GetRequiredString(cmd, portalcmd.ArgAnalyticGoogleSpreadsheetRange)
			if err != nil {
				return err
			}
		default:
			log.Fatalf("unknown output type: %s", outputType)
		}

		getAuditDBCredentials := func() (*config.AuditDatabaseCredentials, error) {
			dbURL, err := binder.GetRequiredString(cmd, portalcmd.ArgAuditDatabaseURL)
			if err != nil {
				return nil, err
			}

			dbSchema, err := binder.GetRequiredString(cmd, portalcmd.ArgAuditDatabaseSchema)
			if err != nil {
				return nil, err
			}

			return &config.AuditDatabaseCredentials{
				DatabaseURL:    dbURL,
				DatabaseSchema: dbSchema,
			}, nil
		}

		var mode analyticlib.OutputGoogleSpreadsheetMode
		var data *analyticlib.ReportData
		dbPool := db.NewPool()
		switch reportType {
		case analytic.ReportTypeUser:
			if periodicalType != periodical.Weekly {
				return fmt.Errorf("invalid period, it should be last-week or in the format YYYY-Www")
			}
			year, week := date.ISOWeek()
			report := analytic.NewUserWeeklyReport(dbPool, dbCredentials)
			data, err = report.Run(cmd.Context(), &analyticlib.UserWeeklyReportOptions{
				Year:        year,
				Week:        week,
				PortalAppID: portalAppID,
			})
			if err != nil {
				return err
			}
		case analytic.ReportTypeProject:
			auditDBCredentials, err := getAuditDBCredentials()
			if err != nil {
				return err
			}

			switch periodicalType {
			case periodical.Hourly:
				mode = analyticlib.OutputGoogleSpreadsheetModeOverwrite
				report := analytic.NewProjectHourlyReport(
					dbPool,
					dbCredentials,
					auditDBCredentials,
				)
				data, err = report.Run(cmd.Context(), &analyticlib.ProjectHourlyReportOptions{
					Time: date,
				})
				if err != nil {
					return err
				}
			case periodical.Weekly:
				year, week := date.ISOWeek()
				report := analytic.NewProjectWeeklyReport(
					dbPool,
					dbCredentials,
					auditDBCredentials,
				)
				data, err = report.Run(cmd.Context(), &analyticlib.ProjectWeeklyReportOptions{
					Year:        year,
					Week:        week,
					PortalAppID: portalAppID,
				})
				if err != nil {
					return err
				}
			case periodical.Monthly:
				report := analytic.NewProjectMonthlyReport(
					dbPool,
					dbCredentials,
					auditDBCredentials,
				)
				data, err = report.Run(cmd.Context(), &analyticlib.ProjectMonthlyReportOptions{
					Year:  date.Year(),
					Month: int(date.Month()),
				})
				if err != nil {
					return err
				}
			default:
				return fmt.Errorf("invalid period for project report: %s", period)
			}
		default:
			log.Fatalf("unknown report type: %s", reportType)
		}

		return analytic.OutputReport(
			cmd.Context(),
			&analytic.OutputReportOptions{
				OutputType:                               outputType,
				CSVOutputFilePath:                        csvOutputFilePath,
				SpreadsheetOutputMode:                    mode,
				GoogleOAuthClientCredentialsJSONFilePath: clientCredentialsJSONFilePath,
				GoogleOAuthTokenFilePath:                 tokenJSONFilePath,
				SpreadsheetID:                            googleSpreadsheetID,
				SpreadsheetRange:                         googleSpreadsheetRange,
			},
			data,
		)
	},
}
