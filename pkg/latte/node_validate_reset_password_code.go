package latte

import (
	"context"

	"github.com/authgear/authgear-server/pkg/lib/workflow"
)

func init() {
	workflow.RegisterNode(&NodeValidateResetPasswordCode{})
}

type NodeValidateResetPasswordCode struct {
	Code string `json:"code"`
}

func (n *NodeValidateResetPasswordCode) Kind() string {
	return "latte.NodeValidateResetPasswordCode"
}

func (n *NodeValidateResetPasswordCode) GetEffects(ctx context.Context, deps *workflow.Dependencies, w *workflow.Workflow) (effs []workflow.Effect, err error) {
	return []workflow.Effect{
		workflow.RunEffect(func(ctx context.Context, deps *workflow.Dependencies) error {
			return n.validate(deps)
		}),
	}, nil
}

func (*NodeValidateResetPasswordCode) CanReactTo(ctx context.Context, deps *workflow.Dependencies, w *workflow.Workflow) ([]workflow.Input, error) {
	return nil, workflow.ErrEOF
}

func (*NodeValidateResetPasswordCode) ReactTo(ctx context.Context, deps *workflow.Dependencies, w *workflow.Workflow, input workflow.Input) (*workflow.Node, error) {
	return nil, workflow.ErrIncompatibleInput
}

func (n *NodeValidateResetPasswordCode) OutputData(ctx context.Context, deps *workflow.Dependencies, w *workflow.Workflow) (interface{}, error) {
	return nil, nil
}

func (n *NodeValidateResetPasswordCode) validate(deps *workflow.Dependencies) error {
	err := deps.ResetPassword.CheckResetPasswordCode(n.Code)
	return err
}
