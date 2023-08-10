package workflowconfig

import (
	"context"

	"github.com/authgear/authgear-server/pkg/lib/workflow"
)

func init() {
	workflow.RegisterNode(&NodeDoCheckAccountStatus{})
}

type NodeDoCheckAccountStatus struct {
	UserID string `json:"user_id"`
}

var _ workflow.NodeSimple = &NodeDoCheckAccountStatus{}

func (n *NodeDoCheckAccountStatus) Kind() string {
	return "workflowconfig.NodeDoCheckAccountStatus"
}

func (n *NodeDoCheckAccountStatus) GetEffects(ctx context.Context, deps *workflow.Dependencies, workflows workflow.Workflows) ([]workflow.Effect, error) {
	return []workflow.Effect{
		workflow.RunEffect(func(ctx context.Context, deps *workflow.Dependencies) error {
			u, err := deps.Users.GetRaw(n.UserID)
			if err != nil {
				return err
			}

			err = u.AccountStatus().Check()
			if err != nil {
				return err
			}

			return nil
		}),
	}, nil
}

func (*NodeDoCheckAccountStatus) CanReactTo(ctx context.Context, deps *workflow.Dependencies, workflows workflow.Workflows) ([]workflow.Input, error) {
	return nil, workflow.ErrEOF
}

func (*NodeDoCheckAccountStatus) ReactTo(ctx context.Context, deps *workflow.Dependencies, workflows workflow.Workflows, input workflow.Input) (*workflow.Node, error) {
	return nil, workflow.ErrIncompatibleInput
}

func (*NodeDoCheckAccountStatus) OutputData(ctx context.Context, deps *workflow.Dependencies, workflows workflow.Workflows) (interface{}, error) {
	return nil, nil
}
