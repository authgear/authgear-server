package webapp

import (
	"net/url"

	coreurl "github.com/skygeario/skygear-server/pkg/core/url"
)

type EndpointsProvider interface {
	AuthenticateEndpointURI() *url.URL
	LogoutEndpointURI() *url.URL
	SettingsEndpointURI() *url.URL
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
}

func (p *URLProvider) AuthenticateURI(options AuthenticateURLOptions) *url.URL {
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
	return coreurl.WithQueryParamsAdded(p.Endpoints.AuthenticateEndpointURI(), q)
}

func (p *URLProvider) LogoutURI() *url.URL {
	return p.Endpoints.LogoutEndpointURI()
}

func (p *URLProvider) SettingsURI() *url.URL {
	return p.Endpoints.SettingsEndpointURI()
}
