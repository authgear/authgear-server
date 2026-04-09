package transport

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	. "github.com/smartystreets/goconvey/convey"

	"github.com/authgear/authgear-server/pkg/api/model"
	portalconfig "github.com/authgear/authgear-server/pkg/portal/config"
	portalmodel "github.com/authgear/authgear-server/pkg/portal/model"
	"github.com/authgear/authgear-server/pkg/portal/service"
	"github.com/authgear/authgear-server/pkg/portal/session"
)

type mockCollaboratorService struct {
	err  error
	coll *portalmodel.Collaborator
}

func (m *mockCollaboratorService) GetCollaboratorByAppAndUser(ctx context.Context, appID string, userID string) (*portalmodel.Collaborator, error) {
	return m.coll, m.err
}

func TestAuthzMiddleware(t *testing.T) {
	appID := "test-app"
	cfg := &portalconfig.AuthgearConfig{AppID: appID}

	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	newRequest := func(ctx context.Context) *http.Request {
		r := httptest.NewRequest("GET", "/api/v1/apps", nil)
		return r.WithContext(ctx)
	}

	Convey("AuthzMiddleware", t, func() {
		Convey("returns 401 when there is no session", func() {
			m := &AuthzMiddleware{
				AuthgearConfig: cfg,
				Collaborators:  &mockCollaboratorService{},
			}
			ctx := context.Background() // no session info
			w := httptest.NewRecorder()
			m.Handle(next).ServeHTTP(w, newRequest(ctx))
			So(w.Code, ShouldEqual, http.StatusUnauthorized)
		})

		Convey("returns 403 when user is not a collaborator", func() {
			m := &AuthzMiddleware{
				AuthgearConfig: cfg,
				Collaborators:  &mockCollaboratorService{err: service.ErrCollaboratorNotFound},
			}
			ctx := session.WithSessionInfo(context.Background(), &model.SessionInfo{
				IsValid: true,
				UserID:  "user-1",
			})
			w := httptest.NewRecorder()
			m.Handle(next).ServeHTTP(w, newRequest(ctx))
			So(w.Code, ShouldEqual, http.StatusForbidden)
		})

		Convey("returns 500 when the collaborator lookup fails", func() {
			m := &AuthzMiddleware{
				AuthgearConfig: cfg,
				Collaborators:  &mockCollaboratorService{err: errors.New("db error")},
			}
			ctx := session.WithSessionInfo(context.Background(), &model.SessionInfo{
				IsValid: true,
				UserID:  "user-1",
			})
			w := httptest.NewRecorder()
			m.Handle(next).ServeHTTP(w, newRequest(ctx))
			So(w.Code, ShouldEqual, http.StatusInternalServerError)
		})

		Convey("calls next handler when user is a collaborator", func() {
			m := &AuthzMiddleware{
				AuthgearConfig: cfg,
				Collaborators: &mockCollaboratorService{
					coll: &portalmodel.Collaborator{AppID: appID, UserID: "user-1"},
				},
			}
			ctx := session.WithSessionInfo(context.Background(), &model.SessionInfo{
				IsValid: true,
				UserID:  "user-1",
			})
			w := httptest.NewRecorder()
			m.Handle(next).ServeHTTP(w, newRequest(ctx))
			So(w.Code, ShouldEqual, http.StatusOK)
		})
	})
}
