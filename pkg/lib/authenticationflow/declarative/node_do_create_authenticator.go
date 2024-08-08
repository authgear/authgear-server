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
	SkipCreate    bool                `json:"skip_create,omitempty"`
	Authenticator *authenticator.Info `json:"authenticator,omitempty"`
}

var _ authflow.NodeSimple = &NodeDoCreateAuthenticator{}
var _ authflow.Milestone = &NodeDoCreateAuthenticator{}
var _ MilestoneDoCreateAuthenticator = &NodeDoCreateAuthenticator{}
var _ authflow.EffectGetter = &NodeDoCreateAuthenticator{}

func (n *NodeDoCreateAuthenticator) Kind() string {
	return "NodeDoCreateAuthenticator"
}

func (n *NodeDoCreateAuthenticator) Milestone() {}
func (n *NodeDoCreateAuthenticator) MilestoneDoCreateAuthenticator() *authenticator.Info {
	return n.Authenticator
}
func (n *NodeDoCreateAuthenticator) MilestoneDoCreateAuthenticatorSkipCreate() {
	n.SkipCreate = true
}
func (n *NodeDoCreateAuthenticator) MilestoneDoCreateAuthenticatorUpdate(newInfo *authenticator.Info) {
	n.Authenticator = newInfo
}

func (n *NodeDoCreateAuthenticator) GetEffects(ctx context.Context, deps *authflow.Dependencies, flows authflow.Flows) (effs []authflow.Effect, err error) {
	if n.SkipCreate {
		return nil, nil
	}

	return []authflow.Effect{
		authflow.RunEffect(func(ctx context.Context, deps *authflow.Dependencies) error {
			return deps.Authenticators.Create(n.Authenticator, false)
		}),
		authflow.OnCommitEffect(func(ctx context.Context, deps *authflow.Dependencies) error {
			if n.Authenticator.Kind != authenticator.KindSecondary {
				return nil
			}
			return deps.Users.UpdateMFAEnrollment(n.Authenticator.UserID, nil)
		}),
	}, nil
}
