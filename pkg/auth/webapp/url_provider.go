package webapp

import (
	"errors"
	"fmt"
	"net/url"
	"path"
	"strings"

	"github.com/authgear/authgear-server/pkg/lib/authn/identity/anonymous"
	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/lib/interaction"
	interactionintents "github.com/authgear/authgear-server/pkg/lib/interaction/intents"
	"github.com/authgear/authgear-server/pkg/util/httputil"
	"github.com/authgear/authgear-server/pkg/util/urlutil"
)

type EndpointsProvider interface {
	AuthenticateEndpointURL() *url.URL
	LogoutEndpointURL() *url.URL
	SettingsEndpointURL() *url.URL
	ResetPasswordEndpointURL() *url.URL
	VerifyIdentityEndpointURL() *url.URL
	SSOCallbackEndpointURL() *url.URL
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

type AnonymousIdentityProvider interface {
	ParseRequestUnverified(requestJWT string) (r *anonymous.Request, err error)
}

type AuthenticateURLOptions struct {
	ClientID    string
	RedirectURI string
	UILocales   string
	Prompt      string
	LoginHint   string
}

type PageService interface {
	CreateSession(session *Session, redirectURI string) (*Result, error)
	PostWithIntent(session *Session, intent interaction.Intent, inputFn func() (interface{}, error)) (*Result, error)
}

type anonymousTokenInput struct{ JWT string }

func (i *anonymousTokenInput) GetAnonymousRequestToken() string { return i.JWT }

type AuthenticateURLProvider struct {
	Endpoints EndpointsProvider
	Anonymous AnonymousIdentityProvider
	Pages     PageService
}

func (p *AuthenticateURLProvider) AuthenticateURL(options AuthenticateURLOptions) (httputil.Result, error) {
	session := NewSession(SessionOptions{
		RedirectURI: options.RedirectURI,
		Prompt:      options.Prompt,
		UILocales:   options.UILocales,
	})
	if options.LoginHint != "" {
		result, err := p.handleLoginHint(options, session)
		if err != nil {
			return nil, err
		}
		if result != nil {
			return result, nil
		}
	}

	return p.Pages.CreateSession(session, p.Endpoints.AuthenticateEndpointURL().String())
}

func (p *AuthenticateURLProvider) handleLoginHint(
	options AuthenticateURLOptions,
	session *Session,
) (httputil.Result, error) {
	if !strings.HasPrefix(options.LoginHint, "https://authgear.com/login_hint?") {
		return nil, nil
	}

	url, err := url.Parse(options.LoginHint)
	if err != nil {
		return nil, err
	}
	query := url.Query()

	switch query.Get("type") {
	case "anonymous":
		jwt := query.Get("jwt")
		request, err := p.Anonymous.ParseRequestUnverified(query.Get("jwt"))
		if err != nil {
			return nil, err
		}

		switch request.Action {
		case anonymous.RequestActionPromote:
			intent := interactionintents.NewIntentPromote()
			inputer := func() (interface{}, error) {
				return &anonymousTokenInput{JWT: jwt}, nil
			}
			return p.Pages.PostWithIntent(session, intent, inputer)

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
