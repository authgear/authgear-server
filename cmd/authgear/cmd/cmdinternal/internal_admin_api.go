package cmdinternal

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/authgear/authgear-server/cmd/authgear/adminapi"
	authgearcmd "github.com/authgear/authgear-server/cmd/authgear/cmd"
	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/lib/infra/db"
)

func init() {
	binder := authgearcmd.GetBinder()

	cmdInternal.AddCommand(cmdInternalAdminAPI)

	cmdInternalAdminAPI.AddCommand(cmdInternalAdminAPIInvoke)

	binder.BindString(cmdInternalAdminAPIInvoke.Flags(), authgearcmd.ArgDatabaseURL)
	binder.BindString(cmdInternalAdminAPIInvoke.Flags(), authgearcmd.ArgDatabaseSchema)

	_ = cmdInternalAdminAPIInvoke.Flags().String("app-id", "", "The target app ID")
	_ = cmdInternalAdminAPIInvoke.MarkFlagRequired("app-id")

	_ = cmdInternalAdminAPIInvoke.Flags().String("endpoint", "", "The endpoint to the Admin API server, excluding the path")
	_ = cmdInternalAdminAPIInvoke.MarkFlagRequired("endpoint")
	_ = cmdInternalAdminAPIInvoke.Flags().String("host", "", "Override HTTP Host header. If unspecified, the host of --endpoint is used.")

	_ = cmdInternalAdminAPIInvoke.Flags().String("query", "", "The GraphQL query string to execute")
	_ = cmdInternalAdminAPIInvoke.Flags().String("query-file", "", "The path to a file containing the GraphQL query to execute")
	cmdInternalAdminAPIInvoke.MarkFlagsMutuallyExclusive("query", "query-file")
	cmdInternalAdminAPIInvoke.MarkFlagsOneRequired("query", "query-file")

	_ = cmdInternalAdminAPIInvoke.Flags().String("operation-name", "", "The operation to execute. Only necessary if the query contains more than 1 operation")

	_ = cmdInternalAdminAPIInvoke.Flags().String("variables-json", "", "The GraphQL variables in JSON format")
	_ = cmdInternalAdminAPIInvoke.Flags().String("variables-json-file", "", "The path to a file containing the GraphQL variables in JSON format")
	cmdInternalAdminAPIInvoke.MarkFlagsMutuallyExclusive("variables-json", "variables-json-file")

	_ = cmdInternalAdminAPIInvoke.Flags().BoolP("verbose", "v", false, "Print the HTTP/1.1 response to stderr")
}

var cmdInternalAdminAPI = &cobra.Command{
	Use:   "admin-api",
	Short: "Admin API commands",
}

var cmdInternalAdminAPIInvoke = &cobra.Command{
	Use:   "invoke",
	Short: "Invoke Admin API against a running Admin API server",
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := cmd.Context()

		binder := authgearcmd.GetBinder()

		dbURL, err := binder.GetRequiredString(cmd, authgearcmd.ArgDatabaseURL)
		if err != nil {
			return err
		}

		dbSchema, err := binder.GetRequiredString(cmd, authgearcmd.ArgDatabaseSchema)
		if err != nil {
			return err
		}
		dbPool := db.NewPool()

		credentials := &config.GlobalDatabaseCredentialsEnvironmentConfig{
			DatabaseURL:    dbURL,
			DatabaseSchema: dbSchema,
		}

		appID, err := cmd.Flags().GetString("app-id")
		if err != nil {
			return err
		}

		endpoint, err := cmd.Flags().GetString("endpoint")
		if err != nil {
			return err
		}

		host, err := cmd.Flags().GetString("host")
		if err != nil {
			return err
		}

		var query string
		if cmd.Flags().Lookup("query-file").Changed {
			var queryBytes []byte
			queryBytes, err = os.ReadFile(cmd.Flags().Lookup("query-file").Value.String())
			if err != nil {
				return err
			}
			query = string(queryBytes)
		} else {
			query = cmd.Flags().Lookup("query").Value.String()
		}

		operationName, err := cmd.Flags().GetString("operation-name")
		if err != nil {
			return err
		}

		var variablesJSON string
		if cmd.Flags().Lookup("variables-json-file").Changed {
			var variablesJSONBytes []byte
			variablesJSONBytes, err = os.ReadFile(cmd.Flags().Lookup("variables-json-file").Value.String())
			if err != nil {
				return err
			}
			variablesJSON = string(variablesJSONBytes)
		} else {
			variablesJSON = cmd.Flags().Lookup("variables-json").Value.String()
		}

		invoker := adminapi.NewInvoker(dbPool, credentials)
		adminAPIKey, err := invoker.FetchAdminAPIKeys(ctx, appID)
		if err != nil {
			return err
		}

		result, err := invoker.Invoke(ctx, adminapi.InvokeOptions{
			AppID:         appID,
			Endpoint:      endpoint,
			Host:          host,
			AdminAPIKey:   adminAPIKey,
			Query:         query,
			OperationName: operationName,
			VariablesJSON: variablesJSON,
		})
		if err != nil {
			return err
		}

		if cmd.Flags().Lookup("verbose").Value.String() == "true" {
			fmt.Fprintf(os.Stderr, "%v\n", string(result.DumpedResponse))
		}

		fmt.Fprintf(os.Stdout, "%v\n", string(result.HTTPBody))
		return nil
	},
}
