package newinteraction

import "github.com/authgear/authgear-server/pkg/auth/dependency/identity"

type AuthenticationStage string

const (
	AuthenticationStagePrimary   AuthenticationStage = "primary"
	AuthenticationStageSecondary AuthenticationStage = "secondary"
)

type EdgeAuthenticationBegin struct {
	Stage    AuthenticationStage
	Identity *identity.Info
}

func (e *EdgeAuthenticationBegin) Instantiate(ctx *Context, graph *Graph, input interface{}) (Node, error) {
	return &NodeAuthenticationBegin{
		Stage:    e.Stage,
		Identity: e.Identity,
	}, nil
}

type NodeAuthenticationBegin struct {
	Stage    AuthenticationStage `json:"stage"`
	Identity *identity.Info      `json:"identity"`
}

func (n *NodeAuthenticationBegin) Apply(ctx *Context, graph *Graph) error {
	panic("implement me")
}

func (n *NodeAuthenticationBegin) DeriveEdges(ctx *Context, graph *Graph) ([]Edge, error) {
	panic("implement me")
}
