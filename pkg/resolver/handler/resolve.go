package handler

import (
	"context"
	"net/http"

	"github.com/authgear/authgear-server/pkg/api/model"
	"github.com/authgear/authgear-server/pkg/lib/authn/identity"
	"github.com/authgear/authgear-server/pkg/lib/session"
	"github.com/authgear/authgear-server/pkg/util/accesscontrol"
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

//go:generate mockgen -source=resolve.go -destination=resolve_mock_test.go -package handler

type IdentityService interface {
	ListByUser(ctx context.Context, userID string) ([]*identity.Info, error)
}

type VerificationService interface {
	IsUserVerified(ctx context.Context, identities []*identity.Info) (bool, error)
}

type Database interface {
	ReadOnly(ctx context.Context, do func(ctx context.Context) error) error
}

type ResolveHandlerLogger struct{ *log.Logger }

func NewResolveHandlerLogger(lf *log.Factory) ResolveHandlerLogger {
	return ResolveHandlerLogger{lf.New("resolve-handler")}
}

type UserProvider interface {
	Get(ctx context.Context, id string, role accesscontrol.Role) (*model.User, error)
}

type RolesAndGroupsProvider interface {
	ListEffectiveRolesByUserID(ctx context.Context, userID string) ([]*model.Role, error)
}

type ResolveHandler struct {
	Database       Database
	Identities     IdentityService
	Verification   VerificationService
	Logger         ResolveHandlerLogger
	Users          UserProvider
	RolesAndGroups RolesAndGroupsProvider
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
		identities, err := h.Identities.ListByUser(ctx, *userID)
		if err != nil {
			return nil, err
		}

		isAnonymous := false
		for _, i := range identities {
			if i.Type == model.IdentityTypeAnonymous {
				isAnonymous = true
				break
			}
		}

		isVerified, err := h.Verification.IsUserVerified(ctx, identities)
		if err != nil {
			return nil, err
		}

		user, err := h.Users.Get(ctx, *userID, accesscontrol.RoleGreatest)
		if err != nil {
			return nil, err
		}

		userCanReauthenticate := user.CanReauthenticate

		roles, err := h.RolesAndGroups.ListEffectiveRolesByUserID(ctx, *userID)
		roleKeys := make([]string, len(roles))
		for i := range roles {
			roleKeys[i] = roles[i].Key
		}
		if err != nil {
			return nil, err
		}

		info = session.NewInfo(s, isAnonymous, isVerified, userCanReauthenticate, roleKeys)
	} else if !valid {
		info = &model.SessionInfo{IsValid: false}
	}

	return info, nil
}
