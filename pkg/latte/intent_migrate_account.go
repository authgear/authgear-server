package latte

import (
	"context"

	"github.com/authgear/authgear-server/pkg/lib/authn/authenticator"
	"github.com/authgear/authgear-server/pkg/lib/authn/identity"
	"github.com/authgear/authgear-server/pkg/lib/workflow"
	"github.com/authgear/authgear-server/pkg/util/validation"
)

func init() {
	workflow.RegisterPrivateIntent(&IntentMigrateAccount{})
}

var IntentMigrateAccountSchema = validation.NewSimpleSchema(`{}`)

type IntentMigrateAccount struct {
	UseID string `json:"user_id"`
}

func (*IntentMigrateAccount) Kind() string {
	return "latte.IntentMigrateAccount"
}

func (*IntentMigrateAccount) JSONSchema() *validation.SimpleSchema {
	return IntentMigrateAccountSchema
}

func (*IntentMigrateAccount) CanReactTo(ctx context.Context, deps *workflow.Dependencies, w *workflow.Workflow) ([]workflow.Input, error) {
	switch len(w.Nodes) {
	case 0:
		// Resolve migration spec from the migration token.
		return []workflow.Input{
			&InputTakeMigrationToken{},
		}, nil
	case 1:
		// Migrate identities.
		return nil, nil
	case 2:
		// Migrate authenticators.
		return nil, nil
	}

	return nil, workflow.ErrEOF
}

func (i *IntentMigrateAccount) ReactTo(ctx context.Context, deps *workflow.Dependencies, w *workflow.Workflow, input workflow.Input) (*workflow.Node, error) {
	switch len(w.Nodes) {
	case 0:
		var inputTakeMigrationToken inputTakeMigrationToken
		if workflow.AsInput(input, &inputTakeMigrationToken) {
			token := inputTakeMigrationToken.GetMigrationToken()
			resp, err := deps.AccountMigrations.Run(token)
			if err != nil {
				return nil, err
			}
			return workflow.NewNodeSimple(&NodeMigrateAccount{
				IdentityMigrateSpecs:      resp.Identities,
				AuthenticatorMigrateSpecs: resp.Authenticators,
			}), nil
		}
	case 1:
		specs := i.getIdentityMigrateSpecs(w)
		return workflow.NewSubWorkflow(&IntentMigrateIdentities{
			UserID:       i.UseID,
			MigrateSpecs: specs,
		}), nil
	case 2:
		specs := i.getAuthenticatorMigrateSpecs(w)
		return workflow.NewSubWorkflow(&IntentMigrateAuthenticators{
			UserID:       i.UseID,
			MigrateSpecs: specs,
		}), nil
	}

	return nil, workflow.ErrIncompatibleInput
}

func (i *IntentMigrateAccount) GetEffects(ctx context.Context, deps *workflow.Dependencies, w *workflow.Workflow) (effs []workflow.Effect, err error) {
	return nil, nil
}

func (*IntentMigrateAccount) OutputData(ctx context.Context, deps *workflow.Dependencies, w *workflow.Workflow) (interface{}, error) {
	return nil, nil
}

func (*IntentMigrateAccount) getIdentityMigrateSpecs(w *workflow.Workflow) []*identity.MigrateSpec {
	node, ok := workflow.FindSingleNode[*NodeMigrateAccount](w)
	if !ok {
		return []*identity.MigrateSpec{}
	}
	return node.IdentityMigrateSpecs
}

func (*IntentMigrateAccount) getAuthenticatorMigrateSpecs(w *workflow.Workflow) []*authenticator.MigrateSpec {
	node, ok := workflow.FindSingleNode[*NodeMigrateAccount](w)
	if !ok {
		return []*authenticator.MigrateSpec{}
	}
	return node.AuthenticatorMigrateSpecs
}
