package declarative

import (
	"context"

	authflow "github.com/authgear/authgear-server/pkg/lib/authenticationflow"
)

func init() {
	authflow.RegisterNode(&NodeDoCreateUser{})
}

type NodeDoCreateUser struct {
	UserID       string `json:"user_id"`
	SkipCreation bool   `json:"skip_creation,omitempty"`
}

var _ authflow.NodeSimple = &NodeDoCreateUser{}
var _ authflow.Milestone = &NodeDoCreateUser{}
var _ MilestoneDoUseUser = &NodeDoCreateUser{}
var _ MilestoneDoCreateUser = &NodeDoCreateUser{}
var _ authflow.EffectGetter = &NodeDoCreateUser{}
var _ MilestoneSwitchToExistingUser = &NodeDoCreateUser{}

func (n *NodeDoCreateUser) Kind() string {
	return "NodeDoCreateUser"
}

func (*NodeDoCreateUser) Milestone()                   {}
func (n *NodeDoCreateUser) MilestoneDoUseUser() string { return n.UserID }
func (n *NodeDoCreateUser) MilestoneDoCreateUser() string {
	if n.SkipCreation {
		return ""
	}
	return n.UserID
}
func (n *NodeDoCreateUser) MilestoneSwitchToExistingUser(newUserID string) {
	n.UserID = newUserID
	// MilestoneSwitchToExistingUser is used in cases that the flow wants to update an existing user
	// instead of creating new user, so set SkipCreation to true
	n.SkipCreation = true
}

func (n *NodeDoCreateUser) GetEffects(ctx context.Context, deps *authflow.Dependencies, flows authflow.Flows) (effs []authflow.Effect, err error) {
	return []authflow.Effect{
		authflow.RunEffect(func(ctx context.Context, deps *authflow.Dependencies) error {
			if n.SkipCreation {
				return nil
			}
			_, err := deps.Users.Create(n.UserID)
			return err
		}),
	}, nil
}
