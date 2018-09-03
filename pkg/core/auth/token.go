package auth

import (
	"github.com/skygeario/skygear-server/pkg/core/config"
	"github.com/skygeario/skygear-server/pkg/server/authtoken"
)

type TokenStoreProvider struct{}

func (p TokenStoreProvider) Provide(tConfig config.TenantConfiguration) interface{} {
	return authtoken.NewJWTStore(tConfig.TokenStore.Secret, tConfig.TokenStore.Expiry)
}

type TokenResolver interface {
	Get(accessToken string, token *authtoken.Token) error
}
