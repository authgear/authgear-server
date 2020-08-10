package nodes

import (
	"github.com/authgear/authgear-server/pkg/auth/dependency/authenticator"
	"github.com/authgear/authgear-server/pkg/auth/dependency/newinteraction"
)

func init() {
	newinteraction.RegisterNode(&NodeAuthenticationEnd{})
}

type EdgeAuthenticationEnd struct {
	Stage                 newinteraction.AuthenticationStage
	Optional              bool
	VerifiedAuthenticator *authenticator.Info
}

func (e *EdgeAuthenticationEnd) Instantiate(ctx *newinteraction.Context, graph *newinteraction.Graph, input interface{}) (newinteraction.Node, error) {
	return &NodeAuthenticationEnd{
		Stage:                 e.Stage,
		Optional:              e.Optional,
		VerifiedAuthenticator: e.VerifiedAuthenticator,
	}, nil
}

type NodeAuthenticationEnd struct {
	Stage                 newinteraction.AuthenticationStage `json:"stage"`
	Optional              bool                               `json:"optional"`
	VerifiedAuthenticator *authenticator.Info                `json:"verified_authenticator"`
}

func (n *NodeAuthenticationEnd) Prepare(ctx *newinteraction.Context, graph *newinteraction.Graph) error {
	return nil
}

func (n *NodeAuthenticationEnd) Apply(perform func(eff newinteraction.Effect) error, graph *newinteraction.Graph) error {
	return nil
}

func (n *NodeAuthenticationEnd) DeriveEdges(graph *newinteraction.Graph) ([]newinteraction.Edge, error) {
	if !n.Optional && n.VerifiedAuthenticator == nil {
		return nil, newinteraction.ErrInvalidCredentials
	}

	return graph.Intent.DeriveEdgesForNode(graph, n)
}
