package nodes

import (
	"errors"
	"net/http"

	"github.com/authgear/authgear-server/pkg/lib/authn"
	"github.com/authgear/authgear-server/pkg/lib/authn/mfa"
	"github.com/authgear/authgear-server/pkg/lib/interaction"
)

type InputUseDeviceToken interface {
	GetDeviceToken() string
}

type EdgeUseDeviceToken struct{}

func (e *EdgeUseDeviceToken) Instantiate(ctx *interaction.Context, graph *interaction.Graph, rawInput interface{}) (interaction.Node, error) {
	var input InputUseDeviceToken
	if !interaction.AsInput(rawInput, &input) {
		return nil, interaction.ErrIncompatibleInput
	}

	userID := graph.MustGetUserID()
	deviceToken := input.GetDeviceToken()

	err := ctx.MFA.VerifyDeviceToken(userID, deviceToken)
	if errors.Is(err, mfa.ErrDeviceTokenNotFound) {
		cookie := ctx.CookieFactory.ClearCookie(ctx.MFADeviceTokenCookie.Def)
		return nil, &interaction.ErrClearCookie{
			Cookies: []*http.Cookie{cookie},
			Inner:   interaction.ErrSameNode,
		}
	} else if err != nil {
		return nil, err
	}

	return &NodeAuthenticationEnd{
		Stage:              authn.AuthenticationStageSecondary,
		AuthenticationType: authn.AuthenticationTypeDeviceToken,
	}, nil
}
