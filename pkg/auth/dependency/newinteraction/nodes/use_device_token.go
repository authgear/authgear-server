package nodes

import (
	"errors"
	"net/http"

	"github.com/authgear/authgear-server/pkg/auth/dependency/newinteraction"
	"github.com/authgear/authgear-server/pkg/lib/authn/mfa"
)

type InputUseDeviceToken interface {
	GetDeviceToken() string
}

type EdgeUseDeviceToken struct{}

func (e *EdgeUseDeviceToken) Instantiate(ctx *newinteraction.Context, graph *newinteraction.Graph, rawInput interface{}) (newinteraction.Node, error) {
	input, ok := rawInput.(InputUseDeviceToken)
	if !ok {
		return nil, newinteraction.ErrIncompatibleInput
	}

	userID := graph.MustGetUserID()
	deviceToken := input.GetDeviceToken()

	err := ctx.MFA.VerifyDeviceToken(userID, deviceToken)
	if errors.Is(err, mfa.ErrDeviceTokenNotFound) {
		cookie := ctx.CookieFactory.ClearCookie(ctx.MFADeviceTokenCookie.Def)
		return nil, &newinteraction.ErrClearCookie{
			Cookies: []*http.Cookie{cookie},
			Inner:   newinteraction.ErrSameNode,
		}
	} else if err != nil {
		return nil, err
	}

	return &NodeAuthenticationEnd{
		Stage:  newinteraction.AuthenticationStageSecondary,
		Result: AuthenticationResultDeviceToken,
	}, nil
}
