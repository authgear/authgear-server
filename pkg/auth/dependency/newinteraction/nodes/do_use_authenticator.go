package nodes

import (
	"github.com/authgear/authgear-server/pkg/auth/dependency/authenticator"
	"github.com/authgear/authgear-server/pkg/auth/dependency/newinteraction"
)

func init() {
	newinteraction.RegisterNode(&NodeDoUseAuthenticator{})
}

type EdgeDoUseAuthenticator struct {
	Stage         newinteraction.AuthenticationStage
	Authenticator *authenticator.Info
}

func (e *EdgeDoUseAuthenticator) Instantiate(ctx *newinteraction.Context, graph *newinteraction.Graph, rawInput interface{}) (newinteraction.Node, error) {
	return &NodeDoUseAuthenticator{
		Stage:         e.Stage,
		Authenticator: e.Authenticator,
	}, nil
}

type NodeDoUseAuthenticator struct {
	Stage         newinteraction.AuthenticationStage `json:"stage"`
	Authenticator *authenticator.Info                `json:"authenticator"`
}

func (n *NodeDoUseAuthenticator) Prepare(ctx *newinteraction.Context, graph *newinteraction.Graph) error {
	return nil
}

func (n *NodeDoUseAuthenticator) Apply(perform func(eff newinteraction.Effect) error, graph *newinteraction.Graph) error {
	return nil
}

func (n *NodeDoUseAuthenticator) DeriveEdges(graph *newinteraction.Graph) ([]newinteraction.Edge, error) {
	return graph.Intent.DeriveEdgesForNode(graph, n)
}

func (n *NodeDoUseAuthenticator) UserAuthenticator(stage newinteraction.AuthenticationStage) (*authenticator.Info, bool) {
	if n.Stage == stage && n.Authenticator != nil {
		return n.Authenticator, true
	}
	return nil, false
}
