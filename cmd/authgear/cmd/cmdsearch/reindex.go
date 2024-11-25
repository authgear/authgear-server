package cmdsearch

import (
	"fmt"
	"log"

	"github.com/spf13/cobra"

	authgearcmd "github.com/authgear/authgear-server/cmd/authgear/cmd"
	cmdpgsearch "github.com/authgear/authgear-server/cmd/authgear/pgsearch"
	"github.com/authgear/authgear-server/pkg/lib/infra/db"
)

var cmdSearchReindex = &cobra.Command{
	Use:   "reindex { app-id }",
	Short: "Reindex all documents of a given app into the search index",
	PreRunE: func(cmd *cobra.Command, args []string) error {
		if len(args) < 1 {
			return fmt.Errorf("expected at least 1 app ID")
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

		searchDBURL, err := binder.GetRequiredString(cmd, authgearcmd.ArgSearchDatabaseURL)
		if err != nil {
			return err
		}

		searchDBSchema, err := binder.GetRequiredString(cmd, authgearcmd.ArgSearchDatabaseSchema)
		if err != nil {
			return err
		}

		dbCredentials := &cmdpgsearch.CmdDBCredential{
			DatabaseURL:    dbURL,
			DatabaseSchema: dbSchema,
		}

		searchDatabaseCredentials := &cmdpgsearch.CmdSearchDBCredential{
			DatabaseURL:    searchDBURL,
			DatabaseSchema: searchDBSchema,
		}

		dbPool := db.NewPool()

		reindexApp := func(appID string) error {
			ctx := cmd.Context()
			log.Printf("App (%s): reindexing\n", appID)
			reindexer := cmdpgsearch.NewReindexer(dbPool, dbCredentials, searchDatabaseCredentials, cmdpgsearch.CmdAppID(appID))
			err := reindexer.Reindex(ctx)
			if err != nil {
				return err
			}

			return nil
		}

		for _, appID := range args {
			err = reindexApp(appID)
			if err != nil {
				return err
			}
		}

		log.Println("Done")
		return nil
	},
}
