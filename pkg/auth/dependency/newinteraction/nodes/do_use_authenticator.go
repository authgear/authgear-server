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

func (n *NodeDoUseAuthenticator) Apply(perform func(eff newinteraction.Effect) error, graph *newinteraction.Graph) error {
	return nil
}

func (n *NodeDoUseAuthenticator) DeriveEdges(ctx *newinteraction.Context, graph *newinteraction.Graph) ([]newinteraction.Edge, error) {
	return graph.Intent.DeriveEdgesForNode(ctx, graph, n)
}

func (n *NodeDoUseAuthenticator) UserAuthenticator() (newinteraction.AuthenticationStage, *authenticator.Info) {
	return n.Stage, n.Authenticator
}
