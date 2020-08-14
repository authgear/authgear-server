package nodes

import (
	"github.com/authgear/authgear-server/pkg/lib/authn/authenticator"
	"github.com/authgear/authgear-server/pkg/lib/interaction"
)

func init() {
	interaction.RegisterNode(&NodeDoRemoveAuthenticator{})
}

type EdgeDoRemoveAuthenticator struct {
	Authenticators []*authenticator.Info
}

func (e *EdgeDoRemoveAuthenticator) Instantiate(ctx *interaction.Context, graph *interaction.Graph, rawInput interface{}) (interaction.Node, error) {
	return &NodeDoRemoveAuthenticator{
		Authenticators: e.Authenticators,
	}, nil
}

type NodeDoRemoveAuthenticator struct {
	Authenticators []*authenticator.Info `json:"authenticators"`
}

func (n *NodeDoRemoveAuthenticator) Prepare(ctx *interaction.Context, graph *interaction.Graph) error {
	return nil
}

func (n *NodeDoRemoveAuthenticator) Apply(perform func(eff interaction.Effect) error, graph *interaction.Graph) error {
	err := perform(interaction.EffectRun(func(ctx *interaction.Context) error {
		for _, ai := range n.Authenticators {
			err := ctx.Authenticators.Delete(ai)
			if err != nil {
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

func (n *NodeDoRemoveAuthenticator) DeriveEdges(graph *interaction.Graph) ([]interaction.Edge, error) {
	return graph.Intent.DeriveEdgesForNode(graph, n)
}
