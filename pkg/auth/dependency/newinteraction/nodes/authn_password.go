package nodes

import (
	"github.com/authgear/authgear-server/pkg/auth/dependency/authenticator"
	"github.com/authgear/authgear-server/pkg/auth/dependency/newinteraction"
)

type InputAuthenticatePassword interface {
	GetPassword() string
}

type EdgeAuthenticatePassword struct {
	Stage newinteraction.AuthenticationStage
}

func (s *EdgeAuthenticatePassword) Instantiate(ctx *newinteraction.Context, graph *newinteraction.Graph, rawInput interface{}) (newinteraction.Node, error) {
	// TODO: authenticate password based on identity
	identity := graph.MustGetUserIdentity()
	_ = identity
	panic("implement me")
}

type NodeAuthenticatePassword struct {
	Stage         newinteraction.AuthenticationStage `json:"stage"`
	Authenticator *authenticator.Info                `json:"authenticator"`
}

func (n *NodeAuthenticatePassword) Apply(ctx *newinteraction.Context, graph *newinteraction.Graph) error {
	panic("implement me")
}

func (n *NodeAuthenticatePassword) DeriveEdges(ctx *newinteraction.Context, graph *newinteraction.Graph) ([]newinteraction.Edge, error) {
	panic("implement me")
}
