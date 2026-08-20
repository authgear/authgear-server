package handler

import (
	"context"
	"net/http"

	"github.com/authgear/authgear-server/pkg/api/model"
	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/lib/oauth"
	"github.com/authgear/authgear-server/pkg/lib/session"
	"github.com/authgear/authgear-server/pkg/lib/session/idpsession"
	"github.com/authgear/authgear-server/pkg/lib/userinfo"
	"github.com/authgear/authgear-server/pkg/util/httproute"
	"github.com/authgear/authgear-server/pkg/util/slogutil"
)

func ConfigureResolveRoute(route httproute.Route) []httproute.Route {
	route = route.WithMethods("HEAD", "GET")
	return []httproute.Route{
		route.WithPathPattern("/resolve"),
		route.WithPathPattern("/_resolver/resolve"),
	}
}

//go:generate go tool mockgen -source=resolve.go -destination=resolve_mock_test.go -package handler

type Database interface {
	ReadOnly(ctx context.Context, do func(ctx context.Context) error) error
}

var ResolveHandlerLogger = slogutil.NewLogger("resolve-handler")

type UserInfoService interface {
	GetUserInfoGreatest(ctx context.Context, userID string) (*userinfo.UserInfo, error)
}

// OAuthClientResolver is needed to tell first-party from third-party
// clients for the opaque-token gate below (docs/plans/dcr/2026-08-17-04-resource-access-policy.md
// §5.6).
type OAuthClientResolver interface {
	ResolveClient(ctx context.Context, clientID string) *config.OAuthClientConfig
}

type ResolveHandler struct {
	Database            Database
	UserInfoService     UserInfoService
	OAuthClientResolver OAuthClientResolver
}

func (h *ResolveHandler) ServeHTTP(rw http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	_ = h.Database.ReadOnly(ctx, func(ctx context.Context) error {
		return h.Handle(ctx, rw, r)
	})
}

func (h *ResolveHandler) Handle(ctx context.Context, rw http.ResponseWriter, r *http.Request) (err error) {
	logger := ResolveHandlerLogger.GetLogger(ctx)
	info, err := h.resolve(ctx, r)
	if err != nil {
		logger.WithError(err).Error(ctx, "failed to resolve user")

		http.Error(rw, "internal error", http.StatusInternalServerError)
		return
	}
	if info != nil {
		info.PopulateHeaders(rw)
	}

	return
}

func (h *ResolveHandler) resolve(ctx context.Context, r *http.Request) (*model.SessionInfo, error) {
	valid := session.HasValidSession(ctx)
	userID := session.GetUserID(ctx)
	s := session.GetSession(ctx)

	var info *model.SessionInfo
	if valid && userID != nil && s != nil {
		// A third-party client's opaque access token must not work here: it
		// is scoped to the userinfo endpoint only (dcr.md, access-token-audience-binding.md).
		// Only an OfflineGrantSession resolved via a bearer access token
		// (TokenTypeOpaque) is subject to this gate; every other case --
		// an IDP session, or an OfflineGrantSession resolved via the app
		// session token cookie (TokenTypeAppSession) or a resource-bound
		// JWT (TokenTypeJWT) -- is unaffected regardless of client kind.
		switch sess := s.(type) {
		case *idpsession.IDPSession:
			// Always accept, matching oauth.SessionClientLike's existing
			// treatment of every IDPSession as first-party -- see
			// idpsession.IDPSession.GetTokenType's doc comment for why an
			// IDPSession can never back a third-party client's access grant.
		case *oauth.OfflineGrantSession:
			switch sess.GetTokenType() {
			case session.TokenTypeJWT, session.TokenTypeAppSession:
				// Always accept.
			case session.TokenTypeOpaque:
				clientLike := oauth.SessionClientLike(ctx, s, h.OAuthClientResolver)
				if !clientLike.IsFirstParty {
					return &model.SessionInfo{IsValid: false}, nil
				}
			default:
				// Every path that hands out an OfflineGrantSession as a
				// ResolvedSession sets TokenType; an unset value means a new
				// resolution path was added without setting it. Fail closed.
				return &model.SessionInfo{IsValid: false}, nil
			}
		default:
			return &model.SessionInfo{IsValid: false}, nil
		}

		userInfo, err := h.UserInfoService.GetUserInfoGreatest(ctx, *userID)
		if err != nil {
			return nil, err
		}

		info = session.NewInfo(
			s,
			userInfo.User.IsAnonymous,
			userInfo.User.IsVerified,
			userInfo.User.CanReauthenticate,
			userInfo.EffectiveRoleKeys,
		)
	} else if !valid {
		info = &model.SessionInfo{IsValid: false}
	}

	return info, nil
}
