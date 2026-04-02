package transport

import (
	"context"
	"errors"
	"net/http"

	portalconfig "github.com/authgear/authgear-server/pkg/portal/config"
	"github.com/authgear/authgear-server/pkg/portal/model"
	"github.com/authgear/authgear-server/pkg/portal/service"
	"github.com/authgear/authgear-server/pkg/portal/session"
)

type AuthzCollaboratorService interface {
	GetCollaboratorByAppAndUser(ctx context.Context, appID string, userID string) (*model.Collaborator, error)
}

type AuthzMiddleware struct {
	AuthgearConfig *portalconfig.AuthgearConfig
	Collaborators  AuthzCollaboratorService
}

func (m *AuthzMiddleware) Handle(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		sessionInfo := session.GetValidSessionInfo(ctx)
		if sessionInfo == nil {
			writeError(w, r, service.ErrUnauthenticated)
			return
		}

		_, err := m.Collaborators.GetCollaboratorByAppAndUser(ctx, m.AuthgearConfig.AppID, sessionInfo.UserID)
		if errors.Is(err, service.ErrCollaboratorNotFound) {
			writeError(w, r, service.ErrForbidden)
			return
		} else if err != nil {
			writeError(w, r, err)
			return
		}

		next.ServeHTTP(w, r)
	})
}
