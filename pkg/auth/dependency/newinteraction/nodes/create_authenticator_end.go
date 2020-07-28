package nodes

import (
	"github.com/authgear/authgear-server/pkg/auth/dependency/authenticator"
	"github.com/authgear/authgear-server/pkg/auth/dependency/newinteraction"
)

func init() {
	newinteraction.RegisterNode(&NodeCreateAuthenticatorEnd{})
}

type EdgeCreateAuthenticatorEnd struct {
	Stage          newinteraction.AuthenticationStage
	Authenticators []*authenticator.Info
}

func (e *EdgeCreateAuthenticatorEnd) Instantiate(ctx *newinteraction.Context, graph *newinteraction.Graph, input interface{}) (newinteraction.Node, error) {
	return &NodeCreateAuthenticatorEnd{
		Stage:          e.Stage,
		Authenticators: e.Authenticators,
	}, nil
}

type NodeCreateAuthenticatorEnd struct {
	Stage          newinteraction.AuthenticationStage `json:"stage"`
	Authenticators []*authenticator.Info              `json:"authenticators"`
}

func (n *NodeCreateAuthenticatorEnd) Apply(perform func(eff newinteraction.Effect) error, graph *newinteraction.Graph) error {
	err := perform(newinteraction.EffectRun(func(ctx *newinteraction.Context) error {
		for _, a := range n.Authenticators {
			if err := ctx.Authenticators.Create(a); err != nil {
				return err
			}
		}

		return nil
	}))
	if err != nil {
		return err
	}

	// TODO(interaction): dispatch authenticator creation event if not creating user
	return nil
}

func (n *NodeCreateAuthenticatorEnd) DeriveEdges(ctx *newinteraction.Context, graph *newinteraction.Graph) ([]newinteraction.Edge, error) {
	return graph.Intent.DeriveEdgesForNode(ctx, graph, n)
}

func (n *NodeCreateAuthenticatorEnd) UserAuthenticator() (newinteraction.AuthenticationStage, *authenticator.Info) {
	if len(n.Authenticators) > 1 {
		panic("interaction: expect at most one primary/secondary authenticator")
	}
	if len(n.Authenticators) == 0 {
		return "", nil
	}
	return n.Stage, n.Authenticators[0]
}

func (n *NodeCreateAuthenticatorEnd) UserNewAuthenticators() []*authenticator.Info {
	return n.Authenticators
}
