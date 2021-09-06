package analytic

import (
	"context"
	"fmt"
	"log"

	"github.com/authgear/authgear-server/cmd/portal/util/google"
	"github.com/authgear/authgear-server/pkg/lib/analytic"
	"google.golang.org/api/sheets/v4"
)

type OutputReportOptions struct {
	OutputType                               string
	GoogleOAuthClientCredentialsJSONFilePath string
	GoogleOAuthTokenFilePath                 string
	SpreadsheetID                            string
	SpreadsheetRange                         string
}

func OutputReport(ctx context.Context, options *OutputReportOptions, data *analytic.ReportData) error {
	switch options.OutputType {
	case ReportOutputTypeStdout:
		fmt.Println(data.Header)
		for _, entry := range data.Values {
			fmt.Println(entry)
		}
	case ReportOutputTypeGoogleSheets:
		oauth2Config, err := google.GetOAuth2Config(
			options.GoogleOAuthClientCredentialsJSONFilePath,
			"https://www.googleapis.com/auth/spreadsheets",
		)
		if err != nil {
			return err
		}

		token, err := google.GetTokenFromFile(
			options.GoogleOAuthTokenFilePath,
		)
		if err != nil {
			return err
		}

		srv, err := google.GetGoogleSheetsService(ctx, oauth2Config, token)
		if err != nil {
			return err
		}

		vr := &sheets.ValueRange{
			Values: data.Values,
		}

		_, err = srv.Spreadsheets.Values.Append(
			options.SpreadsheetID,
			options.SpreadsheetRange,
			vr,
		).ValueInputOption("RAW").Do()

		if err != nil {
			return fmt.Errorf("Unable to update data to sheet: %v", err)
		}

		log.Println("Done")
	default:
		return fmt.Errorf("unsupported output type: %s", options.OutputType)
	}

	return nil
}
