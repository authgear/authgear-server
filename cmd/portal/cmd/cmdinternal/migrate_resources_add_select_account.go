package cmdinternal

import (
	"context"
	"encoding/base64"
	"fmt"

	"github.com/spf13/cobra"
	"sigs.k8s.io/yaml"

	portalcmd "github.com/authgear/authgear-server/cmd/portal/cmd"
	"github.com/authgear/authgear-server/cmd/portal/internal"
)

var cmdInternalMigrateAddSelectAccount = &cobra.Command{
	Use:   "migrate-add-select-account",
	Short: "Add select_account to every login flow whose first step is identify",
	RunE: func(cmd *cobra.Command, args []string) error {
		binder := portalcmd.GetBinder()

		dbURL, err := binder.GetRequiredString(cmd, portalcmd.ArgDatabaseURL)
		if err != nil {
			return err
		}

		dbSchema, err := binder.GetRequiredString(cmd, portalcmd.ArgDatabaseSchema)
		if err != nil {
			return err
		}

		internal.MigrateResources(cmd.Context(), &internal.MigrateResourcesOptions{
			DatabaseURL:            dbURL,
			DatabaseSchema:         dbSchema,
			UpdateConfigSourceFunc: migrateAddSelectAccount,
			DryRun:                 &MigrateResourcesDryRun,
		})

		return nil
	},
}

// migrateAddSelectAccount prepends `- identification: select_account` to the
// `one_of` of the first step of every entry in authentication_flow.login_flows
// whose first step is `type: identify`.
//
// The flow's name is not considered, so an app can receive several insertions.
// Only login_flows is touched; signup_flows, reauth_flows and
// signup_login_flows are left alone. A flow that already offers select_account
// is skipped without affecting its siblings, so this is idempotent.
func migrateAddSelectAccount(ctx context.Context, appID string, configSourceData map[string]string, dryRun bool) error {
	encodedData := configSourceData["authgear.yaml"]
	decoded, err := base64.StdEncoding.DecodeString(encodedData)
	if err != nil {
		return fmt.Errorf("failed decode authgear.yaml: %w", err)
	}

	m := make(map[string]any)
	err = yaml.Unmarshal(decoded, &m)
	if err != nil {
		return fmt.Errorf("failed unmarshal yaml: %w", err)
	}

	loginFlows, ok := mapGet[[]any](m, "authentication_flow", "login_flows")
	if !ok {
		return nil
	}

	updated := false
	for _, loginFlow := range loginFlows {
		flow, ok := loginFlow.(map[string]any)
		if !ok {
			continue
		}

		steps, ok := mapGet[[]any](flow, "steps")
		if !ok || len(steps) == 0 {
			continue
		}

		firstStep, ok := steps[0].(map[string]any)
		if !ok {
			continue
		}

		stepType, ok := mapGet[string](firstStep, "type")
		if !ok || stepType != "identify" {
			continue
		}

		oneOf, ok := mapGet[[]any](firstStep, "one_of")
		if !ok {
			continue
		}

		if hasSelectAccount(oneOf) {
			continue
		}

		newOneOf := make([]any, 0, len(oneOf)+1)
		newOneOf = append(newOneOf, map[string]any{"identification": "select_account"})
		newOneOf = append(newOneOf, oneOf...)
		mapSet(firstStep, newOneOf, "one_of")
		updated = true
	}

	if !updated {
		return nil
	}

	migrated, err := yaml.Marshal(m)
	if err != nil {
		return fmt.Errorf("failed marshal yaml: %w", err)
	}

	configSourceData["authgear.yaml"] = base64.StdEncoding.EncodeToString(migrated)
	return nil
}

func hasSelectAccount(oneOf []any) bool {
	for _, b := range oneOf {
		branch, ok := b.(map[string]any)
		if !ok {
			continue
		}
		identification, ok := mapGet[string](branch, "identification")
		if ok && identification == "select_account" {
			return true
		}
	}
	return false
}

func init() {
	cmdInternalBreakingChangeMigrateResources.AddCommand(cmdInternalMigrateAddSelectAccount)
}
