package nodes

import (
	"github.com/authgear/authgear-server/pkg/auth/dependency/authenticator"
	"github.com/authgear/authgear-server/pkg/auth/dependency/identity"
	"github.com/authgear/authgear-server/pkg/auth/dependency/newinteraction"
)

type InputOOBResendCode interface {
	DoResend()
}

type EdgeOOBResendCode struct {
	Stage            newinteraction.AuthenticationStage
	IsAuthenticating bool
	Identity         *identity.Info
	Authenticator    *authenticator.Info
	Secret           string
}

func (e *EdgeOOBResendCode) Instantiate(ctx *newinteraction.Context, graph *newinteraction.Graph, rawInput interface{}) (newinteraction.Node, error) {
	_, ok := rawInput.(InputOOBResendCode)
	if !ok {
		return nil, newinteraction.ErrIncompatibleInput
	}

	_, err := sendOOBCode(ctx, e.Stage, e.IsAuthenticating, e.Identity, e.Authenticator, e.Secret)
	if err != nil {
		return nil, err
	}

	return nil, newinteraction.ErrSameNode
}
