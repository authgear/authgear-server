package nodes

import (
	"github.com/authgear/authgear-server/pkg/lib/authn/authenticator"
	"github.com/authgear/authgear-server/pkg/lib/authn/identity"
	"github.com/authgear/authgear-server/pkg/lib/interaction"
)

func init() {
	interaction.RegisterNode(&NodeRemoveAuthenticator{})
}

type EdgeRemoveAuthenticator struct {
	// Current we only support removing the matching authenticators of the given identity info.
	IdentityInfo *identity.Info
}

func (e *EdgeRemoveAuthenticator) Instantiate(ctx *interaction.Context, graph *interaction.Graph, rawInput interface{}) (interaction.Node, error) {
	userID := graph.MustGetUserID()
	ais, err := ctx.Authenticators.List(
		userID,
		authenticator.KeepMatchingAuthenticatorOfIdentity(e.IdentityInfo),
	)
	if err != nil {
		return nil, err
	}

	return &NodeRemoveAuthenticator{
		Authenticators: ais,
	}, nil
}

type NodeRemoveAuthenticator struct {
	Authenticators []*authenticator.Info `json:"authenticators"`
}

func (n *NodeRemoveAuthenticator) Prepare(ctx *interaction.Context, graph *interaction.Graph) error {
	return nil
}

func (n *NodeRemoveAuthenticator) Apply(perform func(eff interaction.Effect) error, graph *interaction.Graph) error {
	return nil
}

func (n *NodeRemoveAuthenticator) DeriveEdges(graph *interaction.Graph) ([]interaction.Edge, error) {
	return graph.Intent.DeriveEdgesForNode(graph, n)
}
