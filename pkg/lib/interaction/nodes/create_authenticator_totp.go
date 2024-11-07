package nodes

import (
	"context"

	"github.com/authgear/authgear-server/pkg/lib/authn"
	"github.com/authgear/authgear-server/pkg/lib/authn/authenticator"
	"github.com/authgear/authgear-server/pkg/lib/interaction"
)

func init() {
	interaction.RegisterNode(&NodeCreateAuthenticatorTOTP{})
}

type InputCreateAuthenticatorTOTP interface {
	GetTOTP() string
	GetTOTPDisplayName() string
}

type EdgeCreateAuthenticatorTOTP struct {
	Stage         authn.AuthenticationStage
	Authenticator *authenticator.Info
}

func (e *EdgeCreateAuthenticatorTOTP) Instantiate(goCtx context.Context, ctx *interaction.Context, graph *interaction.Graph, rawInput interface{}) (interaction.Node, error) {
	var input InputCreateAuthenticatorTOTP
	if !interaction.Input(rawInput, &input) {
		return nil, interaction.ErrIncompatibleInput
	}

	info := e.Authenticator
	info.TOTP.DisplayName = input.GetTOTPDisplayName()
	_, err := ctx.Authenticators.VerifyWithSpec(goCtx, info, &authenticator.Spec{
		TOTP: &authenticator.TOTPSpec{
			Code: input.GetTOTP(),
		},
	}, nil)
	if err != nil {
		return nil, err
	}

	return &NodeCreateAuthenticatorTOTP{Stage: e.Stage, Authenticator: info}, nil
}

type NodeCreateAuthenticatorTOTP struct {
	Stage         authn.AuthenticationStage `json:"stage"`
	Authenticator *authenticator.Info       `json:"authenticator"`
}

func (n *NodeCreateAuthenticatorTOTP) Prepare(goCtx context.Context, ctx *interaction.Context, graph *interaction.Graph) error {
	return nil
}

func (n *NodeCreateAuthenticatorTOTP) GetEffects(goCtx context.Context) ([]interaction.Effect, error) {
	return nil, nil
}

func (n *NodeCreateAuthenticatorTOTP) DeriveEdges(goCtx context.Context, graph *interaction.Graph) ([]interaction.Edge, error) {
	return []interaction.Edge{
		&EdgeCreateAuthenticatorEnd{
			Stage:          n.Stage,
			Authenticators: []*authenticator.Info{n.Authenticator},
		},
	}, nil
}
