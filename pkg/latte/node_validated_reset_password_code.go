package latte

import (
	"context"

	"github.com/authgear/authgear-server/pkg/lib/workflow"
)

func init() {
	workflow.RegisterNode(&NodeValidatedResetPasswordCode{})
}

type NodeValidatedResetPasswordCode struct {
	Code   string `json:"code"`
	UserID string `json:"user_id"`
}

func (n *NodeValidatedResetPasswordCode) Kind() string {
	return "latte.NodeValidatedResetPasswordCode"
}

func (n *NodeValidatedResetPasswordCode) GetEffects(ctx context.Context, deps *workflow.Dependencies, w *workflow.Workflow) (effs []workflow.Effect, err error) {
	return nil, nil
}

func (*NodeValidatedResetPasswordCode) CanReactTo(ctx context.Context, deps *workflow.Dependencies, w *workflow.Workflow) ([]workflow.Input, error) {
	return nil, workflow.ErrEOF
}

func (*NodeValidatedResetPasswordCode) ReactTo(ctx context.Context, deps *workflow.Dependencies, w *workflow.Workflow, input workflow.Input) (*workflow.Node, error) {
	return nil, workflow.ErrIncompatibleInput
}

func (n *NodeValidatedResetPasswordCode) OutputData(ctx context.Context, deps *workflow.Dependencies, w *workflow.Workflow) (interface{}, error) {
	type NodeValidatedResetPasswordCodeOutput struct {
		UserID string `json:"user_id"`
	}

	return &NodeValidatedResetPasswordCodeOutput{
		UserID: n.UserID,
	}, nil
}
