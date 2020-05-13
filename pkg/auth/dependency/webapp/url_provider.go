package webapp

import (
	"errors"
	"fmt"
	"net/url"
	"strings"

	"github.com/skygeario/skygear-server/pkg/auth/dependency/identity/anonymous"
	coreurl "github.com/skygeario/skygear-server/pkg/core/url"
)

type EndpointsProvider interface {
	AuthenticateEndpointURI() *url.URL
	PromoteUserEndpointURI() *url.URL
	LogoutEndpointURI() *url.URL
	SettingsEndpointURI() *url.URL
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

type URLProvider struct {
	Endpoints EndpointsProvider
	Anonymous AnonymousFlow
	States    StateStore
}

func (p *URLProvider) AuthenticateURI(options AuthenticateURLOptions) (*url.URL, error) {
	authnURI := p.Endpoints.AuthenticateEndpointURI()
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
		err := p.convertLoginHint(&authnURI, q, options.LoginHint)
		if err != nil {
			return nil, err
		}
	}
	return coreurl.WithQueryParamsAdded(authnURI, q), nil
}

func (p *URLProvider) LogoutURI() *url.URL {
	return p.Endpoints.LogoutEndpointURI()
}

func (p *URLProvider) SettingsURI() *url.URL {
	return p.Endpoints.SettingsEndpointURI()
}

func (p *URLProvider) convertLoginHint(uri **url.URL, q map[string]string, loginHint string) error {
	if !strings.HasPrefix(loginHint, "https://auth.skygear.io/login_hint?") {
		return nil
	}

	url, err := url.Parse(loginHint)
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
			state := NewState()
			state.AnonymousUserID = userID
			p.States.Set(state)
			q["x_sid"] = state.ID
			*uri = p.Endpoints.PromoteUserEndpointURI()
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
