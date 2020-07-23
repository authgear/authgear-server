package nodes

import (
	"github.com/authgear/authgear-server/pkg/auth/dependency/newinteraction"
)

type EdgeAuthenticationBegin struct {
	Stage newinteraction.AuthenticationStage
}

func (e *EdgeAuthenticationBegin) Instantiate(ctx *newinteraction.Context, graph *newinteraction.Graph, input interface{}) (newinteraction.Node, error) {
	return &NodeAuthenticationBegin{
		Stage: e.Stage,
	}, nil
}

type NodeAuthenticationBegin struct {
	Stage newinteraction.AuthenticationStage `json:"stage"`
}

func (n *NodeAuthenticationBegin) Apply(ctx *newinteraction.Context, graph *newinteraction.Graph) error {
	panic("implement me")
}

func (n *NodeAuthenticationBegin) DeriveEdges(ctx *newinteraction.Context, graph *newinteraction.Graph) ([]newinteraction.Edge, error) {
	panic("implement me")
}
