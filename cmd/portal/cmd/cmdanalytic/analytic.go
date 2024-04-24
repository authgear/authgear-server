package cmdanalytic

import (
	"github.com/spf13/cobra"

	portalcmd "github.com/authgear/authgear-server/cmd/portal/cmd"
)

var cmdAnalytic = &cobra.Command{
	Use:   "analytic",
	Short: "Various commands facilitating product analytics",
}

func init() {
	binder := portalcmd.GetBinder()

	cmdAnalytic.AddCommand(cmdAnalyticReport)
	binder.BindString(cmdAnalyticReport.Flags(), portalcmd.ArgDatabaseURL)
	binder.BindString(cmdAnalyticReport.Flags(), portalcmd.ArgDatabaseSchema)
	binder.BindString(cmdAnalyticReport.Flags(), portalcmd.ArgAuditDatabaseURL)
	binder.BindString(cmdAnalyticReport.Flags(), portalcmd.ArgAuditDatabaseSchema)

	binder.BindString(cmdAnalyticReport.Flags(), portalcmd.ArgAnalyticPortalAppID)
	binder.BindString(cmdAnalyticReport.Flags(), portalcmd.ArgAnalyticPeriod)

	binder.BindString(cmdAnalyticReport.Flags(), portalcmd.ArgAnalyticOutputType)
	binder.BindString(cmdAnalyticReport.Flags(), portalcmd.ArgAnalyticCSVOutputFilePath)
	binder.BindString(cmdAnalyticReport.Flags(), portalcmd.ArgAnalyticGoogleOAuthClientCredentialsJSONFilePath)
	binder.BindString(cmdAnalyticReport.Flags(), portalcmd.ArgAnalyticGoogleOAuthTokenFilePath)
	binder.BindString(cmdAnalyticReport.Flags(), portalcmd.ArgAnalyticGoogleSpreadsheetID)
	binder.BindString(cmdAnalyticReport.Flags(), portalcmd.ArgAnalyticGoogleSpreadsheetRange)

	cmdAnalytic.AddCommand(cmdAnalyticCollectCount)
	binder.BindString(cmdAnalyticCollectCount.Flags(), portalcmd.ArgDatabaseURL)
	binder.BindString(cmdAnalyticCollectCount.Flags(), portalcmd.ArgDatabaseSchema)
	binder.BindString(cmdAnalyticCollectCount.Flags(), portalcmd.ArgAuditDatabaseURL)
	binder.BindString(cmdAnalyticCollectCount.Flags(), portalcmd.ArgAuditDatabaseSchema)
	binder.BindString(cmdAnalyticCollectCount.Flags(), portalcmd.ArgAnalyticRedisURL)

	cmdAnalytic.AddCommand(cmdAnalyticSetupGoogleSheetsToken)
	binder.BindString(cmdAnalyticSetupGoogleSheetsToken.Flags(), portalcmd.ArgAnalyticGoogleOAuthClientCredentialsJSONFilePath)
	binder.BindString(cmdAnalyticSetupGoogleSheetsToken.Flags(), portalcmd.ArgAnalyticGoogleOAuthTokenFilePath)

	cmdAnalytic.AddCommand(cmdAnalyticPosthog)

	cmdAnalyticPosthog.AddCommand(cmdAnalyticPosthogGroup)
	binder.BindString(cmdAnalyticPosthogGroup.Flags(), portalcmd.ArgDatabaseURL)
	binder.BindString(cmdAnalyticPosthogGroup.Flags(), portalcmd.ArgDatabaseSchema)
	binder.BindString(cmdAnalyticPosthogGroup.Flags(), portalcmd.ArgAuditDatabaseURL)
	binder.BindString(cmdAnalyticPosthogGroup.Flags(), portalcmd.ArgAuditDatabaseSchema)
	binder.BindString(cmdAnalyticPosthogGroup.Flags(), portalcmd.ArgAnalyticRedisURL)
	binder.BindString(cmdAnalyticPosthogGroup.Flags(), portalcmd.ArgPosthogEndpoint)
	binder.BindString(cmdAnalyticPosthogGroup.Flags(), portalcmd.ArgPosthogAPIKey)

	cmdAnalyticPosthog.AddCommand(cmdAnalyticPosthogUser)
	binder.BindString(cmdAnalyticPosthogUser.Flags(), portalcmd.ArgDatabaseURL)
	binder.BindString(cmdAnalyticPosthogUser.Flags(), portalcmd.ArgDatabaseSchema)
	binder.BindString(cmdAnalyticPosthogUser.Flags(), portalcmd.ArgAuditDatabaseURL)
	binder.BindString(cmdAnalyticPosthogUser.Flags(), portalcmd.ArgAuditDatabaseSchema)
	binder.BindString(cmdAnalyticPosthogUser.Flags(), portalcmd.ArgAnalyticRedisURL)
	binder.BindString(cmdAnalyticPosthogUser.Flags(), portalcmd.ArgPosthogEndpoint)
	binder.BindString(cmdAnalyticPosthogUser.Flags(), portalcmd.ArgPosthogAPIKey)
	binder.BindString(cmdAnalyticPosthogUser.Flags(), portalcmd.ArgAnalyticPortalAppID)

	portalcmd.Root.AddCommand(cmdAnalytic)
}
