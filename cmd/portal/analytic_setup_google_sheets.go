package main

import (
	"github.com/spf13/cobra"

	"github.com/authgear/authgear-server/cmd/portal/util/google"
)

func init() {
	binder := getBinder()
	cmdAnalytic.AddCommand(cmdAnalyticSetupGoogleSheetsToken)
	binder.BindString(cmdAnalyticSetupGoogleSheetsToken.Flags(), ArgAnalyticGoogleOAuthClientCredentialsJSONFilePath)
	binder.BindString(cmdAnalyticSetupGoogleSheetsToken.Flags(), ArgAnalyticGoogleOAuthTokenFilePath)
}

var cmdAnalyticSetupGoogleSheetsToken = &cobra.Command{
	Use:   "setup-google-sheets-token",
	Short: "Setup token file for accessing google sheets",
	RunE: func(cmd *cobra.Command, args []string) (err error) {
		binder := getBinder()
		clientCredentialsJSONFilePath, err := binder.GetRequiredString(cmd, ArgAnalyticGoogleOAuthClientCredentialsJSONFilePath)
		if err != nil {
			return err
		}

		tokenJSONFilePath, err := binder.GetRequiredString(cmd, ArgAnalyticGoogleOAuthTokenFilePath)
		if err != nil {
			return err
		}

		oauth2Config, err := google.GetOAuth2Config(
			clientCredentialsJSONFilePath,
			"https://www.googleapis.com/auth/spreadsheets",
		)
		if err != nil {
			return err
		}

		token, err := google.GetTokenFromWeb(oauth2Config)
		if err != nil {
			return err
		}

		err = google.SaveToken(tokenJSONFilePath, token)
		if err != nil {
			return err
		}

		return nil
	},
}
