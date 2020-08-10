package nodes

import (
	"github.com/authgear/authgear-server/pkg/auth/dependency/authenticator"
	"github.com/authgear/authgear-server/pkg/auth/dependency/newinteraction"
)

func init() {
	newinteraction.RegisterNode(&NodeDoCreateAuthenticator{})
}

type EdgeDoCreateAuthenticator struct {
	Stage          newinteraction.AuthenticationStage
	Authenticators []*authenticator.Info
}

func (e *EdgeDoCreateAuthenticator) Instantiate(ctx *newinteraction.Context, graph *newinteraction.Graph, rawInput interface{}) (newinteraction.Node, error) {
	return &NodeDoCreateAuthenticator{
		Stage:          e.Stage,
		Authenticators: e.Authenticators,
	}, nil
}

type NodeDoCreateAuthenticator struct {
	Stage          newinteraction.AuthenticationStage `json:"stage"`
	Authenticators []*authenticator.Info              `json:"authenticators"`
}

func (n *NodeDoCreateAuthenticator) Prepare(ctx *newinteraction.Context, graph *newinteraction.Graph) error {
	return nil
}

func (n *NodeDoCreateAuthenticator) Apply(perform func(eff newinteraction.Effect) error, graph *newinteraction.Graph) error {
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

	return nil
}

func (n *NodeDoCreateAuthenticator) DeriveEdges(graph *newinteraction.Graph) ([]newinteraction.Edge, error) {
	return graph.Intent.DeriveEdgesForNode(graph, n)
}

func (n *NodeDoCreateAuthenticator) UserAuthenticator(stage newinteraction.AuthenticationStage) (*authenticator.Info, bool) {
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
