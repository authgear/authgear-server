package cmdinternal

import (
	"context"
	"fmt"
	"log"

	"github.com/spf13/cobra"

	authgearcmd "github.com/authgear/authgear-server/cmd/authgear/cmd"
	cmdes "github.com/authgear/authgear-server/cmd/authgear/elasticsearch"
	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/lib/infra/db"
)

func init() {
	binder := authgearcmd.GetBinder()

	cmdInternal.AddCommand(cmdInternalElasticsearch)
	binder.BindString(cmdInternalElasticsearch.PersistentFlags(), authgearcmd.ArgElasticsearchURL)
	cmdInternalElasticsearch.AddCommand(cmdInternalElasticsearchCreateIndex)
	cmdInternalElasticsearch.AddCommand(cmdInternalElasticsearchUpdateIndex)
	cmdInternalElasticsearch.AddCommand(cmdInternalElasticsearchDeleteIndex)
	cmdInternalElasticsearch.AddCommand(cmdInternalElasticsearchReindex)

	binder.BindString(cmdInternalElasticsearchReindex.PersistentFlags(), authgearcmd.ArgDatabaseURL)
	binder.BindString(cmdInternalElasticsearchReindex.PersistentFlags(), authgearcmd.ArgDatabaseSchema)
	_ = cmdInternalElasticsearchReindex.Flags().Bool("all", false, "All apps")
}

var cmdInternalElasticsearch = &cobra.Command{
	Use:   "elasticsearch",
	Short: "Elasticsearch commands",
}

var cmdInternalElasticsearchCreateIndex = &cobra.Command{
	Use:   "create-index",
	Short: "Create the search index of all apps",
	RunE: func(cmd *cobra.Command, args []string) error {
		binder := authgearcmd.GetBinder()
		esURL, err := binder.GetRequiredString(cmd, authgearcmd.ArgElasticsearchURL)
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
		binder := authgearcmd.GetBinder()

		esURL, err := binder.GetRequiredString(cmd, authgearcmd.ArgElasticsearchURL)
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

var cmdInternalElasticsearchUpdateIndex = &cobra.Command{
	Use:   "update-index",
	Short: "Update the mappings of search index of all apps",
	RunE: func(cmd *cobra.Command, args []string) error {
		binder := authgearcmd.GetBinder()

		esURL, err := binder.GetRequiredString(cmd, authgearcmd.ArgElasticsearchURL)
		if err != nil {
			return err
		}

		client, err := cmdes.MakeClient(esURL)
		if err != nil {
			return err
		}

		err = cmdes.UpdateIndex(client)
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
		binder := authgearcmd.GetBinder()

		dbURL, err := binder.GetRequiredString(cmd, authgearcmd.ArgDatabaseURL)
		if err != nil {
			return err
		}

		dbSchema, err := binder.GetRequiredString(cmd, authgearcmd.ArgDatabaseSchema)
		if err != nil {
			return err
		}

		esURL, err := binder.GetRequiredString(cmd, authgearcmd.ArgElasticsearchURL)
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
			log.Printf("App (%s): reindexing\n", appID)
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
				err = reindexApp(appID)
				if err != nil {
					return err
				}
			}
			log.Print("Done\n")
			return nil
		}

		appID := args[0]
		err = reindexApp(appID)
		if err != nil {
			return err
		}

		log.Print("Done\n")
		return nil
	},
}
