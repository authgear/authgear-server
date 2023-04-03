package latte

import (
	"context"

	"github.com/authgear/authgear-server/pkg/lib/workflow"
)

func init() {
	workflow.RegisterNode(&NodeVerifiedCaptcha{})
}

type NodeVerifiedCaptcha struct {
}

func (n *NodeVerifiedCaptcha) Kind() string {
	return "latte.NodeVerifiedCaptcha"
}

func (n *NodeVerifiedCaptcha) GetEffects(ctx context.Context, deps *workflow.Dependencies, w *workflow.Workflow) (effs []workflow.Effect, err error) {
	return nil, nil
}

func (*NodeVerifiedCaptcha) CanReactTo(ctx context.Context, deps *workflow.Dependencies, w *workflow.Workflow) ([]workflow.Input, error) {
	return nil, workflow.ErrEOF
}

func (*NodeVerifiedCaptcha) ReactTo(ctx context.Context, deps *workflow.Dependencies, w *workflow.Workflow, input workflow.Input) (*workflow.Node, error) {
	return nil, workflow.ErrIncompatibleInput
}

func (*NodeVerifiedCaptcha) OutputData(ctx context.Context, deps *workflow.Dependencies, w *workflow.Workflow) (interface{}, error) {
	return nil, nil
}
