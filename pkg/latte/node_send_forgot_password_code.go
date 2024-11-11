package latte

import (
	"context"

	"github.com/authgear/authgear-server/pkg/lib/workflow"
)

func init() {
	workflow.RegisterNode(&NodeSendForgotPasswordCode{})
}

type NodeSendForgotPasswordCode struct {
	LoginID string `json:"login_id"`
}

func (n *NodeSendForgotPasswordCode) Kind() string {
	return "latte.NodeSendForgotPasswordCode"
}

func (n *NodeSendForgotPasswordCode) GetEffects(ctx context.Context, deps *workflow.Dependencies, workflows workflow.Workflows) (effs []workflow.Effect, err error) {
	return nil, nil
}

func (*NodeSendForgotPasswordCode) CanReactTo(ctx context.Context, deps *workflow.Dependencies, workflows workflow.Workflows) ([]workflow.Input, error) {
	return nil, workflow.ErrEOF
}

func (*NodeSendForgotPasswordCode) ReactTo(ctx context.Context, deps *workflow.Dependencies, workflows workflow.Workflows, input workflow.Input) (*workflow.Node, error) {
	return nil, workflow.ErrIncompatibleInput
}

func (n *NodeSendForgotPasswordCode) OutputData(ctx context.Context, deps *workflow.Dependencies, workflows workflow.Workflows) (interface{}, error) {
	return nil, nil
}

func (n *NodeSendForgotPasswordCode) sendCode(ctx context.Context, deps *workflow.Dependencies) error {
	return deps.ForgotPassword.SendCode(ctx, n.LoginID, nil)

}
