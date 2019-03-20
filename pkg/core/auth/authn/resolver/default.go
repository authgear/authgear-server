package resolver

import (
	"context"
	"net/http"

	"github.com/skygeario/skygear-server/pkg/core/auth"
	"github.com/skygeario/skygear-server/pkg/core/auth/authinfo"
	"github.com/skygeario/skygear-server/pkg/core/auth/authn"
	"github.com/skygeario/skygear-server/pkg/core/auth/authtoken"
	"github.com/skygeario/skygear-server/pkg/core/config"
	"github.com/skygeario/skygear-server/pkg/core/model"
)

type AuthContextResolverFactory struct{}

func (f AuthContextResolverFactory) NewResolver(ctx context.Context, tenantConfig config.TenantConfiguration) authn.AuthContextResolver {
	r := &DefaultAuthContextResolver{
		TokenStore:    auth.NewDefaultTokenStore(ctx, tenantConfig),
		AuthInfoStore: auth.NewDefaultAuthInfoStore(ctx, tenantConfig),
	}
	return r
}

type DefaultAuthContextResolver struct {
	TokenStore    authtoken.Store
	AuthInfoStore authinfo.Store
}

func (r DefaultAuthContextResolver) Resolve(req *http.Request, ctx auth.ContextSetter) (err error) {
	keyType := model.GetAccessKeyType(req)

	var (
		token    *authtoken.Token
		authInfo *authinfo.AuthInfo
	)

	if keyType == model.MasterAccessKey {
		token, authInfo, err = masterkeyAuthContextResolver{
			TokenStore:    r.TokenStore,
			AuthInfoStore: r.AuthInfoStore,
		}.Resolve(req)
	} else {
		token, authInfo, err = nonMasterkeyAuthContextResolver{
			TokenStore:    r.TokenStore,
			AuthInfoStore: r.AuthInfoStore,
		}.Resolve(req)
	}

	ctx.SetAccessKeyType(keyType)
	ctx.SetAuthInfo(authInfo)
	ctx.SetToken(token)

	return
}

// this ensures that our structure conform to certain interfaces.
var (
	_ authn.AuthContextResolver = &DefaultAuthContextResolver{}
)
