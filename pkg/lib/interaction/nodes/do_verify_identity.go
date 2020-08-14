package nodes

import (
	"github.com/authgear/authgear-server/pkg/lib/authn/authenticator"
	"github.com/authgear/authgear-server/pkg/lib/authn/identity"
	"github.com/authgear/authgear-server/pkg/lib/interaction"
)

func init() {
	interaction.RegisterNode(&NodeDoVerifyIdentity{})
}

type EdgeDoVerifyIdentity struct {
	Identity         *identity.Info
	NewAuthenticator *authenticator.Info
}

func (e *EdgeDoVerifyIdentity) Instantiate(ctx *interaction.Context, graph *interaction.Graph, rawInput interface{}) (interaction.Node, error) {
	return &NodeDoVerifyIdentity{
		Identity:         e.Identity,
		NewAuthenticator: e.NewAuthenticator,
	}, nil
}

type NodeDoVerifyIdentity struct {
	Identity         *identity.Info      `json:"identity"`
	NewAuthenticator *authenticator.Info `json:"new_authenticator"`
}

func (n *NodeDoVerifyIdentity) Prepare(ctx *interaction.Context, graph *interaction.Graph) error {
	return nil
}

func (n *NodeDoVerifyIdentity) Apply(perform func(eff interaction.Effect) error, graph *interaction.Graph) error {
	err := perform(interaction.EffectRun(func(ctx *interaction.Context) error {
		if n.NewAuthenticator != nil {
			if err := ctx.Authenticators.Create(n.NewAuthenticator); err != nil {
				return err
			}
		}

		return nil
	}))
	if err != nil {
		return err
	}

	return nil
}

func (n *NodeDoVerifyIdentity) DeriveEdges(graph *interaction.Graph) ([]interaction.Edge, error) {
	return graph.Intent.DeriveEdgesForNode(graph, n)
}

func (n *NodeDoVerifyIdentity) UserNewAuthenticators() []*authenticator.Info {
	if n.NewAuthenticator != nil {
		return []*authenticator.Info{n.NewAuthenticator}
	}
	return nil
}
