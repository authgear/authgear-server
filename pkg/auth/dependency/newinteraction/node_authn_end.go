package newinteraction

import (
	"github.com/authgear/authgear-server/pkg/auth/dependency/authenticator"
)

type EdgeAuthenticationEnd struct {
	Stage         AuthenticationStage
	Authenticator *authenticator.Info
}

func (e *EdgeAuthenticationEnd) Instantiate(ctx *Context, graph *Graph, input interface{}) (Node, error) {
	return &NodeAuthenticationEnd{
		Authenticator: e.Authenticator,
	}, nil
}

type NodeAuthenticationEnd struct {
	Stage         AuthenticationStage `json:"stage"`
	Authenticator *authenticator.Info `json:"authenticator"`
}

func (n *NodeAuthenticationEnd) Apply(ctx *Context, graph *Graph) error {
	panic("implement me")
}

func (n *NodeAuthenticationEnd) DeriveEdges(ctx *Context, graph *Graph) ([]Edge, error) {
	panic("implement me")
}

func (n *NodeAuthenticationEnd) UserAuthenticator() (AuthenticationStage, *authenticator.Info) {
	return n.Stage, n.Authenticator
}
