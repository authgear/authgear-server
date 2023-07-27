package analytic

import (
	"context"
	"fmt"

	"google.golang.org/api/sheets/v4"

	"github.com/authgear/authgear-server/pkg/util/googleutil"
)

type OutputGoogleSpreadsheetMode string

const (
	OutputGoogleSpreadsheetModeDefault   OutputGoogleSpreadsheetMode = ""
	OutputGoogleSpreadsheetModeAppend    OutputGoogleSpreadsheetMode = "append"
	OutputGoogleSpreadsheetModeOverwrite OutputGoogleSpreadsheetMode = "overwrite"
)

type OutputGoogleSpreadsheet struct {
	GoogleOAuthClientCredentialsJSONFilePath string
	GoogleOAuthTokenFilePath                 string
	SpreadsheetOutputMode                    OutputGoogleSpreadsheetMode
	SpreadsheetID                            string
	SpreadsheetRange                         string
}

func (o *OutputGoogleSpreadsheet) OutputReport(ctx context.Context, data *ReportData) error {
	oauth2Config, err := googleutil.GetOAuth2Config(
		o.GoogleOAuthClientCredentialsJSONFilePath,
		"https://www.googleapis.com/auth/spreadsheets",
	)
	if err != nil {
		return err
	}

	token, err := googleutil.GetTokenFromFile(
		o.GoogleOAuthTokenFilePath,
	)
	if err != nil {
		return err
	}

	svc, err := googleutil.GetGoogleSheetsService(ctx, oauth2Config, token)
	if err != nil {
		return err
	}

	vr := &sheets.ValueRange{
		Values: data.Values,
	}

	err = o.output(svc, vr)

	if err != nil {
		return fmt.Errorf("unable to update data to sheet: %w", err)
	}

	return nil
}

func (o *OutputGoogleSpreadsheet) output(svc *sheets.Service, vr *sheets.ValueRange) error {
	mode := o.SpreadsheetOutputMode
	if mode == OutputGoogleSpreadsheetModeDefault {
		mode = OutputGoogleSpreadsheetModeAppend
	}

	switch mode {
	case OutputGoogleSpreadsheetModeAppend:
		_, err := svc.Spreadsheets.Values.Append(
			o.SpreadsheetID,
			o.SpreadsheetRange,
			vr,
		).ValueInputOption("RAW").Do()
		if err != nil {
			return err
		}
		return nil
	case OutputGoogleSpreadsheetModeOverwrite:
		_, err := svc.Spreadsheets.Values.Clear(
			o.SpreadsheetID,
			// TODO: we assume there are no more than Z columns. It should be good for now.
			fmt.Sprintf("%v:Z", o.SpreadsheetRange),
			&sheets.ClearValuesRequest{},
		).Do()
		if err != nil {
			return err
		}

		_, err = svc.Spreadsheets.Values.Append(
			o.SpreadsheetID,
			o.SpreadsheetRange,
			vr,
		).ValueInputOption("RAW").Do()
		if err != nil {
			return err
		}
		return nil
	default:
		return fmt.Errorf("unknown mode: %v", mode)
	}
}
