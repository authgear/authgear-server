package latte

import (
	"context"

	"github.com/authgear/authgear-server/pkg/lib/authn/authenticator"
	"github.com/authgear/authgear-server/pkg/lib/workflow"
)

func init() {
	workflow.RegisterNode(&NodeAuthenticateEmailLoginLink{})
}

type NodeAuthenticateEmailLoginLink struct {
	Authenticator *authenticator.Info `json:"authenticator"`
}

func (n *NodeAuthenticateEmailLoginLink) Kind() string {
	return "latte.NodeAuthenticateEmailLoginLink"
}

func (n *NodeAuthenticateEmailLoginLink) GetEffects(ctx context.Context, deps *workflow.Dependencies, w *workflow.Workflow) (effs []workflow.Effect, err error) {
	return nil, nil
}

func (n *NodeAuthenticateEmailLoginLink) CanReactTo(ctx context.Context, deps *workflow.Dependencies, w *workflow.Workflow) ([]workflow.Input, error) {
	return []workflow.Input{
		&InputResendCode{},
	}, nil
}

func (n *NodeAuthenticateEmailLoginLink) ReactTo(ctx context.Context, deps *workflow.Dependencies, w *workflow.Workflow, input workflow.Input) (*workflow.Node, error) {
	return nil, workflow.ErrIncompatibleInput
}

func (n *NodeAuthenticateEmailLoginLink) OutputData(ctx context.Context, deps *workflow.Dependencies, w *workflow.Workflow) (interface{}, error) {
	return map[string]interface{}{}, nil
}
