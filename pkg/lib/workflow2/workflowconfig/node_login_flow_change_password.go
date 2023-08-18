package workflowconfig

import (
	"context"

	"github.com/authgear/authgear-server/pkg/lib/authn/authenticator"
	workflow "github.com/authgear/authgear-server/pkg/lib/workflow2"
)

func init() {
	workflow.RegisterNode(&NodeLoginFlowChangePassword{})
}

type NodeLoginFlowChangePasswordData struct {
	PasswordPolicy *PasswordPolicy `json:"password_policy,omitempty"`
}

func (NodeLoginFlowChangePasswordData) Data() {}

type NodeLoginFlowChangePassword struct {
	Authenticator *authenticator.Info `json:"authenticator,omitempty"`
}

var _ workflow.NodeSimple = &NodeLoginFlowChangePassword{}
var _ workflow.InputReactor = &NodeLoginFlowChangePassword{}
var _ workflow.DataOutputer = &NodeLoginFlowChangePassword{}

func (*NodeLoginFlowChangePassword) Kind() string {
	return "workflowconfig.NodeLoginFlowChangePassword"
}

func (*NodeLoginFlowChangePassword) CanReactTo(ctx context.Context, deps *workflow.Dependencies, workflows workflow.Workflows) (workflow.InputSchema, error) {
	return &InputTakeNewPassword{}, nil
}

func (n *NodeLoginFlowChangePassword) ReactTo(ctx context.Context, deps *workflow.Dependencies, workflows workflow.Workflows, input workflow.Input) (*workflow.Node, error) {
	var inputTakeNewPassword inputTakeNewPassword
	if workflow.AsInput(input, &inputTakeNewPassword) {
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
			// Nothing changed. End this workflow.
			return workflow.NewNodeSimple(&NodeSentinel{}), nil
		}

		return workflow.NewNodeSimple(&NodeDoUpdateAuthenticator{
			Authenticator: newInfo,
		}), nil
	}

	return nil, workflow.ErrIncompatibleInput
}

func (n *NodeLoginFlowChangePassword) OutputData(ctx context.Context, deps *workflow.Dependencies, workflows workflow.Workflows) (workflow.Data, error) {
	return NodeLoginFlowChangePasswordData{
		PasswordPolicy: NewPasswordPolicy(deps.Config.Authenticator.Password.Policy),
	}, nil
}
