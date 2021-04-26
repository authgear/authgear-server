package main

import (
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var MigrateResourcesDryRun bool

var cmdInternalBreakingChangeMigrateResources = &cobra.Command{
	Use:   "migrate-resources",
	Short: "Migrate resources in database config source",
}

func init() {
	ArgDatabaseURL.Bind(cmdInternalBreakingChangeMigrateResources.PersistentFlags(), viper.GetViper())
	ArgDatabaseSchema.Bind(cmdInternalBreakingChangeMigrateResources.PersistentFlags(), viper.GetViper())
	cmdInternalBreakingChangeMigrateResources.PersistentFlags().BoolVar(&MigrateResourcesDryRun, "dry-run", true, "Is dry run?")
}
