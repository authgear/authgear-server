package analytic

import (
	"context"
	"fmt"

	"google.golang.org/api/sheets/v4"

	"github.com/authgear/authgear-server/pkg/util/googleutil"
)

type OutputGoogleSpreadsheet struct {
	GoogleOAuthClientCredentialsJSONFilePath string
	GoogleOAuthTokenFilePath                 string
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

	srv, err := googleutil.GetGoogleSheetsService(ctx, oauth2Config, token)
	if err != nil {
		return err
	}

	vr := &sheets.ValueRange{
		Values: data.Values,
	}

	_, err = srv.Spreadsheets.Values.Append(
		o.SpreadsheetID,
		o.SpreadsheetRange,
		vr,
	).ValueInputOption("RAW").Do()

	if err != nil {
		return fmt.Errorf("unable to update data to sheet: %w", err)
	}

	return nil
}
