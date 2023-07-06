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

func (i *IntentMigrateAuthenticators) CanReactTo(ctx context.Context, deps *workflow.Dependencies, workflows workflow.Workflows) ([]workflow.Input, error) {
	// Create sub-workflows to migrate authenticators
	if len(workflows.Nearest.Nodes) < len(i.MigrateSpecs) {
		return nil, nil
	}
	return nil, workflow.ErrEOF
}

func (i *IntentMigrateAuthenticators) ReactTo(ctx context.Context, deps *workflow.Dependencies, workflows workflow.Workflows, input workflow.Input) (*workflow.Node, error) {
	if len(workflows.Nearest.Nodes) >= len(i.MigrateSpecs) {
		return nil, workflow.ErrIncompatibleInput
	}

	idx := len(workflows.Nearest.Nodes)
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

func (*IntentMigrateAuthenticators) GetEffects(ctx context.Context, deps *workflow.Dependencies, workflows workflow.Workflows) (effs []workflow.Effect, err error) {
	return nil, nil
}

func (*IntentMigrateAuthenticators) OutputData(ctx context.Context, deps *workflow.Dependencies, workflows workflow.Workflows) (interface{}, error) {
	return nil, nil
}

func (*IntentMigrateAuthenticators) GetNewAuthenticators(w *workflow.Workflow) ([]*authenticator.Info, bool) {
	var authenticators []*authenticator.Info
	authenticatorWorkflows := workflow.FindSubWorkflows[NewAuthenticatorGetter](w)
	for _, subWorkflow := range authenticatorWorkflows {
		if a, ok := subWorkflow.Intent.(NewAuthenticatorGetter).GetNewAuthenticators(subWorkflow); ok {
			authenticators = append(authenticators, a...)
		}
	}

	if len(authenticators) == 0 {
		return nil, false
	}

	return authenticators, true
}
