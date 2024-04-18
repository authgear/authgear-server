package declarative

import (
	"context"

	authflow "github.com/authgear/authgear-server/pkg/lib/authenticationflow"
	"github.com/authgear/authgear-server/pkg/lib/authn/identity"
)

func init() {
	authflow.RegisterNode(&NodeDoCreateIdentity{})
}

type NodeDoCreateIdentity struct {
	Identity *identity.Info `json:"identity,omitempty"`
}

var _ authflow.NodeSimple = &NodeDoCreateIdentity{}
var _ authflow.Milestone = &NodeDoCreateIdentity{}
var _ MilestoneDoCreateIdentity = &NodeDoCreateIdentity{}
var _ authflow.EffectGetter = &NodeDoCreateIdentity{}
var _ MilestoneSwitchToExistingUser = &NodeDoCreateIdentity{}

func (i *NodeDoCreateIdentity) MilestoneSwitchToExistingUser(newUserID string) {
	i.Identity = i.Identity.UpdateUserID(newUserID)
}

func (n *NodeDoCreateIdentity) Kind() string {
	return "NodeDoCreateIdentity"
}

func (*NodeDoCreateIdentity) Milestone() {}
func (n *NodeDoCreateIdentity) MilestoneDoCreateIdentity() *identity.Info {
	return n.Identity
}

func (n *NodeDoCreateIdentity) GetEffects(ctx context.Context, deps *authflow.Dependencies, flows authflow.Flows) (effs []authflow.Effect, err error) {
	return []authflow.Effect{
		authflow.RunEffect(func(ctx context.Context, deps *authflow.Dependencies) error {
			err := deps.Identities.Create(n.Identity)
			if err != nil {
				return err
			}

			return nil
		}),
	}, nil
}
