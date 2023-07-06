package latte

import (
	"context"

	"github.com/authgear/authgear-server/pkg/lib/authn/authenticator"
	"github.com/authgear/authgear-server/pkg/lib/workflow"
	"github.com/authgear/authgear-server/pkg/util/validation"
)

func init() {
	workflow.RegisterPrivateIntent(&IntentAuthenticatePassword{})
}

var IntentAuthenticatePasswordSchema = validation.NewSimpleSchema(`{}`)

type IntentAuthenticatePassword struct {
	UserID            string             `json:"user_id,omitempty"`
	AuthenticatorKind authenticator.Kind `json:"authenticator_kind,omitempty"`
}

func (i *IntentAuthenticatePassword) Kind() string {
	return "latte.IntentAuthenticatePassword"
}

func (i *IntentAuthenticatePassword) JSONSchema() *validation.SimpleSchema {
	return IntentAuthenticatePasswordSchema
}

func (i *IntentAuthenticatePassword) CanReactTo(ctx context.Context, deps *workflow.Dependencies, workflows workflow.Workflows) ([]workflow.Input, error) {
	switch len(workflows.Nearest.Nodes) {
	case 0:
		return nil, nil
	}
	return nil, workflow.ErrEOF
}

func (i *IntentAuthenticatePassword) ReactTo(ctx context.Context, deps *workflow.Dependencies, workflows workflow.Workflows, input workflow.Input) (*workflow.Node, error) {
	switch len(workflows.Nearest.Nodes) {
	case 0:
		return workflow.NewNodeSimple(&NodeAuthenticatePassword{
			UserID:            i.UserID,
			AuthenticatorKind: i.AuthenticatorKind,
		}), nil
	}
	return nil, workflow.ErrIncompatibleInput
}

func (i *IntentAuthenticatePassword) GetEffects(ctx context.Context, deps *workflow.Dependencies, workflows workflow.Workflows) (effs []workflow.Effect, err error) {
	return nil, nil
}

func (i *IntentAuthenticatePassword) OutputData(ctx context.Context, deps *workflow.Dependencies, workflows workflow.Workflows) (interface{}, error) {
	return nil, nil
}

func (i *IntentAuthenticatePassword) GetAMR(w *workflow.Workflow) []string {
	node, ok := workflow.FindSingleNode[*NodeVerifiedAuthenticator](w)
	if !ok {
		return []string{}
	}
	return node.GetAMR()
}

var _ AMRGetter = &IntentAuthenticatePassword{}
