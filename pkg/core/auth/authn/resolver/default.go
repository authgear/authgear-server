package resolver

import (
	"context"
	"net/http"

	"github.com/skygeario/skygear-server/pkg/core/auth"
	"github.com/skygeario/skygear-server/pkg/core/auth/authinfo"
	"github.com/skygeario/skygear-server/pkg/core/auth/authn"
	"github.com/skygeario/skygear-server/pkg/core/auth/session"
	redisSession "github.com/skygeario/skygear-server/pkg/core/auth/session/redis"
	"github.com/skygeario/skygear-server/pkg/core/config"
	"github.com/skygeario/skygear-server/pkg/core/model"
)

type AuthContextResolverFactory struct{}

func (f AuthContextResolverFactory) NewResolver(ctx context.Context, tenantConfig config.TenantConfiguration) authn.AuthContextResolver {
	authCtx := auth.NewContextGetterWithContext(ctx)
	r := &DefaultAuthContextResolver{
		SessionProvider: session.NewProvider(redisSession.NewStore(ctx, tenantConfig.AppID), authCtx, tenantConfig.UserConfig.Clients),
		AuthInfoStore:   auth.NewDefaultAuthInfoStore(ctx, tenantConfig),
		ClientConfigs:   tenantConfig.UserConfig.Clients,
	}
	return r
}

type DefaultAuthContextResolver struct {
	SessionProvider session.Provider
	AuthInfoStore   authinfo.Store
	ClientConfigs   map[string]config.APIClientConfiguration
}

func (r DefaultAuthContextResolver) Resolve(req *http.Request, ctx auth.ContextSetter) (err error) {
	key := model.GetAccessKey(req)
	ctx.SetAccessKey(key)

	token, transport, err := model.GetAccessToken(req)
	if err != nil {
		if err == model.ErrTokenConflict {
			err = nil
		}
		return
	}

	s, err := r.SessionProvider.GetByToken(token, auth.SessionTokenKindAccessToken)
	if err != nil {
		if err == session.ErrSessionNotFound {
			err = nil
		}
		return
	}

	if r.ClientConfigs[s.ClientID].SessionTransport != transport {
		err = nil
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
