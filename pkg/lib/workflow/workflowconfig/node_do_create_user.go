package workflowconfig

import (
	"context"

	"github.com/authgear/authgear-server/pkg/lib/workflow"
)

func init() {
	workflow.RegisterNode(&NodeDoCreateUser{})
}

type NodeDoCreateUser struct {
	UserID string `json:"user_id"`
}

var _ workflow.NodeSimple = &NodeDoCreateUser{}

var _ MilestoneDoCreateUser = &NodeDoCreateUser{}

func (*NodeDoCreateUser) Milestone()                      {}
func (n *NodeDoCreateUser) MilestoneDoCreateUser() string { return n.UserID }

func (n *NodeDoCreateUser) Kind() string {
	return "workflowconfig.NodeDoCreateUser"
}

func (n *NodeDoCreateUser) GetEffects(ctx context.Context, deps *workflow.Dependencies, workflows workflow.Workflows) (effs []workflow.Effect, err error) {
	return []workflow.Effect{
		workflow.RunEffect(func(ctx context.Context, deps *workflow.Dependencies) error {
			_, err := deps.Users.Create(n.UserID)
			return err
		}),
	}, nil
}

func (*NodeDoCreateUser) CanReactTo(ctx context.Context, deps *workflow.Dependencies, workflows workflow.Workflows) ([]workflow.Input, error) {
	return nil, workflow.ErrEOF
}

func (*NodeDoCreateUser) ReactTo(ctx context.Context, deps *workflow.Dependencies, workflows workflow.Workflows, input workflow.Input) (*workflow.Node, error) {
	return nil, workflow.ErrIncompatibleInput
}

func (n *NodeDoCreateUser) OutputData(ctx context.Context, deps *workflow.Dependencies, workflows workflow.Workflows) (interface{}, error) {
	return nil, nil
}
