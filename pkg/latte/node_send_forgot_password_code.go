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

func (n *NodeSendForgotPasswordCode) GetEffects(ctx context.Context, deps *workflow.Dependencies, w *workflow.Workflow) (effs []workflow.Effect, err error) {
	return nil, nil
}

func (*NodeSendForgotPasswordCode) CanReactTo(ctx context.Context, deps *workflow.Dependencies, w *workflow.Workflow) ([]workflow.Input, error) {
	return nil, workflow.ErrEOF
}

func (*NodeSendForgotPasswordCode) ReactTo(ctx context.Context, deps *workflow.Dependencies, w *workflow.Workflow, input workflow.Input) (*workflow.Node, error) {
	return nil, workflow.ErrIncompatibleInput
}

func (*NodeSendForgotPasswordCode) OutputData(ctx context.Context, deps *workflow.Dependencies, w *workflow.Workflow) (interface{}, error) {
	return nil, nil
}

func (n *NodeSendForgotPasswordCode) sendCode(ctx context.Context, deps *workflow.Dependencies, w *workflow.Workflow) error {
	return deps.ForgotPassword.SendCode(n.LoginID)

}
