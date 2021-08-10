package service

import (
	"fmt"
	"net/http"
	"net/url"
	"path"
	"strings"
	texttemplate "text/template"

	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/lib/config/configsource"
	portalconfig "github.com/authgear/authgear-server/pkg/portal/config"
)

type AuthzAdder interface {
	AddAuthz(auth config.AdminAPIAuth, appID config.AppID, authKey *config.AdminAPIAuthKey, hdr http.Header) (err error)
}

type AdminAPIService struct {
	AuthgearConfig *portalconfig.AuthgearConfig
	AdminAPIConfig *portalconfig.AdminAPIConfig
	ConfigSource   *configsource.ConfigSource
	AuthzAdder     AuthzAdder
}

func (s *AdminAPIService) ResolveConfig(appID string) (*config.Config, error) {
	appCtx, err := s.ConfigSource.ContextResolver.ResolveContext(appID)
	if err != nil {
		return nil, err
	}
	return appCtx.Config, nil
}

func (s *AdminAPIService) ResolveHost(appID string) (host string, err error) {
	t := texttemplate.New("host-template")
	_, err = t.Parse(s.AdminAPIConfig.HostTemplate)
	if err != nil {
		return
	}
	var buf strings.Builder

	data := map[string]interface{}{
		"AppID": appID,
	}
	err = t.Execute(&buf, data)
	if err != nil {
		return
	}

	host = buf.String()
	return
}

func (s *AdminAPIService) ResolveEndpoint(appID string) (*url.URL, error) {
	switch s.AdminAPIConfig.Type {
	case portalconfig.AdminAPITypeStatic:
		endpoint, err := url.Parse(s.AdminAPIConfig.Endpoint)
		if err != nil {
			return nil, err
		}
		if endpoint.Path == "" {
			endpoint.Path = "/"
		}
		endpoint.Path = path.Join(endpoint.Path, "graphql")
		return endpoint, nil
	default:
		panic(fmt.Errorf("portal: unexpected admin API type: %v", s.AdminAPIConfig.Type))
	}
}

func (s *AdminAPIService) Director(appID string) (director func(*http.Request), err error) {
	cfg, err := s.ResolveConfig(appID)
	if err != nil {
		return
	}

	authKey, ok := cfg.SecretConfig.LookupData(config.AdminAPIAuthKeyKey).(*config.AdminAPIAuthKey)
	if !ok {
		err = fmt.Errorf("failed to look up admin API auth key: %v", appID)
		return
	}

	endpoint, err := s.ResolveEndpoint(appID)
	if err != nil {
		return
	}

	host, err := s.ResolveHost(appID)
	if err != nil {
		return
	}

	director = func(r *http.Request) {
		// It is important to preserve raw query so that GraphiQL ?query=... is not broken.
		rawQuery := r.URL.RawQuery
		r.URL = endpoint
		r.URL.RawQuery = rawQuery
		r.Host = host
		r.Header.Set("X-Forwarded-Host", r.Host)

		err = s.AuthzAdder.AddAuthz(s.AdminAPIConfig.Auth, config.AppID(appID), authKey, r.Header)
		if err != nil {
			panic(err)
		}
	}
	return
}

func (s *AdminAPIService) SelfDirector() (director func(*http.Request), err error) {
	return s.Director(s.AuthgearConfig.AppID)
}
