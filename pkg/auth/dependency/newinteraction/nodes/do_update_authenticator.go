package nodes

import (
	"github.com/authgear/authgear-server/pkg/auth/dependency/newinteraction"
	"github.com/authgear/authgear-server/pkg/lib/authn/authenticator"
)

func init() {
	newinteraction.RegisterNode(&NodeDoUpdateAuthenticator{})
}

type EdgeDoUpdateAuthenticator struct {
	Stage                     newinteraction.AuthenticationStage
	AuthenticatorBeforeUpdate *authenticator.Info
	AuthenticatorAfterUpdate  *authenticator.Info
}

func (e *EdgeDoUpdateAuthenticator) Instantiate(ctx *newinteraction.Context, graph *newinteraction.Graph, rawInput interface{}) (newinteraction.Node, error) {
	return &NodeDoUpdateAuthenticator{
		AuthenticatorBeforeUpdate: e.AuthenticatorBeforeUpdate,
		AuthenticatorAfterUpdate:  e.AuthenticatorAfterUpdate,
	}, nil
}

type NodeDoUpdateAuthenticator struct {
	Stage                     newinteraction.AuthenticationStage `json:"stage"`
	AuthenticatorBeforeUpdate *authenticator.Info                `json:"authenticator_before_update"`
	AuthenticatorAfterUpdate  *authenticator.Info                `json:"authenticator_after_update"`
}

func (n *NodeDoUpdateAuthenticator) Prepare(ctx *newinteraction.Context, graph *newinteraction.Graph) error {
	return nil
}

func (n *NodeDoUpdateAuthenticator) Apply(perform func(eff newinteraction.Effect) error, graph *newinteraction.Graph) error {
	err := perform(newinteraction.EffectRun(func(ctx *newinteraction.Context) error {
		return ctx.Authenticators.Update(n.AuthenticatorAfterUpdate)
	}))
	if err != nil {
		return err
	}

	return nil
}

func (n *NodeDoUpdateAuthenticator) DeriveEdges(graph *newinteraction.Graph) ([]newinteraction.Edge, error) {
	return graph.Intent.DeriveEdgesForNode(graph, n)
}

func (n *NodeDoUpdateAuthenticator) UserAuthenticator(stage newinteraction.AuthenticationStage) (*authenticator.Info, bool) {
	if n.Stage == stage && n.AuthenticatorAfterUpdate != nil {
		return n.AuthenticatorAfterUpdate, true
	}
	return nil, false
}
