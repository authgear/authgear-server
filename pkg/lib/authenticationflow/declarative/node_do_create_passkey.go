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
var _ MilestoneSwitchToExistingUser = &NodeDoCreatePasskey{}

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
func (n *NodeDoCreatePasskey) MilestoneDoCreateIdentitySkipCreate() {
	n.SkipCreate = true
}

func (i *NodeDoCreatePasskey) MilestoneSwitchToExistingUser(deps *authflow.Dependencies, flow *authflow.Flow, newUserID string) error {
	// One user could have multiple passkey, so just create it for the existing user
	i.Identity = i.Identity.UpdateUserID(newUserID)
	i.Authenticator = i.Authenticator.UpdateUserID(newUserID)
	return nil
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
