package webapp

import (
	"errors"
	"fmt"
	"net/url"
	"path"
	"strings"

	"github.com/authgear/authgear-server/pkg/auth/config"
	"github.com/authgear/authgear-server/pkg/auth/dependency/identity/anonymous"
	interactionflows "github.com/authgear/authgear-server/pkg/auth/dependency/interaction/flows"
	coreurl "github.com/authgear/authgear-server/pkg/core/url"
)

type EndpointsProvider interface {
	AuthenticateEndpointURL() *url.URL
	PromoteUserEndpointURL() *url.URL
	LogoutEndpointURL() *url.URL
	SettingsEndpointURL() *url.URL
	ResetPasswordEndpointURL() *url.URL
	SSOCallbackEndpointURL() *url.URL
}

type AnonymousFlow interface {
	DecodeUserID(requestJWT string) (string, anonymous.RequestAction, error)
}

type AuthenticateURLOptions struct {
	ClientID    string
	RedirectURI string
	UILocales   string
	Prompt      string
	LoginHint   string
}

type URLProviderStates interface {
	Set(*interactionflows.State) error
}

type URLProvider struct {
	Endpoints EndpointsProvider
	Anonymous AnonymousFlow
	States    URLProviderStates
}

func (p *URLProvider) AuthenticateURL(options AuthenticateURLOptions) (*url.URL, error) {
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
		err := p.convertLoginHint(&authnURI, q, options)
		if err != nil {
			return nil, err
		}
	}
	return coreurl.WithQueryParamsAdded(authnURI, q), nil
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

func (p *URLProvider) convertLoginHint(uri **url.URL, q map[string]string, options AuthenticateURLOptions) error {
	if !strings.HasPrefix(options.LoginHint, "https://authgear.com/login_hint?") {
		return nil
	}

	url, err := url.Parse(options.LoginHint)
	if err != nil {
		return err
	}
	query := url.Query()

	switch query.Get("type") {
	case "anonymous":
		userID, action, err := p.Anonymous.DecodeUserID(query.Get("jwt"))
		if err != nil {
			return err
		}

		switch action {
		case anonymous.RequestActionPromote:
			// FIXME(webapp): Create promote interaction eagerly.
			state := interactionflows.NewState()
			state.Extra[interactionflows.ExtraAnonymousUserID] = userID
			state.Extra[interactionflows.ExtraRedirectURI] = options.RedirectURI
			err = p.States.Set(state)
			if err != nil {
				return err
			}
			q["x_sid"] = state.ID
			*uri = p.Endpoints.PromoteUserEndpointURL()
			return nil
		case anonymous.RequestActionAuth:
			// TODO(webapp): support anonymous auth
			panic("webapp: anonymous auth through web app is not supported")
		default:
			return errors.New("unknown anonymous request action")
		}

	default:
		return fmt.Errorf("unsupported login hint type: %s", query.Get("type"))
	}
}
func (p *URLProvider) SSOCallbackURL(c config.OAuthSSOProviderConfig) *url.URL {
	u := p.Endpoints.SSOCallbackEndpointURL()
	u.Path = path.Join(u.Path, url.PathEscape(c.Alias))
	return u
}
