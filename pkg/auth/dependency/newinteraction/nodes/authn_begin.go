package nodes

import (
	"github.com/authgear/authgear-server/pkg/auth/dependency/identity"
	"github.com/authgear/authgear-server/pkg/auth/dependency/newinteraction"
)

type EdgeAuthenticationBegin struct {
	Stage    newinteraction.AuthenticationStage
	Identity *identity.Info
}

func (e *EdgeAuthenticationBegin) Instantiate(ctx *newinteraction.Context, graph *newinteraction.Graph, input interface{}) (newinteraction.Node, error) {
	return &NodeAuthenticationBegin{
		Stage:    e.Stage,
		Identity: e.Identity,
	}, nil
}

type NodeAuthenticationBegin struct {
	Stage    newinteraction.AuthenticationStage `json:"stage"`
	Identity *identity.Info                     `json:"identity"`
}

func (n *NodeAuthenticationBegin) Apply(ctx *newinteraction.Context, graph *newinteraction.Graph) error {
	panic("implement me")
}

func (n *NodeAuthenticationBegin) DeriveEdges(ctx *newinteraction.Context, graph *newinteraction.Graph) ([]newinteraction.Edge, error) {
	panic("implement me")
}
