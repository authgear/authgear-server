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

func ProvideOAuthHTTPClient(env *config.EnvironmentConfig) OAuthHTTPClient {
	client := httputil.NewExternalClient(5 * time.Second)

	if env.End2EndHTTPProxy != "" || env.End2EndTLSCACertFile != "" {
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
	}

	return OAuthHTTPClient{client}
}

var DependencySet = wire.NewSet(
	ProvideOAuthHTTPClient,
	wire.Struct(new(OAuthProviderFactory), "*"),
	wire.Struct(new(SimpleStoreRedisFactory), "*"),
)
