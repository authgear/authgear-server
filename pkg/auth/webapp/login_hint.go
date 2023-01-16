package webapp

import (
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"github.com/authgear/authgear-server/pkg/lib/authn/identity/anonymous"
	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/lib/interaction"
	"github.com/authgear/authgear-server/pkg/lib/interaction/intents"
	"github.com/authgear/authgear-server/pkg/lib/interaction/nodes"
	"github.com/authgear/authgear-server/pkg/util/clock"
	"github.com/authgear/authgear-server/pkg/util/httputil"
)

type anonymousTokenInput struct {
	JWT               string
	PromoteUserID     string
	PromoteIdentityID string
}

func (i *anonymousTokenInput) GetAnonymousRequestToken() string { return i.JWT }

func (i *anonymousTokenInput) SignUpAnonymousUserWithoutKey() bool { return false }

func (i *anonymousTokenInput) GetPromoteUserAndIdentityID() (string, string) {
	return i.PromoteUserID, i.PromoteIdentityID
}

var _ nodes.InputUseIdentityAnonymous = &anonymousTokenInput{}

type AnonymousIdentityProvider interface {
	ParseRequestUnverified(requestJWT string) (r *anonymous.Request, err error)
}

type AnonymousPromotionCodeStore interface {
	GetPromotionCode(codeHash string) (*anonymous.PromotionCode, error)
	DeletePromotionCode(code *anonymous.PromotionCode) error
}

type LoginHintPageService interface {
	PostWithIntent(session *Session, intent interaction.Intent, inputFn func() (interface{}, error)) (*Result, error)
}

type LoginHintHandler struct {
	Config                  *config.OAuthConfig
	Anonymous               AnonymousIdentityProvider
	AnonymousPromotionCodes AnonymousPromotionCodeStore
	Clock                   clock.Clock
	Pages                   LoginHintPageService
}

type HandleLoginHintOptions struct {
	SessionOptions      SessionOptions
	LoginHint           string
	UILocales           string
	ColorScheme         string
	OAuthSessionCookies []*http.Cookie
}

func (r *LoginHintHandler) HandleLoginHint(options HandleLoginHintOptions) (httputil.Result, error) {
	loginHint := options.LoginHint
	if !strings.HasPrefix(loginHint, "https://authgear.com/login_hint?") {
		return nil, fmt.Errorf("invalid login_hint: %v", loginHint)
	}

	u, err := url.Parse(loginHint)
	if err != nil {
		return nil, err
	}
	query := u.Query()

	switch query.Get("type") {
	case "anonymous":
		startPromotionInteraction := func(inputer func() (interface{}, error)) (httputil.Result, error) {
			intent := &intents.IntentAuthenticate{
				Kind:                     intents.IntentAuthenticateKindPromote,
				SuppressIDPSessionCookie: options.SessionOptions.SuppressIDPSessionCookie,
			}

			now := r.Clock.NowUTC()
			sessionOpts := options.SessionOptions
			sessionOpts.UpdatedAt = now
			session := NewSession(sessionOpts)
			result, err := r.Pages.PostWithIntent(session, intent, inputer)
			if err != nil {
				return nil, err
			}

			if result != nil {
				result.UILocales = options.UILocales
				result.ColorScheme = options.ColorScheme
				result.Cookies = append(result.Cookies, options.OAuthSessionCookies...)
			}
			return result, nil
		}

		promotionCode := query.Get("promotion_code")
		if promotionCode != "" {
			// promotion code flow
			userID, identityID, err := r.resolvePromotionCode(promotionCode)
			if err != nil {
				return nil, err
			}
			inputer := func() (interface{}, error) {
				return &anonymousTokenInput{
					PromoteUserID:     userID,
					PromoteIdentityID: identityID,
				}, nil
			}
			return startPromotionInteraction(inputer)
		}

		// jwt flow
		jwt := query.Get("jwt")
		request, err := r.Anonymous.ParseRequestUnverified(jwt)
		if err != nil {
			return nil, err
		}

		switch request.Action {
		case anonymous.RequestActionPromote:
			inputer := func() (interface{}, error) {
				return &anonymousTokenInput{JWT: jwt}, nil
			}
			return startPromotionInteraction(inputer)
		case anonymous.RequestActionAuth:
			// TODO(webapp): support anonymous auth
			panic("webapp: anonymous auth through web app is not supported")
		default:
			return nil, errors.New("unknown anonymous request action")
		}
	default:
		return nil, fmt.Errorf("unsupported login hint type: %s", query.Get("type"))
	}
}

func (r *LoginHintHandler) resolvePromotionCode(code string) (userID string, identityID string, err error) {
	codeObj, err := r.AnonymousPromotionCodes.GetPromotionCode(anonymous.HashPromotionCode(code))
	if err != nil {
		return
	}

	err = r.AnonymousPromotionCodes.DeletePromotionCode(codeObj)
	if err != nil {
		return
	}
	userID = codeObj.UserID
	identityID = codeObj.IdentityID
	return
}
