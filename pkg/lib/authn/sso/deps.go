package sso

import (
	"crypto/tls"
	"crypto/x509"
	"net/http"
	"net/url"
	"os"

	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/google/wire"
)

func ProvideHTTPClient(env *config.EnvironmentConfig) *http.Client {
	client := &http.Client{}

	if env.End2EndHTTPProxy != "" || env.End2EndSSLCertFile != "" {
		transport := &http.Transport{
			TLSClientConfig: &tls.Config{
				// TLS 1.2 is minimum version by default
				MinVersion: tls.VersionTLS12,
			},
		}

		if env.End2EndSSLCertFile != "" {
			caCertPool, err := x509.SystemCertPool()
			if err != nil {
				panic(err)
			}
			caCert, err := os.ReadFile(env.End2EndSSLCertFile)
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

	return client
}

var DependencySet = wire.NewSet(
	ProvideHTTPClient,
	wire.Struct(new(OAuthProviderFactory), "*"),
)
