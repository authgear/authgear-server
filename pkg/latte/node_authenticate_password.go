package latte

import (
	"context"
	"errors"

	"github.com/authgear/authgear-server/pkg/api"
	"github.com/authgear/authgear-server/pkg/lib/authn/authenticator"
	"github.com/authgear/authgear-server/pkg/lib/workflow"
)

func init() {
	workflow.RegisterNode(&NodeAuthenticatePassword{})
}

type NodeAuthenticatePassword struct {
	Authenticator *authenticator.Info `json:"authenticator"`
}

func (n *NodeAuthenticatePassword) Kind() string {
	return "latte.NodeAuthenticatePassword"
}

func (n *NodeAuthenticatePassword) GetEffects(ctx context.Context, deps *workflow.Dependencies, w *workflow.Workflow) (effs []workflow.Effect, err error) {
	return nil, nil
}

func (n *NodeAuthenticatePassword) CanReactTo(ctx context.Context, deps *workflow.Dependencies, w *workflow.Workflow) ([]workflow.Input, error) {
	return []workflow.Input{
		&InputPassword{},
	}, nil
}

func (n *NodeAuthenticatePassword) ReactTo(ctx context.Context, deps *workflow.Dependencies, w *workflow.Workflow, input workflow.Input) (*workflow.Node, error) {
	var inputPassword inputPassword
	switch {
	case workflow.AsInput(input, &inputPassword):
		info := n.Authenticator
		_, err := deps.Authenticators.VerifyWithSpec(info, &authenticator.Spec{
			Password: &authenticator.PasswordSpec{
				PlainPassword: inputPassword.GetPassword(),
			},
		})
		if errors.Is(err, authenticator.ErrInvalidCredentials) {
			if err := DispatchAuthenticationFailedEvent(deps.Events, info); err != nil {
				return nil, err
			}
			return nil, api.ErrInvalidCredentials
		} else if err != nil {
			return nil, err
		}
		return workflow.NewNodeSimple(&NodeVerifiedAuthenticator{
			Authenticator: info,
		}), nil
	}
	return nil, workflow.ErrIncompatibleInput
}

func (n *NodeAuthenticatePassword) OutputData(ctx context.Context, deps *workflow.Dependencies, w *workflow.Workflow) (interface{}, error) {
	return map[string]interface{}{}, nil
}
