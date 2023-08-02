package workflowconfig

import (
	"context"

	"github.com/authgear/authgear-server/pkg/lib/authn/authenticator"
	"github.com/authgear/authgear-server/pkg/lib/workflow"
)

func init() {
	workflow.RegisterNode(&NodeLoginFlowChangePassword{})
}

type NodeLoginFlowChangePassword struct {
	Authenticator *authenticator.Info `json:"authenticator,omitempty"`
}

func (*NodeLoginFlowChangePassword) Kind() string {
	return "workflowconfig.NodeLoginFlowChangePassword"
}

func (*NodeLoginFlowChangePassword) CanReactTo(ctx context.Context, deps *workflow.Dependencies, workflows workflow.Workflows) ([]workflow.Input, error) {
	return []workflow.Input{&InputTakeNewPassword{}}, nil
}

func (i *NodeLoginFlowChangePassword) ReactTo(ctx context.Context, deps *workflow.Dependencies, workflows workflow.Workflows, input workflow.Input) (*workflow.Node, error) {
	var inputTakeNewPassword inputTakeNewPassword
	if workflow.AsInput(input, &inputTakeNewPassword) {
		newPassword := inputTakeNewPassword.GetNewPassword()

		oldInfo := i.Authenticator
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

func (*NodeLoginFlowChangePassword) GetEffects(ctx context.Context, deps *workflow.Dependencies, workflows workflow.Workflows) ([]workflow.Effect, error) {
	return nil, nil
}

func (*NodeLoginFlowChangePassword) OutputData(ctx context.Context, deps *workflow.Dependencies, workflows workflow.Workflows) (interface{}, error) {
	return nil, nil
}
