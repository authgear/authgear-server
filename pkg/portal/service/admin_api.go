package service

import (
	"fmt"
	"net/http"
	"net/url"
	"path"

	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/lib/config/configsource"
	portalconfig "github.com/authgear/authgear-server/pkg/portal/config"
)

type AuthzAdder interface {
	AddAuthz(auth config.AdminAPIAuth, appID config.AppID, authKey *config.AdminAPIAuthKey, hdr http.Header) (err error)
}

type AdminAPIService struct {
	AdminAPIConfig *portalconfig.AdminAPIConfig
	ConfigSource   *configsource.ConfigSource
	AuthzAdder     AuthzAdder
}

func (s *AdminAPIService) ResolveConfig(appID string) (*config.Config, error) {
	return s.ConfigSource.ConfigGetter.GetConfig(appID)
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

func (s *AdminAPIService) AddAuthz(appID config.AppID, authKey *config.AdminAPIAuthKey, hdr http.Header) (err error) {
	return s.AuthzAdder.AddAuthz(s.AdminAPIConfig.Auth, appID, authKey, hdr)
}
