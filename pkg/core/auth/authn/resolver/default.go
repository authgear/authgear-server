package resolver

import (
	"context"
	"net/http"

	"github.com/skygeario/skygear-server/pkg/core/auth/role"

	"github.com/skygeario/skygear-server/pkg/core/auth"
	"github.com/skygeario/skygear-server/pkg/core/auth/authinfo"
	"github.com/skygeario/skygear-server/pkg/core/auth/authn"
	"github.com/skygeario/skygear-server/pkg/core/auth/authtoken"
	"github.com/skygeario/skygear-server/pkg/core/config"
	skyContext "github.com/skygeario/skygear-server/pkg/core/handler/context"
	"github.com/skygeario/skygear-server/pkg/core/model"
)

type AuthContextResolverFactory struct{}

func (f AuthContextResolverFactory) NewResolver(ctx context.Context, tenantConfig config.TenantConfiguration) authn.AuthContextResolver {
	r := &DefaultAuthContextResolver{
		TokenStore:    auth.NewDefaultTokenStore(ctx, tenantConfig),
		AuthInfoStore: auth.NewDefaultAuthInfoStore(ctx, tenantConfig),
		RoleStore:     auth.NewDefaultRoleStore(ctx, tenantConfig),
	}
	return r
}

type DefaultAuthContextResolver struct {
	TokenStore    authtoken.Store
	AuthInfoStore authinfo.Store
	RoleStore     role.Store
}

func (r DefaultAuthContextResolver) Resolve(req *http.Request) (ctx skyContext.AuthContext, err error) {
	keyType := model.GetAccessKeyType(req)

	var resolver authn.AuthContextResolver
	if keyType == model.MasterAccessKey {
		resolver = masterkeyAuthContextResolver{
			TokenStore:    r.TokenStore,
			AuthInfoStore: r.AuthInfoStore,
		}
	} else {
		resolver = nonMasterkeyAuthContextResolver{
			TokenStore:    r.TokenStore,
			AuthInfoStore: r.AuthInfoStore,
		}
	}

	ctx, err = resolver.Resolve(req)
	ctx.AccessKeyType = keyType

	if ctx.AuthInfo != nil {
		ctx.Roles, err = r.RoleStore.QueryRoles(ctx.AuthInfo.Roles)
	}
	return
}
