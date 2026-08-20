package webapp

import (
	"context"

	"github.com/authgear/authgear-server/pkg/lib/config"
)

type WebappOAuthClientResolver interface {
	ResolveClient(ctx context.Context, clientID string) *config.OAuthClientConfig
}
