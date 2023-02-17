package latte

import (
	"context"

	"github.com/authgear/authgear-server/pkg/lib/workflow"
)

func init() {
	workflow.RegisterNode(&NodeDoResetPasswordByCode{})
}

type NodeDoResetPasswordByCode struct {
	Code        string `json:"code"`
	NewPassword string `json:"new_password"`
}

func (n *NodeDoResetPasswordByCode) Kind() string {
	return "latte.NodeDoResetPasswordByCode"
}

func (n *NodeDoResetPasswordByCode) GetEffects(ctx context.Context, deps *workflow.Dependencies, w *workflow.Workflow) (effs []workflow.Effect, err error) {
	return []workflow.Effect{
		workflow.OnCommitEffect(func(ctx context.Context, deps *workflow.Dependencies) error {
			err := deps.ResetPassword.ResetPasswordByCode(n.Code, n.NewPassword)
			return err
		}),
	}, nil
}

func (*NodeDoResetPasswordByCode) CanReactTo(ctx context.Context, deps *workflow.Dependencies, w *workflow.Workflow) ([]workflow.Input, error) {
	return nil, workflow.ErrEOF
}

func (*NodeDoResetPasswordByCode) ReactTo(ctx context.Context, deps *workflow.Dependencies, w *workflow.Workflow, input workflow.Input) (*workflow.Node, error) {
	return nil, workflow.ErrIncompatibleInput
}

func (n *NodeDoResetPasswordByCode) OutputData(ctx context.Context, deps *workflow.Dependencies, w *workflow.Workflow) (interface{}, error) {
	return nil, nil
}
