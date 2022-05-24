package cmdanalytic

import (
	"github.com/spf13/cobra"

	portalcmd "github.com/authgear/authgear-server/cmd/portal/cmd"
	"github.com/authgear/authgear-server/cmd/portal/util/google"
)

var cmdAnalyticSetupGoogleSheetsToken = &cobra.Command{
	Use:   "setup-google-sheets-token",
	Short: "Setup token file for accessing google sheets",
	RunE: func(cmd *cobra.Command, args []string) (err error) {
		binder := portalcmd.GetBinder()
		clientCredentialsJSONFilePath, err := binder.GetRequiredString(cmd, portalcmd.ArgAnalyticGoogleOAuthClientCredentialsJSONFilePath)
		if err != nil {
			return err
		}

		tokenJSONFilePath, err := binder.GetRequiredString(cmd, portalcmd.ArgAnalyticGoogleOAuthTokenFilePath)
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
