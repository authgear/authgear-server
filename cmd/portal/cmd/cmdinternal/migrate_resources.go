package cmdinternal

import (
	"github.com/spf13/cobra"

	portalcmd "github.com/authgear/authgear-server/cmd/portal/cmd"
)

var MigrateResourcesDryRun bool

var cmdInternalBreakingChangeMigrateResources = &cobra.Command{
	Use:   "migrate-resources",
	Short: "Migrate resources in database config source",
}

func init() {
	binder := portalcmd.GetBinder()
	binder.BindString(cmdInternalBreakingChangeMigrateResources.PersistentFlags(), portalcmd.ArgDatabaseURL)
	binder.BindString(cmdInternalBreakingChangeMigrateResources.PersistentFlags(), portalcmd.ArgDatabaseSchema)
	cmdInternalBreakingChangeMigrateResources.PersistentFlags().BoolVar(&MigrateResourcesDryRun, "dry-run", true, "Is dry run?")
}
