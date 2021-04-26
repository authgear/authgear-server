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
	cmdInternalBreakingChangeMigrateResources.PersistentFlags().StringVar(&DatabaseURL, "database-url", "", "Database URL")
	cmdInternalBreakingChangeMigrateResources.PersistentFlags().StringVar(&DatabaseSchema, "database-schema", "", "Database schema name")
	cmdInternalBreakingChangeMigrateResources.PersistentFlags().BoolVar(&MigrateResourcesDryRun, "dry-run", true, "Is dry run?")
}
