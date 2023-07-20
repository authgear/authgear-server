package service

import (
	"errors"
	"fmt"
	"net/http"
	"net/url"

	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/lib/config/configsource"
	portalconfig "github.com/authgear/authgear-server/pkg/portal/config"
)

type AuthzAdder interface {
	AddAuthz(
		auth config.AdminAPIAuth,
		appID config.AppID,
		authKey *config.AdminAPIAuthKey,
		auditContext interface{},
		hdr http.Header) (err error)
}

type AdminAPIService struct {
	AuthgearConfig *portalconfig.AuthgearConfig
	AppConfig      *portalconfig.AppConfig
	AdminAPIConfig *portalconfig.AdminAPIConfig
	ConfigSource   *configsource.ConfigSource
	AuthzAdder     AuthzAdder
}

type Usage string

const (
	UsageProxy    Usage = "proxy"
	UsageInternal Usage = "internal"
)

type PortalAdminAPIAuthContext struct {
	Usage       Usage  `json:"usage"`
	ActorUserID string `json:"actor_user_id,omitempty"`
	HTTPReferer string `json:"http_referer,omitempty"`
}

func (s *AdminAPIService) ResolveConfig(appID string) (*config.Config, error) {
	appCtx, err := s.ConfigSource.ContextResolver.ResolveContext(appID)
	if err != nil {
		return nil, err
	}
	return appCtx.Config, nil
}

func (s *AdminAPIService) ResolveHost(appID string) (host string, err error) {
	if s.AppConfig.HostSuffix == "" {
		return "", errors.New("app hostname suffix is not configured")
	}
	return appID + s.AppConfig.HostSuffix, nil
}

func (s *AdminAPIService) ResolveEndpoint(appID string) (*url.URL, error) {
	switch s.AdminAPIConfig.Type {
	case portalconfig.AdminAPITypeStatic:
		endpoint, err := url.Parse(s.AdminAPIConfig.Endpoint)
		if err != nil {
			return nil, err
		}
		return endpoint, nil
	default:
		panic(fmt.Errorf("portal: unexpected admin API type: %v", s.AdminAPIConfig.Type))
	}
}

func (s *AdminAPIService) Director(appID string, p string, actorUserID string, usage Usage) (director func(*http.Request), err error) {
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
	endpoint.Path = p

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

		err = s.AuthzAdder.AddAuthz(
			s.AdminAPIConfig.Auth,
			config.AppID(appID),
			authKey,
			PortalAdminAPIAuthContext{
				Usage:       usage,
				ActorUserID: actorUserID,
				HTTPReferer: r.Header.Get("Referer"),
			},
			r.Header,
		)
		if err != nil {
			panic(err)
		}
	}
	return
}

func (s *AdminAPIService) SelfDirector(actorUserID string, usage Usage) (director func(*http.Request), err error) {
	return s.Director(s.AuthgearConfig.AppID, "/graphql", actorUserID, usage)
}
