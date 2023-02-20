package latte

import (
	"context"
	"fmt"

	"github.com/authgear/authgear-server/pkg/api/model"
	"github.com/authgear/authgear-server/pkg/lib/authn/authenticator"
	"github.com/authgear/authgear-server/pkg/lib/workflow"
	"github.com/authgear/authgear-server/pkg/util/validation"
)

func init() {
	workflow.RegisterPrivateIntent(&IntentMigrateAuthenticators{})
}

var IntentMigrateAuthenticatorsSchema = validation.NewSimpleSchema(`{}`)

type IntentMigrateAuthenticators struct {
	UserID       string                       `json:"user_id"`
	MigrateSpecs []*authenticator.MigrateSpec `json:"migrate_specs"`
}

func (*IntentMigrateAuthenticators) Kind() string {
	return "latte.IntentMigrateAuthenticators"
}

func (*IntentMigrateAuthenticators) JSONSchema() *validation.SimpleSchema {
	return IntentMigrateAuthenticatorsSchema
}

func (i *IntentMigrateAuthenticators) CanReactTo(ctx context.Context, deps *workflow.Dependencies, w *workflow.Workflow) ([]workflow.Input, error) {
	// Create sub-workflows to migrate authenticators
	if len(w.Nodes) < len(i.MigrateSpecs) {
		return nil, nil
	}
	return nil, workflow.ErrEOF
}

func (i *IntentMigrateAuthenticators) ReactTo(ctx context.Context, deps *workflow.Dependencies, w *workflow.Workflow, input workflow.Input) (*workflow.Node, error) {
	if len(w.Nodes) >= len(i.MigrateSpecs) {
		return nil, workflow.ErrIncompatibleInput
	}

	idx := len(w.Nodes)
	spec := i.MigrateSpecs[idx]
	if spec.Type != model.AuthenticatorTypeOOBEmail && spec.Type != model.AuthenticatorTypeOOBSMS {
		panic(fmt.Sprintf("workflow: unsupported authenticator type for account migrations: %T", spec.Type))
	}
	return workflow.NewSubWorkflow(&IntentMigrateOOBOTPAuthenticator{
		UserID:      i.UserID,
		MigrateSpec: spec,
		// Mark the first authenticator in the migrate spec as default
		AuthenticatorIsDefault: idx == 0,
	}), nil
}

func (*IntentMigrateAuthenticators) GetEffects(ctx context.Context, deps *workflow.Dependencies, w *workflow.Workflow) (effs []workflow.Effect, err error) {
	return nil, nil
}

func (*IntentMigrateAuthenticators) OutputData(ctx context.Context, deps *workflow.Dependencies, w *workflow.Workflow) (interface{}, error) {
	return nil, nil
}
