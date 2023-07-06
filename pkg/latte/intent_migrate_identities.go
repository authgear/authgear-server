package latte

import (
	"context"
	"fmt"

	"github.com/authgear/authgear-server/pkg/api/model"
	"github.com/authgear/authgear-server/pkg/lib/authn/identity"
	"github.com/authgear/authgear-server/pkg/lib/workflow"
	"github.com/authgear/authgear-server/pkg/util/validation"
)

func init() {
	workflow.RegisterPrivateIntent(&IntentMigrateIdentities{})
}

var IntentMigrateIdentitiesSchema = validation.NewSimpleSchema(`{}`)

type IntentMigrateIdentities struct {
	UserID       string                  `json:"user_id"`
	MigrateSpecs []*identity.MigrateSpec `json:"migrate_specs"`
}

func (*IntentMigrateIdentities) Kind() string {
	return "latte.IntentMigrateIdentities"
}

func (*IntentMigrateIdentities) JSONSchema() *validation.SimpleSchema {
	return IntentMigrateIdentitiesSchema
}

func (i *IntentMigrateIdentities) CanReactTo(ctx context.Context, deps *workflow.Dependencies, workflows workflow.Workflows) ([]workflow.Input, error) {
	// Create sub-workflows to migrate identities
	if len(workflows.Nearest.Nodes) < len(i.MigrateSpecs) {
		return nil, nil
	}
	return nil, workflow.ErrEOF
}

func (i *IntentMigrateIdentities) ReactTo(ctx context.Context, deps *workflow.Dependencies, workflows workflow.Workflows, input workflow.Input) (*workflow.Node, error) {
	if len(workflows.Nearest.Nodes) >= len(i.MigrateSpecs) {
		return nil, workflow.ErrIncompatibleInput
	}

	idx := len(workflows.Nearest.Nodes)
	spec := i.MigrateSpecs[idx]
	if spec.Type != model.IdentityTypeLoginID {
		panic(fmt.Sprintf("workflow: unsupported identity type for account migrations: %T", spec.Type))
	}
	return workflow.NewSubWorkflow(&IntentMigrateLoginID{
		UserID:      i.UserID,
		MigrateSpec: spec,
	}), nil
}

func (*IntentMigrateIdentities) GetEffects(ctx context.Context, deps *workflow.Dependencies, workflows workflow.Workflows) (effs []workflow.Effect, err error) {
	return nil, nil
}

func (*IntentMigrateIdentities) OutputData(ctx context.Context, deps *workflow.Dependencies, workflows workflow.Workflows) (interface{}, error) {
	return nil, nil
}

func (*IntentMigrateIdentities) GetNewIdentities(w *workflow.Workflow) ([]*identity.Info, bool) {
	var identities []*identity.Info
	identityWorkflows := workflow.FindSubWorkflows[NewIdentityGetter](w)
	for _, subWorkflow := range identityWorkflows {
		if iden, ok := subWorkflow.Intent.(NewIdentityGetter).GetNewIdentities(subWorkflow); ok {
			identities = append(identities, iden...)
		}
	}

	if len(identities) == 0 {
		return nil, false
	}

	return identities, true
}
