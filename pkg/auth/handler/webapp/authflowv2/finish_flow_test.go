package authflowv2

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	handlerwebapp "github.com/authgear/authgear-server/pkg/auth/handler/webapp"
	"github.com/authgear/authgear-server/pkg/auth/webapp"
	authflow "github.com/authgear/authgear-server/pkg/lib/authenticationflow"
	"github.com/authgear/authgear-server/pkg/util/httputil"

	. "github.com/smartystreets/goconvey/convey"
)

type finishFlowTestSessionStore struct {
	updateCount int
}

func (*finishFlowTestSessionStore) Get(ctx context.Context, id string) (*webapp.Session, error) {
	return nil, errors.New("unexpected Get")
}

func (*finishFlowTestSessionStore) Create(ctx context.Context, session *webapp.Session) error {
	return errors.New("unexpected Create")
}

func (s *finishFlowTestSessionStore) Update(ctx context.Context, session *webapp.Session) error {
	s.updateCount++
	return nil
}

func (*finishFlowTestSessionStore) Delete(ctx context.Context, id string) error {
	return errors.New("unexpected Delete")
}

func TestAuthflowV2FinishFlowHandlerAllowsCompletedSession(t *testing.T) {
	Convey("Finish flow", t, func() {
		session := &webapp.Session{
			ID:          "web_session_id",
			RedirectURI: "/after",
			Authflow: &webapp.Authflow{
				AllScreens: map[string]*webapp.AuthflowScreen{
					"step_0": {
						StateToken: &webapp.AuthflowStateToken{
							XStep:      "step_0",
							StateToken: "state_token",
						},
						FinishedUIScreenData: &webapp.AuthflowFinishedUIScreenData{
							FlowType: authflow.FlowTypeLogin,
						},
					},
				},
			},
		}

		req := httptest.NewRequest("GET", "http://example.com/authflow/v2/finish?x_step=step_0", nil)
		req = req.WithContext(webapp.WithSession(req.Context(), session))

		sessionStore := &finishFlowTestSessionStore{}
		controller := &handlerwebapp.AuthflowController{
			Sessions:       sessionStore,
			Cookies:        &httputil.CookieManager{Request: req},
			SignedUpCookie: webapp.NewSignedUpCookieDef(),
		}
		h := &AuthflowV2FinishFlowHandler{
			Controller: controller,
		}

		first := httptest.NewRecorder()
		h.ServeHTTP(first, req)
		So(first.Code, ShouldEqual, http.StatusFound)
		So(first.Header().Get("Location"), ShouldContainSubstring, "/after")
		So(session.IsCompleted, ShouldBeTrue)
		So(sessionStore.updateCount, ShouldEqual, 1)

		second := httptest.NewRecorder()
		h.ServeHTTP(second, req)
		So(second.Code, ShouldEqual, http.StatusFound)
		So(second.Header().Get("Location"), ShouldContainSubstring, "/after")
		So(sessionStore.updateCount, ShouldEqual, 2)
	})
}
