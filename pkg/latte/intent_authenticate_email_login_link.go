package latte

import (
	"context"

	"github.com/authgear/authgear-server/pkg/lib/authn/authenticator"
	"github.com/authgear/authgear-server/pkg/lib/workflow"
	"github.com/authgear/authgear-server/pkg/util/validation"
)

func init() {
	workflow.RegisterPrivateIntent(&IntentAuthenticateEmailLoginLink{})
}

var IntentAuthenticateEmailLoginLinkSchema = validation.NewSimpleSchema(`{}`)

type IntentAuthenticateEmailLoginLink struct {
	Authenticator *authenticator.Info `json:"authenticator"`
}

func (i *IntentAuthenticateEmailLoginLink) Kind() string {
	return "latte.IntentAuthenticateEmailLoginLink"
}

func (i *IntentAuthenticateEmailLoginLink) JSONSchema() *validation.SimpleSchema {
	return IntentAuthenticateEmailLoginLinkSchema
}

func (i *IntentAuthenticateEmailLoginLink) CanReactTo(ctx context.Context, deps *workflow.Dependencies, w *workflow.Workflow) ([]workflow.Input, error) {
	switch len(w.Nodes) {
	case 0:
		return nil, nil
	}
	return nil, workflow.ErrEOF
}

func (i *IntentAuthenticateEmailLoginLink) ReactTo(ctx context.Context, deps *workflow.Dependencies, w *workflow.Workflow, input workflow.Input) (*workflow.Node, error) {
	switch len(w.Nodes) {
	case 0:
		authenticator := i.Authenticator
		return workflow.NewNodeSimple(&NodeAuthenticateEmailLoginLink{
			Authenticator: authenticator,
		}), nil
	}
	return nil, workflow.ErrIncompatibleInput
}

func (i *IntentAuthenticateEmailLoginLink) GetEffects(ctx context.Context, deps *workflow.Dependencies, w *workflow.Workflow) (effs []workflow.Effect, err error) {
	return nil, nil
}

func (i *IntentAuthenticateEmailLoginLink) OutputData(ctx context.Context, deps *workflow.Dependencies, w *workflow.Workflow) (interface{}, error) {
	return nil, nil
}
