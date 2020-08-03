package nodes

import (
	"github.com/authgear/authgear-server/pkg/auth/dependency/authenticator"
	"github.com/authgear/authgear-server/pkg/auth/dependency/identity"
	"github.com/authgear/authgear-server/pkg/auth/dependency/newinteraction"
)

func init() {
	newinteraction.RegisterNode(&NodeRemoveAuthenticator{})
}

type EdgeRemoveAuthenticator struct {
	// Current we only support removing the matching authenticators of the given identity info.
	IdentityInfo *identity.Info
}

func (e *EdgeRemoveAuthenticator) Instantiate(ctx *newinteraction.Context, graph *newinteraction.Graph, rawInput interface{}) (newinteraction.Node, error) {
	userID := graph.MustGetUserID()
	ais, err := ctx.Authenticators.ListAll(userID)
	if err != nil {
		return nil, err
	}

	ais = ctx.Authenticators.FilterMatchingAuthenticators(e.IdentityInfo, ais)

	return &NodeRemoveAuthenticator{
		Authenticators: ais,
	}, nil
}

type NodeRemoveAuthenticator struct {
	Authenticators []*authenticator.Info `json:"authenticators"`
}

func (n *NodeRemoveAuthenticator) Apply(perform func(eff newinteraction.Effect) error, graph *newinteraction.Graph) error {
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

func (n *NodeRemoveAuthenticator) DeriveEdges(ctx *newinteraction.Context, graph *newinteraction.Graph) ([]newinteraction.Edge, error) {
	return graph.Intent.DeriveEdgesForNode(ctx, graph, n)
}
