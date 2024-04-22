package declarative

import (
	"context"

	authflow "github.com/authgear/authgear-server/pkg/lib/authenticationflow"
	"github.com/authgear/authgear-server/pkg/lib/authn/authenticator"
	"github.com/authgear/authgear-server/pkg/lib/authn/identity"
)

func init() {
	authflow.RegisterNode(&NodeDoCreatePasskey{})
}

type NodeDoCreatePasskey struct {
	SkipCreate          bool                `json:"skip_create,omitempty"`
	Identity            *identity.Info      `json:"identity,omitempty"`
	Authenticator       *authenticator.Info `json:"authenticator,omitempty"`
	AttestationResponse []byte              `json:"attestation_response,omitempty"`
}

var _ authflow.NodeSimple = &NodeDoCreatePasskey{}
var _ authflow.EffectGetter = &NodeDoCreatePasskey{}
var _ authflow.Milestone = &NodeDoCreatePasskey{}
var _ MilestoneDoCreateIdentity = &NodeDoCreatePasskey{}
var _ MilestoneDoCreateAuthenticator = &NodeDoCreatePasskey{}
var _ MilestoneDoCreatePasskey = &NodeDoCreatePasskey{}

func (n *NodeDoCreatePasskey) Kind() string {
	return "NodeDoCreatePasskey"
}

func (*NodeDoCreatePasskey) Milestone() {}
func (n *NodeDoCreatePasskey) MilestoneDoCreateIdentity() *identity.Info {
	return n.Identity
}
func (n *NodeDoCreatePasskey) MilestoneDoCreateAuthenticator() *authenticator.Info {
	return n.Authenticator
}
func (n *NodeDoCreatePasskey) MilestoneDoCreateAuthenticatorSkipCreate() {
	n.SkipCreate = true
}
func (n *NodeDoCreatePasskey) MilestoneDoCreateAuthenticatorUpdate(newInfo *authenticator.Info) {
	panic("NodeDoCreatePasskey does not support update authenticator")
}
func (n *NodeDoCreatePasskey) MilestoneDoCreateIdentitySkipCreate() {
	n.SkipCreate = true
}
func (n *NodeDoCreatePasskey) MilestoneDoCreateIdentityUpdate(newInfo *identity.Info) {
	panic("NodeDoCreatePasskey does not support update identity")
}
func (n *NodeDoCreatePasskey) MilestoneDoCreatePasskeyUpdateUserID(userID string) {
	n.Identity = n.Identity.UpdateUserID(userID)
	n.Authenticator = n.Authenticator.UpdateUserID(userID)
}

func (n *NodeDoCreatePasskey) GetEffects(ctx context.Context, deps *authflow.Dependencies, flows authflow.Flows) (effs []authflow.Effect, err error) {
	if n.SkipCreate {
		return nil, nil
	}

	return []authflow.Effect{
		authflow.RunEffect(func(ctx context.Context, deps *authflow.Dependencies) error {
			err := deps.Identities.Create(n.Identity)
			if err != nil {
				return err
			}

			return nil
		}),
		authflow.RunEffect(func(ctx context.Context, deps *authflow.Dependencies) error {
			return deps.Authenticators.Create(n.Authenticator, false)
		}),
		authflow.OnCommitEffect(func(ctx context.Context, deps *authflow.Dependencies) error {
			return deps.PasskeyService.ConsumeAttestationResponse(n.AttestationResponse)
		}),
	}, nil
}
