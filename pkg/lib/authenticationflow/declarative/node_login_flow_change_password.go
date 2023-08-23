package declarative

import (
	"context"

	authflow "github.com/authgear/authgear-server/pkg/lib/authenticationflow"
	"github.com/authgear/authgear-server/pkg/lib/authn/authenticator"
)

func init() {
	authflow.RegisterNode(&NodeLoginFlowChangePassword{})
}

type NodeLoginFlowChangePasswordData struct {
	PasswordPolicy *PasswordPolicy `json:"password_policy,omitempty"`
}

func (NodeLoginFlowChangePasswordData) Data() {}

type NodeLoginFlowChangePassword struct {
	Authenticator *authenticator.Info `json:"authenticator,omitempty"`
}

var _ authflow.NodeSimple = &NodeLoginFlowChangePassword{}
var _ authflow.InputReactor = &NodeLoginFlowChangePassword{}
var _ authflow.DataOutputer = &NodeLoginFlowChangePassword{}

func (*NodeLoginFlowChangePassword) Kind() string {
	return "NodeLoginFlowChangePassword"
}

func (*NodeLoginFlowChangePassword) CanReactTo(ctx context.Context, deps *authflow.Dependencies, flows authflow.Flows) (authflow.InputSchema, error) {
	return &InputTakeNewPassword{}, nil
}

func (n *NodeLoginFlowChangePassword) ReactTo(ctx context.Context, deps *authflow.Dependencies, flows authflow.Flows, input authflow.Input) (*authflow.Node, error) {
	var inputTakeNewPassword inputTakeNewPassword
	if authflow.AsInput(input, &inputTakeNewPassword) {
		newPassword := inputTakeNewPassword.GetNewPassword()

		oldInfo := n.Authenticator
		changed, newInfo, err := deps.Authenticators.WithSpec(oldInfo, &authenticator.Spec{
			Password: &authenticator.PasswordSpec{
				PlainPassword: newPassword,
			},
		})
		if err != nil {
			return nil, err
		}

		if !changed {
			// Nothing changed. End this flow.
			return authflow.NewNodeSimple(&NodeSentinel{}), nil
		}

		return authflow.NewNodeSimple(&NodeDoUpdateAuthenticator{
			Authenticator: newInfo,
		}), nil
	}

	return nil, authflow.ErrIncompatibleInput
}

func (n *NodeLoginFlowChangePassword) OutputData(ctx context.Context, deps *authflow.Dependencies, flows authflow.Flows) (authflow.Data, error) {
	return NodeLoginFlowChangePasswordData{
		PasswordPolicy: NewPasswordPolicy(deps.Config.Authenticator.Password.Policy),
	}, nil
}
