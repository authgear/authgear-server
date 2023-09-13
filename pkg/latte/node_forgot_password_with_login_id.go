package latte

import (
	"context"

	"github.com/authgear/authgear-server/pkg/lib/workflow"
)

func init() {
	workflow.RegisterNode(&NodeForgotPasswordWithLoginID{})
}

type NodeForgotPasswordWithLoginID struct {
	LoginID string `json:"login_id"`
}

func (n *NodeForgotPasswordWithLoginID) Kind() string {
	return "latte.NodeForgotPasswordWithLoginID"
}

func (n *NodeForgotPasswordWithLoginID) GetEffects(ctx context.Context, deps *workflow.Dependencies, workflows workflow.Workflows) (effs []workflow.Effect, err error) {
	return nil, nil
}

func (*NodeForgotPasswordWithLoginID) CanReactTo(ctx context.Context, deps *workflow.Dependencies, workflows workflow.Workflows) ([]workflow.Input, error) {
	return nil, workflow.ErrEOF
}

func (*NodeForgotPasswordWithLoginID) ReactTo(ctx context.Context, deps *workflow.Dependencies, workflows workflow.Workflows, input workflow.Input) (*workflow.Node, error) {
	return nil, workflow.ErrIncompatibleInput
}

func (n *NodeForgotPasswordWithLoginID) OutputData(ctx context.Context, deps *workflow.Dependencies, workflows workflow.Workflows) (interface{}, error) {
	output := map[string]interface{}{}
	userIDHint, err := reauthUserIDHint(ctx)

	if err == nil && userIDHint != "" {
		output["login_id"] = n.LoginID
	}

	return output, nil
}
