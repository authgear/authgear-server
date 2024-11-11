package latte

import (
	"context"

	"github.com/authgear/authgear-server/pkg/lib/authn/authenticator"
	"github.com/authgear/authgear-server/pkg/lib/authn/otp"
	"github.com/authgear/authgear-server/pkg/lib/workflow"
	"github.com/authgear/authgear-server/pkg/util/validation"
)

func init() {
	workflow.RegisterPrivateIntent(&IntentAuthenticateEmailLoginLink{})
}

var IntentAuthenticateEmailLoginLinkSchema = validation.NewSimpleSchema(`{}`)

type IntentAuthenticateEmailLoginLink struct {
	Authenticator *authenticator.Info `json:"authenticator,omitempty"`
}

func (i *IntentAuthenticateEmailLoginLink) Kind() string {
	return "latte.IntentAuthenticateEmailLoginLink"
}

func (i *IntentAuthenticateEmailLoginLink) JSONSchema() *validation.SimpleSchema {
	return IntentAuthenticateEmailLoginLinkSchema
}

func (i *IntentAuthenticateEmailLoginLink) CanReactTo(ctx context.Context, deps *workflow.Dependencies, workflows workflow.Workflows) ([]workflow.Input, error) {
	switch len(workflows.Nearest.Nodes) {
	case 0:
		return nil, nil
	}
	return nil, workflow.ErrEOF
}

func (i *IntentAuthenticateEmailLoginLink) ReactTo(ctx context.Context, deps *workflow.Dependencies, workflows workflow.Workflows, input workflow.Input) (*workflow.Node, error) {
	switch len(workflows.Nearest.Nodes) {
	case 0:
		authenticator := i.Authenticator
		err := (&SendOOBCode{
			WorkflowID:        workflow.GetWorkflowID(ctx),
			Deps:              deps,
			Stage:             authenticatorKindToStage(authenticator.Kind),
			IsAuthenticating:  true,
			AuthenticatorInfo: authenticator,
			OTPForm:           otp.FormLink,
		}).Do(ctx)
		if err != nil {
			return nil, err
		}

		return workflow.NewNodeSimple(&NodeAuthenticateEmailLoginLink{
			Authenticator: authenticator,
		}), nil
	}
	return nil, workflow.ErrIncompatibleInput
}

func (i *IntentAuthenticateEmailLoginLink) GetEffects(ctx context.Context, deps *workflow.Dependencies, workflows workflow.Workflows) (effs []workflow.Effect, err error) {
	return nil, nil
}

func (i *IntentAuthenticateEmailLoginLink) OutputData(ctx context.Context, deps *workflow.Dependencies, workflows workflow.Workflows) (interface{}, error) {
	return map[string]interface{}{}, nil
}

func (i *IntentAuthenticateEmailLoginLink) GetAMR(w *workflow.Workflow) []string {
	node, ok := workflow.FindSingleNode[*NodeVerifiedAuthenticator](w)
	if !ok {
		return []string{}
	}
	return node.GetAMR()
}

var _ AMRGetter = &IntentAuthenticateEmailLoginLink{}
