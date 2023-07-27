package analytic

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/authgear/authgear-server/pkg/lib/analytic"
)

type OutputReportOptions struct {
	OutputType                               string
	CSVOutputFilePath                        string
	GoogleOAuthClientCredentialsJSONFilePath string
	GoogleOAuthTokenFilePath                 string
	SpreadsheetOutputMode                    analytic.OutputGoogleSpreadsheetMode
	SpreadsheetID                            string
	SpreadsheetRange                         string
}

func OutputReport(ctx context.Context, options *OutputReportOptions, data *analytic.ReportData) error {
	switch options.OutputType {
	case ReportOutputTypeCSV:
		f, err := os.Create(options.CSVOutputFilePath)
		defer f.Close()
		if err != nil {
			return fmt.Errorf("Unable to create csv file: %w", err)
		}

		outputCSV := analytic.OutputCSV{
			Writer: f,
		}

		err = outputCSV.OutputReport(ctx, data)
		if err != nil {
			return fmt.Errorf("failed to output csv: %w", err)
		}

		log.Println("Done")
	case ReportOutputTypeGoogleSheets:
		outputGoogle := analytic.OutputGoogleSpreadsheet{
			GoogleOAuthClientCredentialsJSONFilePath: options.GoogleOAuthClientCredentialsJSONFilePath,
			GoogleOAuthTokenFilePath:                 options.GoogleOAuthTokenFilePath,
			SpreadsheetOutputMode:                    options.SpreadsheetOutputMode,
			SpreadsheetID:                            options.SpreadsheetID,
			SpreadsheetRange:                         options.SpreadsheetRange,
		}

		err := outputGoogle.OutputReport(ctx, data)
		if err != nil {
			return fmt.Errorf("failed to output google spreadsheet: %w", err)
		}

		log.Println("Done")
	default:
		return fmt.Errorf("unsupported output type: %v", options.OutputType)
	}

	return nil
}
