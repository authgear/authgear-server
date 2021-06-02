package nodes

import (
	"github.com/authgear/authgear-server/pkg/lib/authn"
	"github.com/authgear/authgear-server/pkg/lib/authn/authenticator"
	"github.com/authgear/authgear-server/pkg/lib/interaction"
)

type InputOOBResendCode interface {
	DoResend()
}

type EdgeOOBResendCode struct {
	Stage            authn.AuthenticationStage
	IsAuthenticating bool
	Authenticator    *authenticator.Info
}

func (e *EdgeOOBResendCode) Instantiate(ctx *interaction.Context, graph *interaction.Graph, rawInput interface{}) (interaction.Node, error) {
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
	}).Do()
	if err != nil {
		return nil, err
	}

	return nil, interaction.ErrSameNode
}
