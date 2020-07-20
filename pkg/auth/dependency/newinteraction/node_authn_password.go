package newinteraction

import (
	"github.com/authgear/authgear-server/pkg/auth/dependency/authenticator"
)

type InputAuthenticatePassword interface {
	GetPassword() string
}

type EdgeAuthenticatePassword struct {
	Stage AuthenticationStage
}

func (s *EdgeAuthenticatePassword) Instantiate(ctx *Context, graph *Graph, rawInput interface{}) (Node, error) {
	// TODO: authenticate password based on identity
	identity := graph.MustGetUserIdentity()
	_ = identity
	panic("implement me")
}

type NodeAuthenticatePassword struct {
	Stage         AuthenticationStage
	Authenticator *authenticator.Info `json:"authenticator"`
}

func (n *NodeAuthenticatePassword) Apply(ctx *Context, graph *Graph) error {
	panic("implement me")
}

func (n *NodeAuthenticatePassword) DeriveEdges(ctx *Context, graph *Graph) ([]Edge, error) {
	panic("implement me")
}
