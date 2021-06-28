package main

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"

	cmdes "github.com/authgear/authgear-server/cmd/authgear/elasticsearch"
	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/lib/infra/db"
)

var cmdInternal = &cobra.Command{
	Use:   "internal",
	Short: "Internal commands",
}

var cmdInternalElasticsearch = &cobra.Command{
	Use:   "elasticsearch",
	Short: "Elasticsearch commands",
}

var cmdInternalElasticsearchCreateIndex = &cobra.Command{
	Use:   "create-index",
	Short: "Create the search index of all apps",
	RunE: func(cmd *cobra.Command, args []string) error {
		binder := getBinder()
		esURL, err := binder.GetRequiredString(cmd, ArgElasticsearchURL)
		if err != nil {
			return err
		}

		client, err := cmdes.MakeClient(esURL)
		if err != nil {
			return err
		}

		err = cmdes.CreateIndex(client)
		if err != nil {
			return err
		}

		return nil
	},
}

var cmdInternalElasticsearchDeleteIndex = &cobra.Command{
	Use:   "delete-index",
	Short: "Delete the search index of all apps",
	RunE: func(cmd *cobra.Command, args []string) error {
		binder := getBinder()

		esURL, err := binder.GetRequiredString(cmd, ArgElasticsearchURL)
		if err != nil {
			return err
		}

		client, err := cmdes.MakeClient(esURL)
		if err != nil {
			return err
		}

		err = cmdes.DeleteIndex(client)
		if err != nil {
			return err
		}

		return nil
	},
}

var cmdInternalElasticsearchReindex = &cobra.Command{
	Use:   "reindex { app-id | --all }",
	Short: "Reindex all documents of a given app into the search index",
	PreRunE: func(cmd *cobra.Command, args []string) error {
		all, err := cmd.Flags().GetBool("all")
		if err == nil && all {
			if len(args) != 0 {
				return fmt.Errorf("no app ID is expected when --all is specified")
			}
		} else {
			if len(args) != 1 {
				return fmt.Errorf("expected exactly 1 argument of app ID")
			}
		}
		return nil
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		binder := getBinder()

		dbURL, err := binder.GetRequiredString(cmd, ArgDatabaseURL)
		if err != nil {
			return err
		}

		dbSchema, err := binder.GetRequiredString(cmd, ArgDatabaseSchema)
		if err != nil {
			return err
		}

		esURL, err := binder.GetRequiredString(cmd, ArgElasticsearchURL)
		if err != nil {
			return err
		}

		client, err := cmdes.MakeClient(esURL)
		if err != nil {
			return err
		}

		dbCredentials := &config.DatabaseCredentials{
			DatabaseURL:    dbURL,
			DatabaseSchema: dbSchema,
		}

		dbPool := db.NewPool()

		reindexApp := func(appID string) error {
			reindexer := cmdes.NewReindexer(context.Background(), dbPool, dbCredentials, config.AppID(appID))

			err = reindexer.Reindex(client)
			if err != nil {
				return err
			}

			return nil
		}

		if all, err := cmd.Flags().GetBool("all"); err == nil && all {
			appLister := cmdes.NewAppLister(context.Background(), dbPool, dbCredentials)
			appIDs, err := appLister.ListApps()
			if err != nil {
				return err
			}
			for _, appID := range appIDs {
				fmt.Printf("Reindexing app (%s)\n", appID)
				err = reindexApp(appID)
				if err != nil {
					return err
				}
			}
			return nil
		}

		appID := args[0]
		return reindexApp(appID)
	},
}

func init() {
	binder := getBinder()

	cmdInternal.AddCommand(cmdInternalElasticsearch)
	binder.BindString(cmdInternalElasticsearch.PersistentFlags(), ArgElasticsearchURL)
	cmdInternalElasticsearch.AddCommand(cmdInternalElasticsearchCreateIndex)
	cmdInternalElasticsearch.AddCommand(cmdInternalElasticsearchDeleteIndex)
	cmdInternalElasticsearch.AddCommand(cmdInternalElasticsearchReindex)

	binder.BindString(cmdInternalElasticsearchReindex.PersistentFlags(), ArgDatabaseURL)
	binder.BindString(cmdInternalElasticsearchReindex.PersistentFlags(), ArgDatabaseSchema)
	_ = cmdInternalElasticsearchReindex.Flags().Bool("all", false, "All apps")
}
