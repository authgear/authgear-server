package nodes

import (
	"github.com/authgear/authgear-server/pkg/api/apierrors"
	"github.com/authgear/authgear-server/pkg/api/model"
	"github.com/authgear/authgear-server/pkg/lib/authn"
	"github.com/authgear/authgear-server/pkg/lib/authn/authenticator"
	"github.com/authgear/authgear-server/pkg/lib/authn/otp"
	"github.com/authgear/authgear-server/pkg/lib/feature/verification"
	"github.com/authgear/authgear-server/pkg/lib/interaction"
)

func init() {
	interaction.RegisterNode(&NodeCreateAuthenticatorLoginLinkOTP{})
}

type InputCreateAuthenticatorLoginLinkOTP interface {
	VerifyLoginLink()
}

type EdgeCreateAuthenticatorLoginLinkOTP struct {
	Stage         authn.AuthenticationStage
	Authenticator *authenticator.Info
}

func (e *EdgeCreateAuthenticatorLoginLinkOTP) Instantiate(ctx *interaction.Context, graph *interaction.Graph, rawInput interface{}) (interaction.Node, error) {
	var input InputCreateAuthenticatorLoginLinkOTP
	if !interaction.Input(rawInput, &input) {
		return nil, interaction.ErrIncompatibleInput
	}

	email := e.Authenticator.OOBOTP.Email
	err := ctx.OTPCodeService.VerifyOTP(
		otp.KindOOBOTPLink(ctx.Config, model.AuthenticatorOOBChannelEmail),
		email,
		"",
		&otp.VerifyOptions{
			UseSubmittedCode: true,
			UserID:           e.Authenticator.UserID,
		},
	)
	if apierrors.IsKind(err, otp.InvalidOTPCode) {
		return nil, verification.ErrInvalidVerificationCode
	} else if err != nil {
		return nil, err
	}

	return &NodeCreateAuthenticatorLoginLinkOTP{Stage: e.Stage, Authenticator: e.Authenticator}, nil
}

type NodeCreateAuthenticatorLoginLinkOTP struct {
	Stage         authn.AuthenticationStage `json:"stage"`
	Authenticator *authenticator.Info       `json:"authenticator"`
	Target        string                    `json:"target"`
	Channel       string                    `json:"channel"`
}

func (n *NodeCreateAuthenticatorLoginLinkOTP) Prepare(ctx *interaction.Context, graph *interaction.Graph) error {
	return nil
}

func (n *NodeCreateAuthenticatorLoginLinkOTP) GetEffects() ([]interaction.Effect, error) {
	return nil, nil
}

func (n *NodeCreateAuthenticatorLoginLinkOTP) DeriveEdges(graph *interaction.Graph) ([]interaction.Edge, error) {
	return []interaction.Edge{
		&EdgeCreateAuthenticatorEnd{
			Stage:          n.Stage,
			Authenticators: []*authenticator.Info{n.Authenticator},
		},
	}, nil
}
