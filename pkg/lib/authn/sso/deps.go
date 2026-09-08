package sso

import (
	"crypto/tls"
	"crypto/x509"
	"net/http"
	"net/url"
	"os"
	"time"

	"github.com/google/wire"

	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/util/httputil"
)

func ProvideOAuthHTTPClient(env *config.EnvironmentConfig, f *config.HTTPFeatureConfig) OAuthHTTPClient {
	// The end-to-end test transport replaces the transport wholesale, proxy
	// and all, so the address policy below cannot apply to it and does not
	// try to. This branch is reachable only when the deployment sets one of
	// the End2End* environment variables.
	if env.End2EndHTTPProxy != "" || env.End2EndTLSCACertFile != "" {
		client := httputil.NewExternalClient(5 * time.Second)

		transport := &http.Transport{
			TLSClientConfig: &tls.Config{
				// TLS 1.2 is minimum version by default
				MinVersion: tls.VersionTLS12,
			},
		}

		if env.End2EndTLSCACertFile != "" {
			caCertPool, err := x509.SystemCertPool()
			if err != nil {
				panic(err)
			}
			caCert, err := os.ReadFile(env.End2EndTLSCACertFile)
			if err != nil {
				panic(err)
			}
			caCertPool.AppendCertsFromPEM(caCert)
			transport.TLSClientConfig.RootCAs = caCertPool
		}

		if env.End2EndHTTPProxy != "" {
			proxyUrl, err := url.Parse(env.End2EndHTTPProxy)
			if err != nil {
				panic(err)
			}
			transport.Proxy = http.ProxyURL(proxyUrl)
		}

		client.Transport = transport
		return OAuthHTTPClient{client}
	}

	// An OAuth provider's discovery_document_endpoint comes from the
	// project's config, and the jwks_uri this client fetches next comes from
	// whatever that endpoint returned -- so the second hop is chosen by the
	// discovery server, not by anyone here. Both go through the same client,
	// so both are bound by the same address policy.
	return OAuthHTTPClient{
		httputil.NewSSRFSafeExternalClient(5*time.Second, httputil.SSRFSafeExternalClientOptions{
			AllowNonPublicAddresses: f.IsInsecureFetchAddressAllowed(),
			Sink:                    "sso.oauth.providers.discovery_document_endpoint",
		}),
	}
}

var DependencySet = wire.NewSet(
	ProvideOAuthHTTPClient,
	wire.Struct(new(OAuthProviderFactory), "*"),
	wire.Struct(new(SimpleStoreRedisFactory), "*"),
)
