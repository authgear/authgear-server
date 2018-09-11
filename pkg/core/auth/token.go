package auth

import (
	"context"
	"net/http"

	"github.com/skygeario/skygear-server/pkg/core/config"
	"github.com/skygeario/skygear-server/pkg/server/authtoken"
)

func GetAccessToken(r *http.Request) string {
	return r.Header.Get("X-Skygear-Access-Token")
}

type TokenStoreProvider struct{}

func (p TokenStoreProvider) Provide(ctx context.Context, tConfig config.TenantConfiguration) interface{} {
	return authtoken.NewJWTStore(tConfig.TokenStore.Secret, tConfig.TokenStore.Expiry)
}

type TokenStore interface {
	NewToken(appName string, authInfoID string) (authtoken.Token, error)
	Get(accessToken string, token *authtoken.Token) error
	Put(token *authtoken.Token) error
	Delete(accessToken string) error
}
