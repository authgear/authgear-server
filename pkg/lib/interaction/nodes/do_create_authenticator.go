package nodes

import (
	"github.com/authgear/authgear-server/pkg/lib/authn"
	"github.com/authgear/authgear-server/pkg/lib/authn/authenticator"
	"github.com/authgear/authgear-server/pkg/lib/interaction"
)

func init() {
	interaction.RegisterNode(&NodeDoCreateAuthenticator{})
}

type EdgeDoCreateAuthenticator struct {
	Stage          authn.AuthenticationStage
	Authenticators []*authenticator.Info
}

func (e *EdgeDoCreateAuthenticator) Instantiate(ctx *interaction.Context, graph *interaction.Graph, rawInput interface{}) (interaction.Node, error) {
	return &NodeDoCreateAuthenticator{
		Stage:          e.Stage,
		Authenticators: e.Authenticators,
	}, nil
}

type NodeDoCreateAuthenticator struct {
	Stage          authn.AuthenticationStage `json:"stage"`
	Authenticators []*authenticator.Info     `json:"authenticators"`
}

func (n *NodeDoCreateAuthenticator) Prepare(ctx *interaction.Context, graph *interaction.Graph) error {
	return nil
}

func (n *NodeDoCreateAuthenticator) GetEffects() ([]interaction.Effect, error) {
	return []interaction.Effect{
		interaction.EffectRun(func(ctx *interaction.Context, graph *interaction.Graph, nodeIndex int) error {
			for _, a := range n.Authenticators {
				err := ctx.Authenticators.Create(a)
				if err != nil {
					return err
				}
			}

			return nil
		}),
	}, nil
}

func (n *NodeDoCreateAuthenticator) DeriveEdges(graph *interaction.Graph) ([]interaction.Edge, error) {
	return graph.Intent.DeriveEdgesForNode(graph, n)
}

func (n *NodeDoCreateAuthenticator) UserAuthenticator(stage authn.AuthenticationStage) (*authenticator.Info, bool) {
	if len(n.Authenticators) > 1 {
		panic("interaction: expect at most one primary/secondary authenticator")
	}
	if len(n.Authenticators) == 0 {
		return nil, false
	}
	if n.Stage == stage && n.Authenticators[0] != nil {
		return n.Authenticators[0], true
	}
	return nil, false
}

func (n *NodeDoCreateAuthenticator) UserNewAuthenticators() []*authenticator.Info {
	return n.Authenticators
}
