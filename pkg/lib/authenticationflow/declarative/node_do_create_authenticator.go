package declarative

import (
	"context"

	authflow "github.com/authgear/authgear-server/pkg/lib/authenticationflow"
	"github.com/authgear/authgear-server/pkg/lib/authn/authenticator"
)

func init() {
	authflow.RegisterNode(&NodeDoCreateAuthenticator{})
}

type NodeDoCreateAuthenticator struct {
	Authenticator *authenticator.Info `json:"authenticator,omitempty"`
}

var _ authflow.NodeSimple = &NodeDoCreateAuthenticator{}
var _ authflow.Milestone = &NodeDoCreateAuthenticator{}
var _ MilestoneDoCreateAuthenticator = &NodeDoCreateAuthenticator{}
var _ MilestoneSwitchToExistingUser = &NodeDoCreateAuthenticator{}
var _ authflow.EffectGetter = &NodeDoCreateAuthenticator{}

func (n *NodeDoCreateAuthenticator) Kind() string {
	return "NodeDoCreateAuthenticator"
}

func (n *NodeDoCreateAuthenticator) Milestone() {}
func (n *NodeDoCreateAuthenticator) MilestoneDoCreateAuthenticator() *authenticator.Info {
	return n.Authenticator
}
func (i *NodeDoCreateAuthenticator) MilestoneSwitchToExistingUser(newUserID string) {
	// TODO(tung): Skip creation if already have one
	i.Authenticator = i.Authenticator.UpdateUserID(newUserID)
}

func (n *NodeDoCreateAuthenticator) GetEffects(ctx context.Context, deps *authflow.Dependencies, flows authflow.Flows) (effs []authflow.Effect, err error) {
	return []authflow.Effect{
		authflow.RunEffect(func(ctx context.Context, deps *authflow.Dependencies) error {
			return deps.Authenticators.Create(n.Authenticator, false)
		}),
	}, nil
}
