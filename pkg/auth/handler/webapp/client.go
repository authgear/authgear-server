package webapp

import "github.com/authgear/authgear-server/pkg/lib/config"

type WebappOAuthClientResolver interface {
	ResolveClient(clientID string) *config.OAuthClientConfig
}
