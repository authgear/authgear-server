package nodes

import (
	"github.com/authgear/authgear-server/pkg/auth/dependency/authenticator"
	"github.com/authgear/authgear-server/pkg/auth/dependency/newinteraction"
)

type EdgeAuthenticationEnd struct {
	Stage         newinteraction.AuthenticationStage
	Authenticator *authenticator.Info
}

func (e *EdgeAuthenticationEnd) Instantiate(ctx *newinteraction.Context, graph *newinteraction.Graph, input interface{}) (newinteraction.Node, error) {
	return &NodeAuthenticationEnd{
		Authenticator: e.Authenticator,
	}, nil
}

type NodeAuthenticationEnd struct {
	Stage         newinteraction.AuthenticationStage `json:"stage"`
	Authenticator *authenticator.Info                `json:"authenticator"`
}

func (n *NodeAuthenticationEnd) Apply(ctx *newinteraction.Context, graph *newinteraction.Graph) error {
	panic("implement me")
}

func (n *NodeAuthenticationEnd) DeriveEdges(ctx *newinteraction.Context, graph *newinteraction.Graph) ([]newinteraction.Edge, error) {
	panic("implement me")
}

func (n *NodeAuthenticationEnd) UserAuthenticator() (newinteraction.AuthenticationStage, *authenticator.Info) {
	return n.Stage, n.Authenticator
}
