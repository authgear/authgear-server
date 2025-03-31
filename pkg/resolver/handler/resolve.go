package handler

import (
	"context"
	"net/http"

	"github.com/authgear/authgear-server/pkg/api/model"
	"github.com/authgear/authgear-server/pkg/lib/session"
	"github.com/authgear/authgear-server/pkg/lib/userinfo"
	"github.com/authgear/authgear-server/pkg/util/httproute"
	"github.com/authgear/authgear-server/pkg/util/log"
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

type ResolveHandlerLogger struct{ *log.Logger }

func NewResolveHandlerLogger(lf *log.Factory) ResolveHandlerLogger {
	return ResolveHandlerLogger{lf.New("resolve-handler")}
}

type UserInfoService interface {
	GetUserInfoGreatest(ctx context.Context, userID string) (*userinfo.UserInfo, error)
}

type ResolveHandler struct {
	Database        Database
	Logger          ResolveHandlerLogger
	UserInfoService UserInfoService
}

func (h *ResolveHandler) ServeHTTP(rw http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	_ = h.Database.ReadOnly(ctx, func(ctx context.Context) error {
		return h.Handle(ctx, rw, r)
	})
}

func (h *ResolveHandler) Handle(ctx context.Context, rw http.ResponseWriter, r *http.Request) (err error) {
	info, err := h.resolve(ctx, r)
	if err != nil {
		h.Logger.WithError(err).Error("failed to resolve user")

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
