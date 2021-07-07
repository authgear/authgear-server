package webapp

import (
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"path"

	"github.com/authgear/authgear-server/pkg/lib/authn/identity/anonymous"
	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/lib/interaction"
	interactionintents "github.com/authgear/authgear-server/pkg/lib/interaction/intents"
	"github.com/authgear/authgear-server/pkg/lib/session"
	"github.com/authgear/authgear-server/pkg/util/clock"
	"github.com/authgear/authgear-server/pkg/util/httputil"
	"github.com/authgear/authgear-server/pkg/util/urlutil"
)

type EndpointsProvider interface {
	BaseURL() *url.URL
	LoginEndpointURL() *url.URL
	SignupEndpointURL() *url.URL
	LogoutEndpointURL() *url.URL
	SettingsEndpointURL() *url.URL
	ResetPasswordEndpointURL() *url.URL
	VerifyIdentityEndpointURL() *url.URL
	SSOCallbackEndpointURL() *url.URL
	WeChatAuthorizeEndpointURL() *url.URL
	WeChatCallbackEndpointURL() *url.URL
}

type URLProvider struct {
	Endpoints EndpointsProvider
}

func (p *URLProvider) LogoutURL(redirectURI *url.URL) *url.URL {
	return urlutil.WithQueryParamsAdded(
		p.Endpoints.LogoutEndpointURL(),
		map[string]string{"redirect_uri": redirectURI.String()},
	)
}

func (p *URLProvider) SettingsURL() *url.URL {
	return p.Endpoints.SettingsEndpointURL()
}

func (p *URLProvider) ResetPasswordURL(code string) *url.URL {
	return urlutil.WithQueryParamsAdded(
		p.Endpoints.ResetPasswordEndpointURL(),
		map[string]string{"code": code},
	)
}

func (p *URLProvider) VerifyIdentityURL(code string, id string) *url.URL {
	return urlutil.WithQueryParamsAdded(
		p.Endpoints.VerifyIdentityEndpointURL(),
		map[string]string{"code": code, "id": id},
	)
}

func (p *URLProvider) SSOCallbackURL(c config.OAuthSSOProviderConfig) *url.URL {
	u := p.Endpoints.SSOCallbackEndpointURL()
	u.Path = path.Join(u.Path, url.PathEscape(c.Alias))
	return u
}

type AnonymousRequest struct {
	JWT     string
	Request *anonymous.Request
}

type RawSessionCookieRequest struct {
	Value string
}

type AuthenticateURLOptions struct {
	ClientID         string
	RedirectURI      string
	UILocales        string
	Prompt           []string
	Page             string
	WebhookState     string
	UserIDHint       string
	AuthenticateHint interface{}
}

type PageService interface {
	CreateSession(session *Session, redirectURI string) (*Result, error)
	PostWithIntent(session *Session, intent interaction.Intent, inputFn func() (interface{}, error)) (*Result, error)
}

type anonymousTokenInput struct{ JWT string }

func (i *anonymousTokenInput) GetAnonymousRequestToken() string { return i.JWT }

type AuthenticateURLProvider struct {
	Endpoints     EndpointsProvider
	Pages         PageService
	SessionCookie session.CookieDef
	CookieFactory CookieFactory
	Clock         clock.Clock
}

func (p *AuthenticateURLProvider) AuthenticateURL(options AuthenticateURLOptions) (httputil.Result, error) {
	now := p.Clock.NowUTC()
	session := NewSession(SessionOptions{
		RedirectURI:  options.RedirectURI,
		ClientID:     options.ClientID,
		Prompt:       options.Prompt,
		WebhookState: options.WebhookState,
		UserIDHint:   options.UserIDHint,
		UpdatedAt:    now,
	})

	var result *Result
	var err error
	if options.AuthenticateHint != nil {
		result, err = p.handleHint(options, session)
	} else {
		endpoint := p.Endpoints.BaseURL().String()
		switch options.Page {
		case "login":
			endpoint = p.Endpoints.LoginEndpointURL().String()
		case "signup":
			endpoint = p.Endpoints.SignupEndpointURL().String()
		}
		result, err = p.Pages.CreateSession(session, endpoint)
	}
	if result != nil {
		result.UILocales = options.UILocales
	}

	return result, err
}

func (p *AuthenticateURLProvider) handleHint(
	options AuthenticateURLOptions,
	session *Session,
) (*Result, error) {
	switch hint := options.AuthenticateHint.(type) {
	case AnonymousRequest:
		switch hint.Request.Action {
		case anonymous.RequestActionPromote:
			intent := interactionintents.NewIntentPromote()
			inputer := func() (interface{}, error) {
				return &anonymousTokenInput{JWT: hint.JWT}, nil
			}
			return p.Pages.PostWithIntent(session, intent, inputer)

		case anonymous.RequestActionAuth:
			// TODO(webapp): support anonymous auth
			panic("webapp: anonymous auth through web app is not supported")

		default:
			return nil, errors.New("unknown anonymous request action")
		}

	case RawSessionCookieRequest:
		cookie := p.CookieFactory.ValueCookie(p.SessionCookie.Def, hint.Value)
		return &Result{
			Cookies:     []*http.Cookie{cookie},
			RedirectURI: options.RedirectURI,
		}, nil

	default:
		return nil, fmt.Errorf("unsupported authenticate hint type: %T", options.AuthenticateHint)
	}
}
