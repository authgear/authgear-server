package nodes

import (
	"context"

	"github.com/authgear/authgear-server/pkg/lib/authn"
	"github.com/authgear/authgear-server/pkg/lib/authn/authenticator"
	"github.com/authgear/authgear-server/pkg/lib/interaction"
)

func init() {
	interaction.RegisterNode(&NodeDoUpdateAuthenticator{})
}

type EdgeDoUpdateAuthenticator struct {
	Stage                     authn.AuthenticationStage
	AuthenticatorBeforeUpdate *authenticator.Info
	AuthenticatorAfterUpdate  *authenticator.Info
}

func (e *EdgeDoUpdateAuthenticator) Instantiate(goCtx context.Context, ctx *interaction.Context, graph *interaction.Graph, rawInput interface{}) (interaction.Node, error) {
	return &NodeDoUpdateAuthenticator{
		Stage:                     e.Stage,
		AuthenticatorBeforeUpdate: e.AuthenticatorBeforeUpdate,
		AuthenticatorAfterUpdate:  e.AuthenticatorAfterUpdate,
	}, nil
}

type NodeDoUpdateAuthenticator struct {
	Stage                     authn.AuthenticationStage `json:"stage"`
	AuthenticatorBeforeUpdate *authenticator.Info       `json:"authenticator_before_update"`
	AuthenticatorAfterUpdate  *authenticator.Info       `json:"authenticator_after_update"`
}

func (n *NodeDoUpdateAuthenticator) Prepare(goCtx context.Context, ctx *interaction.Context, graph *interaction.Graph) error {
	return nil
}

func (n *NodeDoUpdateAuthenticator) GetEffects(goCtx context.Context) ([]interaction.Effect, error) {
	return []interaction.Effect{
		interaction.EffectRun(func(goCtx context.Context, ctx *interaction.Context, graph *interaction.Graph, nodeIndex int) error {
			return ctx.Authenticators.Update(goCtx, n.AuthenticatorAfterUpdate)
		}),
	}, nil
}

func (n *NodeDoUpdateAuthenticator) DeriveEdges(goCtx context.Context, graph *interaction.Graph) ([]interaction.Edge, error) {
	return graph.Intent.DeriveEdgesForNode(goCtx, graph, n)
}

func (n *NodeDoUpdateAuthenticator) UserAuthenticator(stage authn.AuthenticationStage) (*authenticator.Info, bool) {
	if n.Stage == stage && n.AuthenticatorAfterUpdate != nil {
		return n.AuthenticatorAfterUpdate, true
	}
	return nil, false
}
