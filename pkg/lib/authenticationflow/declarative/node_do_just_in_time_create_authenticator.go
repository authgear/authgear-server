package declarative

import (
	"context"

	authflow "github.com/authgear/authgear-server/pkg/lib/authenticationflow"
	"github.com/authgear/authgear-server/pkg/lib/authn/authenticator"
)

func init() {
	authflow.RegisterNode(&NodeDoJustInTimeCreateAuthenticator{})
}

type NodeDoJustInTimeCreateAuthenticator struct {
	SkipCreate    bool                `json:"skip_create,omitempty"`
	Authenticator *authenticator.Info `json:"authenticator,omitempty"`
}

var _ authflow.NodeSimple = &NodeDoJustInTimeCreateAuthenticator{}
var _ authflow.Milestone = &NodeDoJustInTimeCreateAuthenticator{}
var _ MilestoneDoCreateAuthenticator = &NodeDoJustInTimeCreateAuthenticator{}
var _ MilestoneDidSelectAuthenticator = &NodeDoJustInTimeCreateAuthenticator{}
var _ authflow.EffectGetter = &NodeDoJustInTimeCreateAuthenticator{}

func (*NodeDoJustInTimeCreateAuthenticator) Kind() string {
	return "NodeDoJustInTimeCreateAuthenticator"
}

func (n *NodeDoJustInTimeCreateAuthenticator) Milestone() {}
func (n *NodeDoJustInTimeCreateAuthenticator) MilestoneDidSelectAuthenticator() *authenticator.Info {
	return n.Authenticator
}
func (n *NodeDoJustInTimeCreateAuthenticator) MilestoneDoCreateAuthenticator() *authenticator.Info {
	return n.Authenticator
}
func (n *NodeDoJustInTimeCreateAuthenticator) MilestoneDoCreateAuthenticatorSkipCreate() {
	n.SkipCreate = true
}

func (n *NodeDoJustInTimeCreateAuthenticator) GetEffects(ctx context.Context, deps *authflow.Dependencies, flows authflow.Flows) (effs []authflow.Effect, err error) {
	if n.SkipCreate {
		return nil, nil
	}

	return []authflow.Effect{
		authflow.RunEffect(func(ctx context.Context, deps *authflow.Dependencies) error {
			return deps.Authenticators.Create(n.Authenticator, false)
		}),
	}, nil
}
