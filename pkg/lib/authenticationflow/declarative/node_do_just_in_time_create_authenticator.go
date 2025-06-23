package declarative

import (
	"context"

	"github.com/authgear/authgear-server/pkg/api/model"
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
func (n *NodeDoJustInTimeCreateAuthenticator) MilestoneDoCreateAuthenticator() (*authenticator.Info, bool) {
	return n.Authenticator, !n.SkipCreate
}
func (n *NodeDoJustInTimeCreateAuthenticator) MilestoneDoCreateAuthenticatorSkipCreate() {
	n.SkipCreate = true
}
func (n *NodeDoJustInTimeCreateAuthenticator) MilestoneDoCreateAuthenticatorAuthentication() (*model.Authentication, bool) {
	if n.Authenticator == nil || n.SkipCreate {
		return nil, false
	}
	authn := n.Authenticator.ToAuthentication()
	authnModel := n.Authenticator.ToModel()
	return &model.Authentication{
		Authentication: authn,
		Authenticator:  &authnModel,
	}, true
}
func (n *NodeDoJustInTimeCreateAuthenticator) MilestoneDoCreateAuthenticatorUpdate(newInfo *authenticator.Info) {
	n.Authenticator = newInfo
}

func (n *NodeDoJustInTimeCreateAuthenticator) GetEffects(ctx context.Context, deps *authflow.Dependencies, flows authflow.Flows) (effs []authflow.Effect, err error) {
	if n.SkipCreate {
		return nil, nil
	}

	return []authflow.Effect{
		authflow.RunEffect(func(ctx context.Context, deps *authflow.Dependencies) error {
			return deps.Authenticators.Create(ctx, n.Authenticator, false)
		}),
		authflow.OnCommitEffect(func(ctx context.Context, deps *authflow.Dependencies) error {
			if n.Authenticator.Kind != authenticator.KindSecondary {
				return nil
			}
			return deps.Users.UpdateMFAEnrollment(ctx, n.Authenticator.UserID, nil)
		}),
	}, nil
}
