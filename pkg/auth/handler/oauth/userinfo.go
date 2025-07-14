package oauth

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/lib/infra/db/appdb"
	"github.com/authgear/authgear-server/pkg/lib/oauth"
	"github.com/authgear/authgear-server/pkg/lib/session"
	"github.com/authgear/authgear-server/pkg/util/httproute"
	"github.com/authgear/authgear-server/pkg/util/slogutil"
)

func ConfigureUserInfoRoute(route httproute.Route) httproute.Route {
	return route.
		WithMethods("GET", "POST", "OPTIONS").
		WithPathPattern("/oauth2/userinfo")
}

type ProtocolUserInfoProvider interface {
	GetUserInfo(ctx context.Context, userID string, clientLike *oauth.ClientLike) (map[string]interface{}, error)
}

var UserInfoHandlerLogger = slogutil.NewLogger("handler-user-info")

type OAuthClientResolver interface {
	ResolveClient(clientID string) *config.OAuthClientConfig
}

type UserInfoHandler struct {
	Database            *appdb.Handle
	UserInfoProvider    ProtocolUserInfoProvider
	OAuth               *config.OAuthConfig
	OAuthClientResolver OAuthClientResolver
}

func (h *UserInfoHandler) ServeHTTP(rw http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	s := session.GetSession(ctx)
	clientLike := oauth.SessionClientLike(s, h.OAuthClientResolver)
	var userInfo map[string]interface{}
	err := h.Database.WithTx(ctx, func(ctx context.Context) (err error) {
		userInfo, err = h.UserInfoProvider.GetUserInfo(ctx, s.GetAuthenticationInfo().UserID, clientLike)
		return
	})

	if err != nil {
		logger := UserInfoHandlerLogger.GetLogger(ctx)
		logger.WithError(err).Error(ctx, "oidc userinfo handler failed")
		http.Error(rw, "Internal Server Error", 500)
		return
	}

	rw.Header().Set("Content-Type", "application/json")

	encoder := json.NewEncoder(rw)
	err = encoder.Encode(userInfo)
	if err != nil {
		http.Error(rw, err.Error(), 500)
	}
}
