package nodes

import (
	"github.com/authgear/authgear-server/pkg/lib/authn"
	"github.com/authgear/authgear-server/pkg/lib/authn/authenticator"
	"github.com/authgear/authgear-server/pkg/lib/interaction"
)

func init() {
	interaction.RegisterNode(&NodeAuthenticationWhatsapp{})
}

type InputAuthenticationWhatsapp interface {
	VerifyWhatsappOTP()
}

type EdgeAuthenticationWhatsapp struct {
	Stage         authn.AuthenticationStage
	Authenticator *authenticator.Info
}

func (e *EdgeAuthenticationWhatsapp) Instantiate(ctx *interaction.Context, graph *interaction.Graph, rawInput interface{}) (interaction.Node, error) {
	var input InputAuthenticationWhatsapp
	if !interaction.Input(rawInput, &input) {
		return nil, interaction.ErrIncompatibleInput
	}

	phone := e.Authenticator.OOBOTP.Phone
	_, err := ctx.WhatsappCodeProvider.VerifyCode(phone, ctx.WebSessionID, true)
	if err != nil {
		return nil, err
	}

	return &NodeAuthenticationWhatsapp{Stage: e.Stage, Authenticator: e.Authenticator}, nil
}

type NodeAuthenticationWhatsapp struct {
	Stage         authn.AuthenticationStage `json:"stage"`
	Authenticator *authenticator.Info       `json:"authenticator"`
}

func (n *NodeAuthenticationWhatsapp) Prepare(ctx *interaction.Context, graph *interaction.Graph) error {
	return nil
}

func (n *NodeAuthenticationWhatsapp) GetEffects() ([]interaction.Effect, error) {
	return nil, nil
}

func (n *NodeAuthenticationWhatsapp) DeriveEdges(graph *interaction.Graph) ([]interaction.Edge, error) {
	return []interaction.Edge{
		&EdgeAuthenticationEnd{
			Stage:                 n.Stage,
			AuthenticationType:    authn.AuthenticationTypeOOBOTPSMS,
			VerifiedAuthenticator: n.Authenticator,
		},
	}, nil
}
