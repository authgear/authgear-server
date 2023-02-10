package latte

import (
	"context"

	"github.com/authgear/authgear-server/pkg/lib/workflow"
)

func init() {
	workflow.RegisterNode(&NodeVerifiedLoginLink{})
}

type NodeVerifiedLoginLink struct {
	Code string `json:"code"`
}

func (n *NodeVerifiedLoginLink) Kind() string {
	return "latte.NodeVerifiedLoginLink"
}

func (n *NodeVerifiedLoginLink) GetEffects(ctx context.Context, deps *workflow.Dependencies, w *workflow.Workflow) (effs []workflow.Effect, err error) {
	return []workflow.Effect{
		workflow.OnCommitEffect(func(ctx context.Context, deps *workflow.Dependencies) error {
			_, err := deps.OTPCodes.SetUserInputtedMagicLinkCode(n.Code)
			if err != nil {
				return err
			}

			// TODO(newman): Send websocket event for refreshing web session

			return nil
		}),
	}, nil
}

func (n *NodeVerifiedLoginLink) CanReactTo(ctx context.Context, deps *workflow.Dependencies, w *workflow.Workflow) ([]workflow.Input, error) {
	return []workflow.Input{
		&InputTakeLoginLinkCode{},
	}, nil
}

func (n *NodeVerifiedLoginLink) ReactTo(ctx context.Context, deps *workflow.Dependencies, w *workflow.Workflow, input workflow.Input) (*workflow.Node, error) {
	return nil, workflow.ErrEOF
}

func (n *NodeVerifiedLoginLink) OutputData(ctx context.Context, deps *workflow.Dependencies, w *workflow.Workflow) (interface{}, error) {
	return map[string]interface{}{}, nil
}
