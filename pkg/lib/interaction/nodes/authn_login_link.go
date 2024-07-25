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
	interaction.RegisterNode(&NodeAuthenticationLoginLink{})
}

type InputAuthenticationLoginLink interface {
	VerifyLoginLink()
}

type EdgeAuthenticationLoginLink struct {
	Stage         authn.AuthenticationStage
	Authenticator *authenticator.Info
}

func (e *EdgeAuthenticationLoginLink) Instantiate(ctx *interaction.Context, graph *interaction.Graph, rawInput interface{}) (interaction.Node, error) {
	var input InputAuthenticationLoginLink
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

	return &NodeAuthenticationLoginLink{Stage: e.Stage, Authenticator: e.Authenticator}, nil
}

type NodeAuthenticationLoginLink struct {
	Stage         authn.AuthenticationStage `json:"stage"`
	Authenticator *authenticator.Info       `json:"authenticator"`
}

func (n *NodeAuthenticationLoginLink) Prepare(ctx *interaction.Context, graph *interaction.Graph) error {
	return nil
}

func (n *NodeAuthenticationLoginLink) GetEffects() ([]interaction.Effect, error) {
	return nil, nil
}

func (n *NodeAuthenticationLoginLink) DeriveEdges(graph *interaction.Graph) ([]interaction.Edge, error) {
	return []interaction.Edge{
		&EdgeAuthenticationEnd{
			Stage:                 n.Stage,
			AuthenticationType:    authn.AuthenticationTypeOOBOTPEmail,
			VerifiedAuthenticator: n.Authenticator,
		},
	}, nil
}
