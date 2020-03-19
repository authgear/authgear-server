package session

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/skygeario/skygear-server/pkg/core/auth/authinfo"
	"github.com/skygeario/skygear-server/pkg/core/authn"
	"github.com/skygeario/skygear-server/pkg/core/db"
	. "github.com/smartystreets/goconvey/convey"
)

type mockResolver struct {
	Sessions        []IDPSession
	AccessedSession []string
}

func (r *mockResolver) GetByToken(token string) (*IDPSession, error) {
	for _, s := range r.Sessions {
		if s.TokenHash == token {
			return &s, nil
		}
	}
	return nil, ErrSessionNotFound
}

func (r *mockResolver) Access(s *IDPSession) error {
	r.AccessedSession = append(r.AccessedSession, s.ID)
	return nil
}

func TestMiddleware(t *testing.T) {
	Convey("Middleware", t, func() {
		config := CookieConfiguration{
			Name:   CookieName,
			Path:   "/",
			Domain: "app.test",
			Secure: true,
			MaxAge: nil,
		}
		resolver := &mockResolver{}
		userStore := authinfo.NewMockStore()

		userStore.AuthInfoMap["user-id"] = authinfo.AuthInfo{
			ID:       "user-id",
			Verified: true,
		}
		resolver.Sessions = []IDPSession{
			{
				ID: "session-id",
				Attrs: authn.Attrs{
					UserID: "user-id",
				},
				TokenHash: "token",
			},
		}

		m := Middleware{
			CookieConfiguration: config,
			SessionResolver:     resolver,
			AuthInfoStore:       userStore,
			TxContext:           db.NewMockTxContext(),
		}
		var valid bool
		var s authn.Session
		var u *authinfo.AuthInfo
		handler := m.Handle(http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
			valid = authn.IsValidAuthn(r.Context())
			s = authn.GetSession(r.Context())
			u = authn.GetUser(r.Context())
		}))

		Convey("resolve without session cookie", func() {
			rw := httptest.NewRecorder()
			r, _ := http.NewRequest("POST", "/", nil)
			handler.ServeHTTP(rw, r)

			So(valid, ShouldBeTrue)
			So(s, ShouldBeNil)
			So(u, ShouldBeNil)
			So(rw.Result().Cookies(), ShouldBeEmpty)
		})

		Convey("resolve with invalid session cookie", func() {
			rw := httptest.NewRecorder()
			r, _ := http.NewRequest("POST", "/", nil)
			r.AddCookie(&http.Cookie{Name: CookieName, Value: "invalid"})
			handler.ServeHTTP(rw, r)

			So(valid, ShouldBeFalse)
			So(s, ShouldBeNil)
			So(u, ShouldBeNil)
			So(rw.Result().Cookies(), ShouldHaveLength, 1)
			So(rw.Result().Cookies()[0].Raw, ShouldEqual, "session=; Path=/; Domain=app.test; Expires=Thu, 01 Jan 1970 00:00:00 GMT; HttpOnly; Secure; SameSite=Lax")
		})

		Convey("resolve with valid session cookie", func() {
			rw := httptest.NewRecorder()
			r, _ := http.NewRequest("POST", "/", nil)
			r.AddCookie(&http.Cookie{Name: CookieName, Value: "token"})
			handler.ServeHTTP(rw, r)

			So(valid, ShouldBeTrue)
			So(u.ID, ShouldEqual, "user-id")
			So(s.SessionID(), ShouldEqual, "session-id")
			So(rw.Result().Cookies(), ShouldBeEmpty)
		})
	})
}
