package nodes

import (
	"context"

	"github.com/authgear/authgear-server/pkg/api/model"
	"github.com/authgear/authgear-server/pkg/lib/authn"
	"github.com/authgear/authgear-server/pkg/lib/authn/authenticator"
	"github.com/authgear/authgear-server/pkg/lib/authn/otp"
	"github.com/authgear/authgear-server/pkg/lib/facade"
	"github.com/authgear/authgear-server/pkg/lib/interaction"
)

func init() {
	interaction.RegisterNode(&NodeCreateAuthenticatorWhatsappOTP{})
}

type InputCreateAuthenticatorWhatsappOTP interface {
	GetWhatsappOTP() string
}

type EdgeCreateAuthenticatorWhatsappOTP struct {
	Stage         authn.AuthenticationStage
	Authenticator *authenticator.Info
}

func (e *EdgeCreateAuthenticatorWhatsappOTP) Instantiate(goCtx context.Context, ctx *interaction.Context, graph *interaction.Graph, rawInput interface{}) (interaction.Node, error) {
	var input InputCreateAuthenticatorWhatsappOTP
	if !interaction.Input(rawInput, &input) {
		return nil, interaction.ErrIncompatibleInput
	}
	code := input.GetWhatsappOTP()
	channel := model.AuthenticatorOOBChannelWhatsapp
	_, err := ctx.Authenticators.VerifyWithSpec(goCtx, e.Authenticator, &authenticator.Spec{
		OOBOTP: &authenticator.OOBOTPSpec{
			Code: code,
		},
	}, &facade.VerifyOptions{
		Form:       otp.FormCode,
		OOBChannel: &channel,
	})
	if err != nil {
		return nil, err
	}

	return &NodeCreateAuthenticatorWhatsappOTP{Stage: e.Stage, Authenticator: e.Authenticator}, nil
}

type NodeCreateAuthenticatorWhatsappOTP struct {
	Stage         authn.AuthenticationStage `json:"stage"`
	Authenticator *authenticator.Info       `json:"authenticator"`
}

func (n *NodeCreateAuthenticatorWhatsappOTP) Prepare(goCtx context.Context, ctx *interaction.Context, graph *interaction.Graph) error {
	return nil
}

func (n *NodeCreateAuthenticatorWhatsappOTP) GetEffects(goCtx context.Context) ([]interaction.Effect, error) {
	return nil, nil
}

func (n *NodeCreateAuthenticatorWhatsappOTP) DeriveEdges(goCtx context.Context, graph *interaction.Graph) ([]interaction.Edge, error) {
	return []interaction.Edge{
		&EdgeCreateAuthenticatorEnd{
			Stage:          n.Stage,
			Authenticators: []*authenticator.Info{n.Authenticator},
		},
	}, nil
}
