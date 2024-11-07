package nodes

import (
	"context"

	"github.com/authgear/authgear-server/pkg/lib/authn"
	"github.com/authgear/authgear-server/pkg/lib/authn/authenticator"
	"github.com/authgear/authgear-server/pkg/lib/authn/otp"
	"github.com/authgear/authgear-server/pkg/lib/interaction"
)

type InputOOBResendCode interface {
	DoResend()
}

type EdgeOOBResendCode struct {
	Stage            authn.AuthenticationStage
	IsAuthenticating bool
	Authenticator    *authenticator.Info
	OTPForm          otp.Form
}

func (e *EdgeOOBResendCode) Instantiate(goCtx context.Context, ctx *interaction.Context, graph *interaction.Graph, rawInput interface{}) (interaction.Node, error) {
	var input InputOOBResendCode
	if !interaction.Input(rawInput, &input) {
		return nil, interaction.ErrIncompatibleInput
	}

	_, err := (&SendOOBCode{
		Context:              ctx,
		Stage:                e.Stage,
		IsAuthenticating:     e.IsAuthenticating,
		AuthenticatorInfo:    e.Authenticator,
		IgnoreRatelimitError: false,
		OTPForm:              e.OTPForm,
	}).Do(goCtx)
	if err != nil {
		return nil, err
	}

	return nil, interaction.ErrSameNode
}
