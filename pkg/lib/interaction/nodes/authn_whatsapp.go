package nodes

import (
	"github.com/authgear/authgear-server/pkg/api/model"
	"github.com/authgear/authgear-server/pkg/lib/authn"
	"github.com/authgear/authgear-server/pkg/lib/authn/authenticator"
	"github.com/authgear/authgear-server/pkg/lib/authn/otp"
	"github.com/authgear/authgear-server/pkg/lib/facade"
	"github.com/authgear/authgear-server/pkg/lib/interaction"
)

func init() {
	interaction.RegisterNode(&NodeAuthenticationWhatsapp{})
}

type InputAuthenticationWhatsapp interface {
	GetWhatsappOTP() string
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
	code := input.GetWhatsappOTP()
	channel := model.AuthenticatorOOBChannelWhatsapp
	info := e.Authenticator
	_, err := ctx.Authenticators.VerifyWithSpec(e.Authenticator, &authenticator.Spec{
		OOBOTP: &authenticator.OOBOTPSpec{
			Code: code,
		},
	}, &facade.VerifyOptions{
		Form:       otp.FormCode,
		OOBChannel: &channel,
		AuthenticationDetails: facade.NewAuthenticationDetails(
			graph.MustGetUserID(),
			e.Stage,
			authn.AuthenticationTypeOOBOTPSMS,
		),
	})
	if err != nil {
		return nil, err
	}

	return &NodeAuthenticationWhatsapp{Stage: e.Stage, Authenticator: info}, nil
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
