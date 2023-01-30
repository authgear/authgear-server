package nodes

import (
	"github.com/authgear/authgear-server/pkg/lib/authn"
	"github.com/authgear/authgear-server/pkg/lib/authn/authenticator"
	"github.com/authgear/authgear-server/pkg/lib/interaction"
)

func init() {
	interaction.RegisterNode(&NodeCreateAuthenticatorEnd{})
}

type EdgeCreateAuthenticatorEnd struct {
	Stage          authn.AuthenticationStage
	Authenticators []*authenticator.Info
	DeferVerify    bool
}

func (e *EdgeCreateAuthenticatorEnd) Instantiate(ctx *interaction.Context, graph *interaction.Graph, input interface{}) (interaction.Node, error) {
	return &NodeCreateAuthenticatorEnd{
		Stage:          e.Stage,
		Authenticators: e.Authenticators,
		DeferVerify:    e.DeferVerify,
	}, nil
}

type NodeCreateAuthenticatorEnd struct {
	Stage          authn.AuthenticationStage `json:"stage"`
	Authenticators []*authenticator.Info     `json:"authenticators"`
	DeferVerify    bool                      `json:"defer_verify"`
}

func (n *NodeCreateAuthenticatorEnd) Prepare(ctx *interaction.Context, graph *interaction.Graph) error {
	return nil
}

func (n *NodeCreateAuthenticatorEnd) GetEffects() ([]interaction.Effect, error) {
	return nil, nil
}

func (n *NodeCreateAuthenticatorEnd) DeriveEdges(graph *interaction.Graph) ([]interaction.Edge, error) {
	return graph.Intent.DeriveEdgesForNode(graph, n)
}
