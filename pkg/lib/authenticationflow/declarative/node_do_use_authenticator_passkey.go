package declarative

import (
	"context"

	authflow "github.com/authgear/authgear-server/pkg/lib/authenticationflow"
	"github.com/authgear/authgear-server/pkg/lib/authn/authenticator"
)

func init() {
	authflow.RegisterNode(&NodeDoUseAuthenticatorPasskey{})
}

type NodeDoUseAuthenticatorPasskey struct {
	AssertionResponse []byte              `json:"assertion_response,omitempty"`
	Authenticator     *authenticator.Info `json:"authenticator,omitempty"`
	RequireUpdate     bool                `json:"require_update,omitempty"`
}

var _ authflow.NodeSimple = &NodeDoUseAuthenticatorPasskey{}
var _ authflow.EffectGetter = &NodeDoUseAuthenticatorPasskey{}
var _ authflow.Milestone = &NodeDoUseAuthenticatorPasskey{}
var _ MilestoneDidSelectAuthenticator = &NodeDoUseAuthenticatorPasskey{}
var _ MilestoneDidAuthenticate = &NodeDoUseAuthenticatorPasskey{}

func (*NodeDoUseAuthenticatorPasskey) Kind() string {
	return "NodeDoUseAuthenticatorPasskey"
}

func (n *NodeDoUseAuthenticatorPasskey) GetEffects(ctx context.Context, deps *authflow.Dependencies, flows authflow.Flows) ([]authflow.Effect, error) {
	return []authflow.Effect{
		authflow.RunEffect(func(ctx context.Context, deps *authflow.Dependencies) error {
			if n.RequireUpdate {
				return deps.Authenticators.Update(ctx, n.Authenticator)
			}
			return nil
		}),
		authflow.OnCommitEffect(func(ctx context.Context, deps *authflow.Dependencies) error {
			return deps.PasskeyService.ConsumeAssertionResponse(ctx, n.AssertionResponse)
		}),
	}, nil
}

func (*NodeDoUseAuthenticatorPasskey) Milestone() {}
func (n *NodeDoUseAuthenticatorPasskey) MilestoneDidSelectAuthenticator() *authenticator.Info {
	return n.Authenticator
}
func (n *NodeDoUseAuthenticatorPasskey) MilestoneDidAuthenticate() (amr []string) {
	return n.Authenticator.AMR()
}
