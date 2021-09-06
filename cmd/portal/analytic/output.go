package analytic

import (
	"context"
	"encoding/csv"
	"fmt"
	"log"
	"os"

	"github.com/authgear/authgear-server/cmd/portal/util/google"
	"github.com/authgear/authgear-server/pkg/lib/analytic"
	"google.golang.org/api/sheets/v4"
)

type OutputReportOptions struct {
	OutputType                               string
	CSVOutputFilePath                        string
	GoogleOAuthClientCredentialsJSONFilePath string
	GoogleOAuthTokenFilePath                 string
	SpreadsheetID                            string
	SpreadsheetRange                         string
}

func OutputReport(ctx context.Context, options *OutputReportOptions, data *analytic.ReportData) error {
	switch options.OutputType {
	case ReportOutputTypeCSV:
		f, err := os.Create(options.CSVOutputFilePath)
		defer f.Close()
		if err != nil {
			return fmt.Errorf("Unable to create csv file: %v", err)
		}

		w := csv.NewWriter(f)
		defer w.Flush()
		csvData, err := convertReportDataToCSVRecords(data)
		if err != nil {
			return fmt.Errorf("Unable to convert to csv data: %v", err)
		}

		err = w.WriteAll(csvData)
		if err != nil {
			return fmt.Errorf("Unable to save csv file: %v", err)
		}

		log.Println("Done")
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

func convertReportDataToCSVRecords(data *analytic.ReportData) (records [][]string, err error) {
	toString := func(i interface{}) (string, error) {
		switch i := i.(type) {
		case int:
			return fmt.Sprintf("%d", i), nil
		case string:
			return i, nil
		default:
			// support convert more types if needed
			return "", fmt.Errorf("Unknown data type: %T", i)
		}
	}

	// perpare the header
	row := make([]string, len(data.Header))
	for i, d := range data.Header {
		row[i], err = toString(d)
		if err != nil {
			return
		}
	}
	records = append(records, row)

	// perpare the data
	for _, valRow := range data.Values {
		row := make([]string, len(valRow))
		for i, d := range valRow {
			row[i], err = toString(d)
			if err != nil {
				return
			}
		}
		records = append(records, row)
	}

	return
}
