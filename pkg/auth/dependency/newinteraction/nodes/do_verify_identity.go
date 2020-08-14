package nodes

import (
	"github.com/authgear/authgear-server/pkg/auth/dependency/newinteraction"
	"github.com/authgear/authgear-server/pkg/lib/authn/authenticator"
	"github.com/authgear/authgear-server/pkg/lib/authn/identity"
)

func init() {
	newinteraction.RegisterNode(&NodeDoVerifyIdentity{})
}

type EdgeDoVerifyIdentity struct {
	Identity         *identity.Info
	NewAuthenticator *authenticator.Info
}

func (e *EdgeDoVerifyIdentity) Instantiate(ctx *newinteraction.Context, graph *newinteraction.Graph, rawInput interface{}) (newinteraction.Node, error) {
	return &NodeDoVerifyIdentity{
		Identity:         e.Identity,
		NewAuthenticator: e.NewAuthenticator,
	}, nil
}

type NodeDoVerifyIdentity struct {
	Identity         *identity.Info      `json:"identity"`
	NewAuthenticator *authenticator.Info `json:"new_authenticator"`
}

func (n *NodeDoVerifyIdentity) Prepare(ctx *newinteraction.Context, graph *newinteraction.Graph) error {
	return nil
}

func (n *NodeDoVerifyIdentity) Apply(perform func(eff newinteraction.Effect) error, graph *newinteraction.Graph) error {
	err := perform(newinteraction.EffectRun(func(ctx *newinteraction.Context) error {
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

func (n *NodeDoVerifyIdentity) DeriveEdges(graph *newinteraction.Graph) ([]newinteraction.Edge, error) {
	return graph.Intent.DeriveEdgesForNode(graph, n)
}

func (n *NodeDoVerifyIdentity) UserNewAuthenticators() []*authenticator.Info {
	if n.NewAuthenticator != nil {
		return []*authenticator.Info{n.NewAuthenticator}
	}
	return nil
}
