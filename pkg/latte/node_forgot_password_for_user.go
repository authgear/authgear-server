package latte

import (
	"context"

	"github.com/authgear/authgear-server/pkg/lib/workflow"
)

func init() {
	workflow.RegisterNode(&NodeForgotPasswordForUser{})
}

type NodeForgotPasswordForUser struct {
	LoginID string  `json:"login_id"`
	UserID  *string `json:"user_id"`
}

func (n *NodeForgotPasswordForUser) Kind() string {
	return "latte.NodeForgotPasswordForUser"
}

func (n *NodeForgotPasswordForUser) GetEffects(ctx context.Context, deps *workflow.Dependencies, workflows workflow.Workflows) (effs []workflow.Effect, err error) {
	return nil, nil
}

func (*NodeForgotPasswordForUser) CanReactTo(ctx context.Context, deps *workflow.Dependencies, workflows workflow.Workflows) ([]workflow.Input, error) {
	return nil, workflow.ErrEOF
}

func (*NodeForgotPasswordForUser) ReactTo(ctx context.Context, deps *workflow.Dependencies, workflows workflow.Workflows, input workflow.Input) (*workflow.Node, error) {
	return nil, workflow.ErrIncompatibleInput
}

func (*NodeForgotPasswordForUser) OutputData(ctx context.Context, deps *workflow.Dependencies, workflows workflow.Workflows) (interface{}, error) {
	return nil, nil
}
