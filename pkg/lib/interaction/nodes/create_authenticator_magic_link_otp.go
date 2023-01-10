package nodes

import (
	"github.com/authgear/authgear-server/pkg/lib/authn"
	"github.com/authgear/authgear-server/pkg/lib/authn/authenticator"
	"github.com/authgear/authgear-server/pkg/lib/interaction"
)

func init() {
	interaction.RegisterNode(&NodeCreateAuthenticatorMagicLinkOTP{})
}

type InputCreateAuthenticatorMagicLinkOTP interface {
	VerifyMagicLinkOTP()
}

type EdgeCreateAuthenticatorMagicLinkOTP struct {
	Stage         authn.AuthenticationStage
	Authenticator *authenticator.Info
}

func (e *EdgeCreateAuthenticatorMagicLinkOTP) Instantiate(ctx *interaction.Context, graph *interaction.Graph, rawInput interface{}) (interaction.Node, error) {
	var input InputCreateAuthenticatorMagicLinkOTP
	if !interaction.Input(rawInput, &input) {
		return nil, interaction.ErrIncompatibleInput
	}

	// TODO(newman): Verify magic link

	return &NodeCreateAuthenticatorMagicLinkOTP{Stage: e.Stage, Authenticator: e.Authenticator}, nil
}

type NodeCreateAuthenticatorMagicLinkOTP struct {
	Stage         authn.AuthenticationStage `json:"stage"`
	Authenticator *authenticator.Info       `json:"authenticator"`
}

func (n *NodeCreateAuthenticatorMagicLinkOTP) Prepare(ctx *interaction.Context, graph *interaction.Graph) error {
	return nil
}

func (n *NodeCreateAuthenticatorMagicLinkOTP) GetEffects() ([]interaction.Effect, error) {
	return nil, nil
}

func (n *NodeCreateAuthenticatorMagicLinkOTP) DeriveEdges(graph *interaction.Graph) ([]interaction.Edge, error) {
	return []interaction.Edge{
		&EdgeCreateAuthenticatorEnd{
			Stage:          n.Stage,
			Authenticators: []*authenticator.Info{n.Authenticator},
		},
	}, nil
}
