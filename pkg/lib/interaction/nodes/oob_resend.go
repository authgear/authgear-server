package nodes

import (
	"github.com/authgear/authgear-server/pkg/lib/authn/authenticator"
	"github.com/authgear/authgear-server/pkg/lib/interaction"
)

type InputOOBResendCode interface {
	DoResend()
}

type EdgeOOBResendCode struct {
	Stage            interaction.AuthenticationStage
	IsAuthenticating bool
	Authenticator    *authenticator.Info
	Secret           string
}

func (e *EdgeOOBResendCode) Instantiate(ctx *interaction.Context, graph *interaction.Graph, rawInput interface{}) (interaction.Node, error) {
	var input InputOOBResendCode
	if !interaction.Input(rawInput, &input) {
		return nil, interaction.ErrIncompatibleInput
	}

	_, err := sendOOBCode(ctx, e.Stage, e.IsAuthenticating, e.Authenticator, e.Secret)
	if err != nil {
		return nil, err
	}

	return nil, interaction.ErrSameNode
}
