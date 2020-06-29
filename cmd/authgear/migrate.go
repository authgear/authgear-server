package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"strings"

	"github.com/spf13/cobra"

	"github.com/skygeario/skygear-server/cmd/authgear/migrate"
	"github.com/skygeario/skygear-server/pkg/auth/config"
)

var SecretConfigPath string

func init() {
	cmdMigrate.AddCommand(cmdMigrateNew)
	cmdMigrate.AddCommand(cmdMigrateUp)
	cmdMigrate.AddCommand(cmdMigrateDown)
	cmdMigrate.AddCommand(cmdMigrateStatus)

	for _, cmd := range []*cobra.Command{cmdMigrateUp, cmdMigrateDown} {
		cmd.Flags().StringVarP(&SecretConfigPath, "secret-config", "f", "authgear.secrets.yaml", "App secrets YAML path")
	}
}

var cmdMigrate = &cobra.Command{
	Use:   "migrate [up|down]",
	Short: "Migrate database schema",
}

var cmdMigrateNew = &cobra.Command{
	Use:    "new",
	Hidden: true,
	Run: func(cmd *cobra.Command, args []string) {
		name := strings.Join(args, "_")
		migrate.CreateMigration(name)
	},
}

var cmdMigrateUp = &cobra.Command{
	Use:   "up",
	Short: "Migrate database schema to latest version",
	Run: func(cmd *cobra.Command, args []string) {
		credentials, err := loadDBCredentials()
		if err != nil {
			log.Fatalf("cannot load secret config: %s", err)
		}

		migrate.Up(migrate.Options{
			DatabaseURL:    credentials.DatabaseURL,
			DatabaseSchema: credentials.DatabaseSchema,
		})
	},
}

var cmdMigrateDown = &cobra.Command{
	Use:    "down",
	Hidden: true,
	Run: func(cmd *cobra.Command, args []string) {
		credentials, err := loadDBCredentials()
		if err != nil {
			log.Fatalf("cannot load secret config: %s", err)
		}

		migrate.Down(migrate.Options{
			DatabaseURL:    credentials.DatabaseURL,
			DatabaseSchema: credentials.DatabaseSchema,
		})
	},
}

var cmdMigrateStatus = &cobra.Command{
	Use:   "status",
	Short: "Get database schema migration status",
	Run: func(cmd *cobra.Command, args []string) {
		credentials, err := loadDBCredentials()
		if err != nil {
			log.Fatalf("cannot load secret config: %s", err)
		}

		latest := migrate.Status(migrate.Options{
			DatabaseURL:    credentials.DatabaseURL,
			DatabaseSchema: credentials.DatabaseSchema,
		})
		if !latest {
			os.Exit(1)
		}
	},
}

func loadDBCredentials() (*config.DatabaseCredentials, error) {
	yaml, err := ioutil.ReadFile(SecretConfigPath)
	if err != nil {
		return nil, fmt.Errorf("cannot read secret config file: %w", err)
	}

	cfg, err := config.ParseSecret(yaml)
	if err != nil {
		return nil, fmt.Errorf("cannot parse secret config: %w", err)
	}

	credentials := cfg.LookupData(config.DatabaseCredentialsKey).(*config.DatabaseCredentials)
	return credentials, nil
}
