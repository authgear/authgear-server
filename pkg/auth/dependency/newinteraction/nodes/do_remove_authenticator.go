package nodes

import (
	"github.com/authgear/authgear-server/pkg/auth/dependency/newinteraction"
	"github.com/authgear/authgear-server/pkg/lib/authn/authenticator"
)

func init() {
	newinteraction.RegisterNode(&NodeDoRemoveAuthenticator{})
}

type EdgeDoRemoveAuthenticator struct {
	Authenticators []*authenticator.Info
}

func (e *EdgeDoRemoveAuthenticator) Instantiate(ctx *newinteraction.Context, graph *newinteraction.Graph, rawInput interface{}) (newinteraction.Node, error) {
	return &NodeDoRemoveAuthenticator{
		Authenticators: e.Authenticators,
	}, nil
}

type NodeDoRemoveAuthenticator struct {
	Authenticators []*authenticator.Info `json:"authenticators"`
}

func (n *NodeDoRemoveAuthenticator) Prepare(ctx *newinteraction.Context, graph *newinteraction.Graph) error {
	return nil
}

func (n *NodeDoRemoveAuthenticator) Apply(perform func(eff newinteraction.Effect) error, graph *newinteraction.Graph) error {
	err := perform(newinteraction.EffectRun(func(ctx *newinteraction.Context) error {
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

func (n *NodeDoRemoveAuthenticator) DeriveEdges(graph *newinteraction.Graph) ([]newinteraction.Edge, error) {
	return graph.Intent.DeriveEdgesForNode(graph, n)
}
