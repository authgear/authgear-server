package token

import (
	"context"

	"github.com/skygeario/skygear-server/pkg/core/config"
)

type TokenStoreProvider struct{}

func (p TokenStoreProvider) Provide(ctx context.Context, tConfig config.TenantConfiguration) interface{} {
	return NewJWTStore(tConfig.AppName, tConfig.TokenStore.Secret, tConfig.TokenStore.Expiry)
}

type TokenStore interface {
	NewToken(authInfoID string) (Token, error)
	Get(accessToken string, token *Token) error
	Put(token *Token) error
	Delete(accessToken string) error
}
