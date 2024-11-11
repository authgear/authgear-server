package declarative

import (
	"context"

	authflow "github.com/authgear/authgear-server/pkg/lib/authenticationflow"
	"github.com/authgear/authgear-server/pkg/lib/authn/authenticator"
)

func init() {
	authflow.RegisterNode(&NodeDoUpdateAuthenticator{})
}

type NodeDoUpdateAuthenticator struct {
	Authenticator *authenticator.Info `json:"authenticator,omitempty"`
}

var _ authflow.NodeSimple = &NodeDoUpdateAuthenticator{}
var _ authflow.EffectGetter = &NodeDoUpdateAuthenticator{}

func (*NodeDoUpdateAuthenticator) Kind() string {
	return "NodeDoUpdateAuthenticator"
}

func (n *NodeDoUpdateAuthenticator) GetEffects(ctx context.Context, deps *authflow.Dependencies, flows authflow.Flows) ([]authflow.Effect, error) {
	return []authflow.Effect{
		authflow.RunEffect(func(ctx context.Context, deps *authflow.Dependencies) error {
			return deps.Authenticators.Update(ctx, n.Authenticator)
		}),
	}, nil
}
