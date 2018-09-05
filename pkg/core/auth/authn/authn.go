package authn

import (
	"context"
	"net/http"

	"github.com/skygeario/skygear-server/pkg/core/auth"
	"github.com/skygeario/skygear-server/pkg/core/config"
	"github.com/skygeario/skygear-server/pkg/core/handler"
	"github.com/skygeario/skygear-server/pkg/core/inject"
	"github.com/skygeario/skygear-server/pkg/server/skydb"

	"github.com/skygeario/skygear-server/pkg/server/authtoken"
)

type AuthInfoResolverFactory interface {
	NewResolver(context.Context, config.TenantConfiguration) AuthInfoResolver
}

type AuthInfoResolver interface {
	Resolve(*http.Request) (*authtoken.Token, *skydb.AuthInfo, error)
}

type StatefulJWTAuthInfoResolverFactory struct {
	handler.ProviderGraph
}

func (f StatefulJWTAuthInfoResolverFactory) NewResolver(ctx context.Context, tenantConfig config.TenantConfiguration) AuthInfoResolver {
	r := &StatefulJWTAuthInfoResolver{}
	inject.DefaultInject(r, f.ProviderGraph, ctx, tenantConfig)
	return r
}

type StatefulJWTAuthInfoResolver struct {
	auth.TokenStore    `dependency:"TokenStore"`
	auth.AuthInfoStore `dependency:"AuthInfoStore"`
}

func (r StatefulJWTAuthInfoResolver) Resolve(req *http.Request) (token *authtoken.Token, authInfo *skydb.AuthInfo, err error) {
	tokenStr := auth.GetAccessToken(req)

	token = &authtoken.Token{}
	err = r.TokenStore.Get(tokenStr, token)
	if err != nil {
		// TODO:
		// handle error properly
		token = nil
		return
	}

	authInfo = &skydb.AuthInfo{}
	err = r.AuthInfoStore.GetAuth(token.AuthInfoID, authInfo)
	if err != nil {
		// TODO:
		// handle error properly
		authInfo = nil
		return
	}

	return
}
