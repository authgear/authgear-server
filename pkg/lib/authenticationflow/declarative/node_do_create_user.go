package declarative

import (
	"context"

	authflow "github.com/authgear/authgear-server/pkg/lib/authenticationflow"
)

func init() {
	authflow.RegisterNode(&NodeDoCreateUser{})
}

type NodeDoCreateUser struct {
	UserID string `json:"user_id"`
}

var _ authflow.NodeSimple = &NodeDoCreateUser{}
var _ authflow.Milestone = &NodeDoCreateUser{}
var _ MilestoneDoUseUser = &NodeDoCreateUser{}
var _ MilestoneDoCreateUser = &NodeDoCreateUser{}
var _ authflow.EffectGetter = &NodeDoCreateUser{}

func (n *NodeDoCreateUser) Kind() string {
	return "NodeDoCreateUser"
}

func (*NodeDoCreateUser) Milestone()                      {}
func (n *NodeDoCreateUser) MilestoneDoUseUser() string    { return n.UserID }
func (n *NodeDoCreateUser) MilestoneDoCreateUser() string { return n.UserID }

func (n *NodeDoCreateUser) GetEffects(ctx context.Context, deps *authflow.Dependencies, flows authflow.Flows) (effs []authflow.Effect, err error) {
	return []authflow.Effect{
		authflow.RunEffect(func(ctx context.Context, deps *authflow.Dependencies) error {
			_, err := deps.Users.Create(n.UserID)
			return err
		}),
	}, nil
}
