package latte

import (
	"context"

	"github.com/authgear/authgear-server/pkg/api"
	"github.com/authgear/authgear-server/pkg/api/model"
	"github.com/authgear/authgear-server/pkg/lib/authn/authenticator"
	"github.com/authgear/authgear-server/pkg/lib/authn/identity"
	"github.com/authgear/authgear-server/pkg/lib/workflow"
	"github.com/authgear/authgear-server/pkg/util/validation"
)

func init() {
	workflow.RegisterPrivateIntent(&IntentLogin{})
}

var IntentLoginSchema = validation.NewSimpleSchema(`{}`)

type IntentLogin struct {
	Identity *identity.Info `json:"identity"`
}

func (*IntentLogin) Kind() string {
	return "latte.IntentLogin"
}

func (*IntentLogin) JSONSchema() *validation.SimpleSchema {
	return IntentLoginSchema
}

func (*IntentLogin) CanReactTo(ctx context.Context, deps *workflow.Dependencies, w *workflow.Workflow) ([]workflow.Input, error) {
	switch len(w.Nodes) {
	case 0:
		return nil, nil
	}

	return nil, workflow.ErrEOF
}

func (i *IntentLogin) ReactTo(ctx context.Context, deps *workflow.Dependencies, w *workflow.Workflow, input workflow.Input) (*workflow.Node, error) {
	switch len(w.Nodes) {
	case 0:
		// 1st step: authenticate oob otp phone
		phoneAuthenticator, err := i.getAuthenticator(deps,
			authenticator.KeepPrimaryAuthenticatorOfIdentity(i.Identity),
			authenticator.KeepType(model.AuthenticatorTypeOOBSMS),
		)
		if err != nil {
			return nil, err
		}
		return workflow.NewSubWorkflow(&IntentAuthenticateOOBOTPPhone{
			Authenticator: phoneAuthenticator,
		}), nil
	}
	return nil, workflow.ErrIncompatibleInput
}

func (*IntentLogin) GetEffects(ctx context.Context, deps *workflow.Dependencies, w *workflow.Workflow) (effs []workflow.Effect, err error) {
	return nil, nil
}

func (i *IntentLogin) OutputData(ctx context.Context, deps *workflow.Dependencies, w *workflow.Workflow) (interface{}, error) {
	return nil, nil
}

func (i *IntentLogin) getAuthenticator(deps *workflow.Dependencies, filters ...authenticator.Filter) (*authenticator.Info, error) {
	ais, err := deps.Authenticators.List(i.Identity.UserID, filters...)
	if err != nil {
		return nil, err
	}

	if len(ais) == 0 {
		return nil, api.ErrNoAuthenticator
	}

	return ais[0], nil
}
