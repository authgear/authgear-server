package workflowconfig

import (
	"context"

	workflow "github.com/authgear/authgear-server/pkg/lib/workflow2"
)

func init() {
	workflow.RegisterNode(&NodeDoCheckAccountStatus{})
}

type NodeDoCheckAccountStatus struct {
	UserID string `json:"user_id"`
}

var _ workflow.NodeSimple = &NodeDoCheckAccountStatus{}
var _ workflow.EffectGetter = &NodeDoCheckAccountStatus{}

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
