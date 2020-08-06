package webapp

import (
	"errors"
	"fmt"
	"net/url"
	"path"
	"strings"

	"github.com/authgear/authgear-server/pkg/auth/config"
	"github.com/authgear/authgear-server/pkg/auth/dependency/identity/anonymous"
	interactionintents "github.com/authgear/authgear-server/pkg/auth/dependency/newinteraction/intents"
	coreurl "github.com/authgear/authgear-server/pkg/core/url"
	"github.com/authgear/authgear-server/pkg/httputil"
)

type EndpointsProvider interface {
	AuthenticateEndpointURL() *url.URL
	PromoteUserEndpointURL() *url.URL
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
	return coreurl.WithQueryParamsAdded(
		p.Endpoints.LogoutEndpointURL(),
		map[string]string{"redirect_uri": redirectURI.String()},
	)
}

func (p *URLProvider) SettingsURL() *url.URL {
	return p.Endpoints.SettingsEndpointURL()
}

func (p *URLProvider) ResetPasswordURL(code string) *url.URL {
	return coreurl.WithQueryParamsAdded(
		p.Endpoints.ResetPasswordEndpointURL(),
		map[string]string{"code": code},
	)
}

func (p *URLProvider) VerifyIdentityURL(code string, webStateID string) *url.URL {
	return coreurl.WithQueryParamsAdded(
		p.Endpoints.VerifyIdentityEndpointURL(),
		map[string]string{"code": code, "state": webStateID},
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
	PostIntent(intent *Intent, inputer func() (interface{}, error)) (result *Result, err error)
}

type anonymousTokenInput struct{ JWT string }

func (i *anonymousTokenInput) GetAnonymousRequestToken() string { return i.JWT }

type AuthenticateURLProvider struct {
	Endpoints EndpointsProvider
	Anonymous AnonymousIdentityProvider
	Pages     PageService
}

func (p *AuthenticateURLProvider) AuthenticateURL(options AuthenticateURLOptions) (httputil.Result, error) {
	authnURI := p.Endpoints.AuthenticateEndpointURL()
	q := map[string]string{
		"redirect_uri": options.RedirectURI,
		"client_id":    options.ClientID,
	}
	if options.Prompt != "" {
		q["prompt"] = options.Prompt
	}
	if options.UILocales != "" {
		q["ui_locales"] = options.UILocales
	}
	if options.LoginHint != "" {
		resp, err := p.responseFromLoginHint(options)
		if err != nil {
			return nil, err
		} else if resp != nil {
			return resp, nil
		}
	}
	return &httputil.ResultRedirect{
		URL: coreurl.WithQueryParamsAdded(authnURI, q).String(),
	}, nil
}

func (p *AuthenticateURLProvider) responseFromLoginHint(options AuthenticateURLOptions) (httputil.Result, error) {
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
			return p.Pages.PostIntent(&Intent{
				RedirectURI: options.RedirectURI,
				KeepState:   false,
				Intent:      interactionintents.NewIntentPromote(),
			}, func() (interface{}, error) {
				return &anonymousTokenInput{JWT: jwt}, nil
			})

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
