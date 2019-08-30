package resolver

import (
	"context"
	"net/http"

	"github.com/skygeario/skygear-server/pkg/core/auth"
	"github.com/skygeario/skygear-server/pkg/core/auth/authinfo"
	"github.com/skygeario/skygear-server/pkg/core/auth/authn"
	"github.com/skygeario/skygear-server/pkg/core/auth/session"
	"github.com/skygeario/skygear-server/pkg/core/config"
	"github.com/skygeario/skygear-server/pkg/core/model"
)

type AuthContextResolverFactory struct{}

func (f AuthContextResolverFactory) NewResolver(ctx context.Context, tenantConfig config.TenantConfiguration) authn.AuthContextResolver {
	r := &DefaultAuthContextResolver{
		SessionProvider: auth.NewDefaultSessionProvider(ctx, tenantConfig),
		AuthInfoStore:   auth.NewDefaultAuthInfoStore(ctx, tenantConfig),
	}
	return r
}

type DefaultAuthContextResolver struct {
	SessionProvider session.Provider
	AuthInfoStore   authinfo.Store
}

func (r DefaultAuthContextResolver) Resolve(req *http.Request, ctx auth.ContextSetter) (err error) {
	keyType := model.GetAccessKeyType(req)
	ctx.SetAccessKeyType(keyType)

	token := model.GetAccessToken(req)
	s, err := r.SessionProvider.GetByToken(token, session.TokenKindAccessToken)
	if err != nil {
		if err == session.ErrSessionNotFound {
			err = nil
		}
		return
	}

	info := &authinfo.AuthInfo{}
	err = r.AuthInfoStore.GetAuth(s.UserID, info)
	if err != nil {
		return
	}

	ctx.SetAuthInfo(info)
	ctx.SetSession(s)

	return
}

// this ensures that our structure conform to certain interfaces.
var (
	_ authn.AuthContextResolver = &DefaultAuthContextResolver{}
)
