package cmdsearch

import (
	"context"
	"fmt"
	"log"

	"github.com/spf13/cobra"

	authgearcmd "github.com/authgear/authgear-server/cmd/authgear/cmd"
	cmdpgsearch "github.com/authgear/authgear-server/cmd/authgear/pgsearch"
	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/lib/infra/db"
	"github.com/authgear/authgear-server/pkg/lib/infra/db/searchdb"
	"github.com/authgear/authgear-server/pkg/lib/search/pgsearch"
	authgearlog "github.com/authgear/authgear-server/pkg/util/log"
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

		searchDatabaseCredentials := &config.SearchDatabaseCredentials{
			DatabaseURL:    searchDBURL,
			DatabaseSchema: searchDBSchema,
		}

		dbPool := db.NewPool()
		loggerFactory := authgearlog.NewFactory(authgearlog.LevelInfo)

		reindexApp := func(appID string) error {
			ctx := context.Background()
			log.Printf("App (%s): reindexing\n", appID)
			searchdbHandle := searchdb.NewHandle(
				dbPool,
				config.NewDefaultDatabaseEnvironmentConfig(),
				searchDatabaseCredentials,
				loggerFactory)
			searchdbSQLBuilder := searchdb.NewSQLBuilder(searchDatabaseCredentials)
			searchdbSQLExecutor := searchdb.NewSQLExecutor(searchdbHandle)
			store := pgsearch.NewStore(config.AppID(appID), searchdbSQLBuilder, searchdbSQLExecutor)

			reindexer := cmdpgsearch.NewReindexer(dbPool, dbCredentials, cmdpgsearch.CmdAppID(appID))
			err = searchdbHandle.WithTx(ctx, func(ctx context.Context) error {
				return reindexer.Reindex(ctx, store)
			})
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
