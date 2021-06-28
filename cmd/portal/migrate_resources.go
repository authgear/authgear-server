package main

import (
	"github.com/spf13/cobra"
)

var MigrateResourcesDryRun bool

var cmdInternalBreakingChangeMigrateResources = &cobra.Command{
	Use:   "migrate-resources",
	Short: "Migrate resources in database config source",
}

func init() {
	binder := getBinder()
	binder.BindString(cmdInternalBreakingChangeMigrateResources.PersistentFlags(), ArgDatabaseURL)
	binder.BindString(cmdInternalBreakingChangeMigrateResources.PersistentFlags(), ArgDatabaseSchema)
	cmdInternalBreakingChangeMigrateResources.PersistentFlags().BoolVar(&MigrateResourcesDryRun, "dry-run", true, "Is dry run?")
}
