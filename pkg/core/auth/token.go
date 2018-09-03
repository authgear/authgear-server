package auth

import (
	"github.com/skygeario/skygear-server/pkg/core/config"
	"github.com/skygeario/skygear-server/pkg/server/authtoken"
)

type TokenStoreProvider struct{}

func (p TokenStoreProvider) Provide(tConfig config.TenantConfiguration) interface{} {
	// TODO:
	// mock token store
	return authtoken.NewJWTStore("my_skygear_jwt_secret", 3600)
}

type TokenResolver interface {
	Get(accessToken string, token *authtoken.Token) error
}
