package workflowconfig

import (
	"context"

	"github.com/authgear/authgear-server/pkg/api/model"
	"github.com/authgear/authgear-server/pkg/lib/authn/authenticator"
	"github.com/authgear/authgear-server/pkg/lib/config"
	workflow "github.com/authgear/authgear-server/pkg/lib/workflow2"
)

func init() {
	workflow.RegisterNode(&NodeLoginFlowChangePassword{})
}

type NodeLoginFlowChangePasswordData struct {
	Candidates []AuthenticationCandidate `json:"candidates"`
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

func (*NodeLoginFlowChangePassword) CanReactTo(ctx context.Context, deps *workflow.Dependencies, workflows workflow.Workflows) ([]workflow.Input, error) {
	return []workflow.Input{&InputTakeNewPassword{}}, nil
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
	var candidate AuthenticationCandidate
	switch n.Authenticator.Kind {
	case model.AuthenticatorKindPrimary:
		candidate = NewAuthenticationCandidateFromMethod(config.WorkflowAuthenticationMethodPrimaryPassword)
	case model.AuthenticatorKindSecondary:
		candidate = NewAuthenticationCandidateFromMethod(config.WorkflowAuthenticationMethodSecondaryPassword)
	}

	return NodeLoginFlowChangePasswordData{
		Candidates: []AuthenticationCandidate{candidate},
	}, nil
}
