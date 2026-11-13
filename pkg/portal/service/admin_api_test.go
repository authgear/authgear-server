package service

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	. "github.com/smartystreets/goconvey/convey"

	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/lib/config/configsource"
	portalconfig "github.com/authgear/authgear-server/pkg/portal/config"
)

type fakeAdminAPIConfigResolver struct{ origin string }

func (f *fakeAdminAPIConfigResolver) ResolveContext(ctx context.Context, appID string, fn func(context.Context, *config.AppContext) error) error {
	return fn(ctx, &config.AppContext{Config: &config.Config{
		AppConfig:    &config.AppConfig{HTTP: &config.HTTPConfig{PublicOrigin: f.origin}},
		SecretConfig: &config.SecretConfig{Secrets: []config.SecretItem{{Key: config.AdminAPIAuthKeyKey, Data: &config.AdminAPIAuthKey{}}}},
	}})
}

type fakeAdminAPIDefaultDomains struct{}

func (*fakeAdminAPIDefaultDomains) GetLatestAppHost(string) (string, error) {
	return "original.example.com", nil
}

type fakeAdminAPIAuthzAdder struct{}

func (*fakeAdminAPIAuthzAdder) AddAuthz(config.AdminAPIAuth, config.AppID, *config.AdminAPIAuthKey, any, http.Header) error {
	return nil
}

func TestAdminAPIDirectorPublicOrigin(t *testing.T) {
	Convey("Director uses the configured public origin while preserving internal routing", t, func() {
		for _, host := range []string{"original.example.com", "custom.example.com", "custom.example.com:8443"} {
			svc := &AdminAPIService{
				AdminAPIConfig: &portalconfig.AdminAPIConfig{Type: portalconfig.AdminAPITypeStatic, Endpoint: "http://admin.internal:3002"},
				ConfigSource:   &configsource.ConfigSource{ContextResolver: &fakeAdminAPIConfigResolver{origin: "https://" + host}},
				DefaultDomains: &fakeAdminAPIDefaultDomains{},
				AuthzAdder:     &fakeAdminAPIAuthzAdder{},
			}
			director, err := svc.Director(context.Background(), "app", "/graphql", "actor", UsageProxy)
			So(err, ShouldBeNil)
			req := httptest.NewRequest("POST", "http://portal.example.com/proxy?query=%7Busers%7D&x=1", nil)
			director(req)
			So(req.Host, ShouldEqual, host)
			So(req.Header.Get("X-Forwarded-Host"), ShouldEqual, host)
			So(req.URL.Scheme, ShouldEqual, "http")
			So(req.URL.Host, ShouldEqual, "admin.internal:3002")
			So(req.URL.Path, ShouldEqual, "/graphql")
			So(req.URL.RawQuery, ShouldEqual, "query=%7Busers%7D&x=1")
			So(req.Header.Get("X-Authgear-Portal-Is-Proxied"), ShouldEqual, "true")
		}
	})
}
