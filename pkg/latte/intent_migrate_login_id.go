package latte

import (
	"context"
	"fmt"

	"github.com/authgear/authgear-server/pkg/api/model"
	"github.com/authgear/authgear-server/pkg/lib/authn/identity"
	"github.com/authgear/authgear-server/pkg/lib/feature/verification"
	"github.com/authgear/authgear-server/pkg/lib/workflow"
	"github.com/authgear/authgear-server/pkg/util/validation"
)

func init() {
	workflow.RegisterPrivateIntent(&IntentMigrateLoginID{})
}

var IntentMigrateLoginIDSchema = validation.NewSimpleSchema(`{}`)

type IntentMigrateLoginID struct {
	UserID      string                `json:"user_id"`
	MigrateSpec *identity.MigrateSpec `json:"migrate_spec"`
}

func (*IntentMigrateLoginID) Kind() string {
	return "latte.IntentMigrateLoginID"
}

func (*IntentMigrateLoginID) JSONSchema() *validation.SimpleSchema {
	return IntentMigrateLoginIDSchema
}

func (*IntentMigrateLoginID) CanReactTo(ctx context.Context, deps *workflow.Dependencies, workflows workflow.Workflows) ([]workflow.Input, error) {
	switch len(workflows.Nearest.Nodes) {
	case 0:
		// Create identity.
		return nil, nil
	case 1:
		// Populate standard attributes.
		return nil, nil
	case 2:
		// Mark identity as verified automatically.
		return nil, nil
	default:
		return nil, workflow.ErrEOF
	}
}

func (i *IntentMigrateLoginID) ReactTo(ctx context.Context, deps *workflow.Dependencies, workflows workflow.Workflows, input workflow.Input) (*workflow.Node, error) {
	switch len(workflows.Nearest.Nodes) {
	case 0:
		spec := i.MigrateSpec.GetSpec()
		info, err := deps.Identities.New(i.UserID, spec, identity.NewIdentityOptions{
			LoginIDEmailByPassBlocklistAllowlist: false,
		})
		if err != nil {
			return nil, err
		}

		_, err = deps.Identities.CheckDuplicated(info)
		if err != nil {
			return nil, err
		}

		return workflow.NewNodeSimple(&NodeDoCreateIdentity{
			Identity: info,
		}), nil
	case 1:
		iden := i.identityInfo(workflows.Nearest)
		return workflow.NewNodeSimple(&NodePopulateStandardAttributes{
			Identity: iden,
		}), nil
	case 2:
		iden := i.identityInfo(workflows.Nearest)
		var verifiedClaim *verification.Claim
		switch iden.LoginID.LoginIDType {
		case model.LoginIDKeyTypeEmail:
			verifiedClaim = deps.Verification.NewVerifiedClaim(i.UserID, string(model.ClaimEmail), iden.LoginID.LoginID)
		case model.LoginIDKeyTypePhone:
			verifiedClaim = deps.Verification.NewVerifiedClaim(i.UserID, string(model.ClaimPhoneNumber), iden.LoginID.LoginID)
		}
		return workflow.NewNodeSimple(&NodeVerifiedIdentity{
			IdentityID:       iden.ID,
			NewVerifiedClaim: verifiedClaim,
		}), nil
	}
	return nil, workflow.ErrIncompatibleInput
}

func (i *IntentMigrateLoginID) identityInfo(w *workflow.Workflow) *identity.Info {
	node, ok := workflow.FindSingleNode[*NodeDoCreateIdentity](w)
	if !ok {
		panic(fmt.Errorf("workflow: expected NodeCreateIdentity"))
	}
	return node.Identity
}

func (*IntentMigrateLoginID) GetEffects(ctx context.Context, deps *workflow.Dependencies, workflows workflow.Workflows) (effs []workflow.Effect, err error) {
	return nil, nil
}

func (*IntentMigrateLoginID) OutputData(ctx context.Context, deps *workflow.Dependencies, workflows workflow.Workflows) (interface{}, error) {
	return nil, nil
}

func (*IntentMigrateLoginID) GetNewIdentities(w *workflow.Workflow) ([]*identity.Info, bool) {
	node, ok := workflow.FindSingleNode[*NodeDoCreateIdentity](w)
	if !ok {
		return nil, false
	}
	return []*identity.Info{node.Identity}, true
}
