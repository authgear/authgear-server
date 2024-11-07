package latte

import (
	"context"
	"errors"

	"github.com/authgear/authgear-server/pkg/api"
	"github.com/authgear/authgear-server/pkg/api/model"
	"github.com/authgear/authgear-server/pkg/lib/authn"
	"github.com/authgear/authgear-server/pkg/lib/authn/authenticator"
	"github.com/authgear/authgear-server/pkg/lib/facade"
	"github.com/authgear/authgear-server/pkg/lib/workflow"
)

func init() {
	workflow.RegisterNode(&NodeAuthenticatePassword{})
}

type NodeAuthenticatePassword struct {
	UserID            string             `json:"user_id,omitempty"`
	AuthenticatorKind authenticator.Kind `json:"authenticator_kind,omitempty"`
}

func (n *NodeAuthenticatePassword) Kind() string {
	return "latte.NodeAuthenticatePassword"
}

func (n *NodeAuthenticatePassword) GetEffects(ctx context.Context, deps *workflow.Dependencies, workflows workflow.Workflows) (effs []workflow.Effect, err error) {
	return nil, nil
}

func (n *NodeAuthenticatePassword) CanReactTo(ctx context.Context, deps *workflow.Dependencies, workflows workflow.Workflows) ([]workflow.Input, error) {
	return []workflow.Input{
		&InputTakePassword{},
	}, nil
}

func (n *NodeAuthenticatePassword) ReactTo(ctx context.Context, deps *workflow.Dependencies, workflows workflow.Workflows, input workflow.Input) (*workflow.Node, error) {
	var inputTakePassword inputTakePassword
	switch {
	case workflow.AsInput(input, &inputTakePassword):
		info, err := n.getPasswordAuthenticator(ctx, deps)
		// The user doesn't have the password authenticator
		// always returns invalid credentials error
		if errors.Is(err, api.ErrNoAuthenticator) {
			return nil, api.ErrInvalidCredentials
		} else if err != nil {
			return nil, err
		}
		_, err = deps.Authenticators.VerifyWithSpec(ctx, info, &authenticator.Spec{
			Password: &authenticator.PasswordSpec{
				PlainPassword: inputTakePassword.GetPassword(),
			},
		}, &facade.VerifyOptions{
			AuthenticationDetails: facade.NewAuthenticationDetails(
				info.UserID,
				authn.AuthenticationStageSecondary,
				authn.AuthenticationTypePassword,
			),
		})
		if err != nil {
			return nil, err
		}
		return workflow.NewNodeSimple(&NodeVerifiedAuthenticator{
			Authenticator: info,
		}), nil
	}
	return nil, workflow.ErrIncompatibleInput
}

func (n *NodeAuthenticatePassword) OutputData(ctx context.Context, deps *workflow.Dependencies, workflows workflow.Workflows) (interface{}, error) {
	return map[string]interface{}{}, nil
}

func (n *NodeAuthenticatePassword) getPasswordAuthenticator(ctx context.Context, deps *workflow.Dependencies) (*authenticator.Info, error) {
	ais, err := deps.Authenticators.List(ctx,
		n.UserID,
		authenticator.KeepKind(n.AuthenticatorKind),
		authenticator.KeepType(model.AuthenticatorTypePassword),
	)
	if err != nil {
		return nil, err
	}

	if len(ais) == 0 {
		return nil, api.ErrNoAuthenticator
	}

	return ais[0], nil
}
