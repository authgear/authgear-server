package nodes

import (
	"github.com/authgear/authgear-server/pkg/lib/authn/authenticator"
	"github.com/authgear/authgear-server/pkg/lib/interaction"
)

func init() {
	interaction.RegisterNode(&NodeAuthenticationOOB{})
}

type InputAuthenticationOOB interface {
	GetOOBOTP() string
}

type EdgeAuthenticationOOB struct {
	Stage         interaction.AuthenticationStage
	Authenticator *authenticator.Info
	Secret        string
}

func (e *EdgeAuthenticationOOB) Instantiate(ctx *interaction.Context, graph *interaction.Graph, rawInput interface{}) (interaction.Node, error) {
	var input InputAuthenticationOOB
	if !interaction.Input(rawInput, &input) {
		return nil, interaction.ErrIncompatibleInput
	}

	info := e.Authenticator
	err := ctx.Authenticators.VerifySecret(info, nil, input.GetOOBOTP())
	if err != nil {
		info = nil
	}

	return &NodeAuthenticationOOB{Stage: e.Stage, Authenticator: info}, nil
}

type NodeAuthenticationOOB struct {
	Stage         interaction.AuthenticationStage `json:"stage"`
	Authenticator *authenticator.Info             `json:"authenticator"`
}

func (n *NodeAuthenticationOOB) Prepare(ctx *interaction.Context, graph *interaction.Graph) error {
	return nil
}

func (n *NodeAuthenticationOOB) GetEffects() ([]interaction.Effect, error) {
	return nil, nil
}

func (n *NodeAuthenticationOOB) DeriveEdges(graph *interaction.Graph) ([]interaction.Edge, error) {
	return []interaction.Edge{
		&EdgeAuthenticationEnd{
			Stage:                 n.Stage,
			VerifiedAuthenticator: n.Authenticator,
		},
	}, nil
}
