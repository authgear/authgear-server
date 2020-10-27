package nodes

import (
	"github.com/authgear/authgear-server/pkg/lib/authn/authenticator"
	"github.com/authgear/authgear-server/pkg/lib/interaction"
)

func init() {
	interaction.RegisterNode(&NodeDoUpdateAuthenticator{})
}

type EdgeDoUpdateAuthenticator struct {
	Stage                     interaction.AuthenticationStage
	AuthenticatorBeforeUpdate *authenticator.Info
	AuthenticatorAfterUpdate  *authenticator.Info
}

func (e *EdgeDoUpdateAuthenticator) Instantiate(ctx *interaction.Context, graph *interaction.Graph, rawInput interface{}) (interaction.Node, error) {
	return &NodeDoUpdateAuthenticator{
		AuthenticatorBeforeUpdate: e.AuthenticatorBeforeUpdate,
		AuthenticatorAfterUpdate:  e.AuthenticatorAfterUpdate,
	}, nil
}

type NodeDoUpdateAuthenticator struct {
	Stage                     interaction.AuthenticationStage `json:"stage"`
	AuthenticatorBeforeUpdate *authenticator.Info             `json:"authenticator_before_update"`
	AuthenticatorAfterUpdate  *authenticator.Info             `json:"authenticator_after_update"`
}

func (n *NodeDoUpdateAuthenticator) Prepare(ctx *interaction.Context, graph *interaction.Graph) error {
	return nil
}

func (n *NodeDoUpdateAuthenticator) GetEffects() ([]interaction.Effect, error) {
	return []interaction.Effect{
		interaction.EffectRun(func(ctx *interaction.Context, graph *interaction.Graph, nodeIndex int) error {
			return ctx.Authenticators.Update(n.AuthenticatorAfterUpdate)
		}),
	}, nil
}

func (n *NodeDoUpdateAuthenticator) DeriveEdges(graph *interaction.Graph) ([]interaction.Edge, error) {
	return graph.Intent.DeriveEdgesForNode(graph, n)
}

func (n *NodeDoUpdateAuthenticator) UserAuthenticator(stage interaction.AuthenticationStage) (*authenticator.Info, bool) {
	if n.Stage == stage && n.AuthenticatorAfterUpdate != nil {
		return n.AuthenticatorAfterUpdate, true
	}
	return nil, false
}
