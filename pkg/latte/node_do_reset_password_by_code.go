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

func (n *NodeDoResetPasswordByCode) GetEffects(ctx context.Context, deps *workflow.Dependencies, workflows workflow.Workflows) (effs []workflow.Effect, err error) {
	return []workflow.Effect{
		workflow.OnCommitEffect(func(ctx context.Context, deps *workflow.Dependencies) error {
			err := deps.ResetPassword.ResetPasswordByEndUser(n.Code, n.NewPassword)
			return err
		}),
	}, nil
}

func (*NodeDoResetPasswordByCode) CanReactTo(ctx context.Context, deps *workflow.Dependencies, workflows workflow.Workflows) ([]workflow.Input, error) {
	return nil, workflow.ErrEOF
}

func (*NodeDoResetPasswordByCode) ReactTo(ctx context.Context, deps *workflow.Dependencies, workflows workflow.Workflows, input workflow.Input) (*workflow.Node, error) {
	return nil, workflow.ErrIncompatibleInput
}

func (n *NodeDoResetPasswordByCode) OutputData(ctx context.Context, deps *workflow.Dependencies, workflows workflow.Workflows) (interface{}, error) {
	return nil, nil
}
