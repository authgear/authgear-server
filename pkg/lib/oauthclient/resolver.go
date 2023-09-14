package oauthclient

import (
	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/lib/tester"
)

type Resolver struct {
	OAuthConfig     *config.OAuthConfig
	TesterEndpoints tester.EndpointsProvider
}

func (r *Resolver) ResolveClientID(clientID string) *config.OAuthClientConfig {
	if clientID == tester.ClientIDTester {
		return tester.NewTesterClient(r.TesterEndpoints.TesterURL().String())
	}

	if client, ok := r.OAuthConfig.GetClient(clientID); ok {
		return client
	}
	return nil
}
