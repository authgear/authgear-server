package cmdinternal

import (
	"fmt"

	"github.com/spf13/cobra"

	portalcmd "github.com/authgear/authgear-server/cmd/portal/cmd"
	"github.com/authgear/authgear-server/cmd/portal/internal"
	"github.com/authgear/authgear-server/pkg/portal/model"
)

var cmdInternalCollaborator = &cobra.Command{
	Use:   "collaborator",
	Short: "Collaborator commands.",
}

var cmdInternalCollaboratorAdd = &cobra.Command{
	Use:   "add",
	Short: "Add a user as a collaborator of an app",
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := cmd.Context()

		binder := portalcmd.GetBinder()

		dbURL, err := binder.GetRequiredString(cmd, portalcmd.ArgDatabaseURL)
		if err != nil {
			return err
		}

		dbSchema, err := binder.GetRequiredString(cmd, portalcmd.ArgDatabaseSchema)
		if err != nil {
			return err
		}

		appID, err := cmd.Flags().GetString("app-id")
		if err != nil {
			return err
		}

		userID, err := cmd.Flags().GetString("user-id")
		if err != nil {
			return err
		}

		role, err := cmd.Flags().GetString("role")
		if err != nil {
			return err
		}
		switch role {
		case string(model.CollaboratorRoleOwner):
			break
		case string(model.CollaboratorRoleEditor):
			break
		default:
			return fmt.Errorf("invalid role: %v", role)
		}

		result, err := internal.AddCollaborator(ctx, internal.AddCollaboratorOptions{
			DatabaseURL:    dbURL,
			DatabaseSchema: dbSchema,
			AppID:          appID,
			UserID:         userID,
			Role:           model.CollaboratorRole(role),
		})
		if err != nil {
			return err
		}
		switch result {
		case internal.AddCollaboratorResultNoop:
			fmt.Printf("user (%v) is already (%v) of the app (%v)\n", userID, role, appID)
		case internal.AddCollaboratorResultInserted:
			fmt.Printf("user (%v) is added as (%v) to the app (%v)\n", userID, role, appID)
		case internal.AddCollaboratorResultUpdated:
			fmt.Printf("user (%v) is updated as (%v) of the app (%v)\n", userID, role, appID)
		}
		return nil
	},
}
