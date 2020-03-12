package session

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/skygeario/skygear-server/pkg/core/auth/authinfo"
	"github.com/skygeario/skygear-server/pkg/core/db"
	. "github.com/smartystreets/goconvey/convey"
)

type mockResolver struct {
	Sessions        []Session
	AccessedSession []string
}

func (r *mockResolver) GetByToken(token string) (*Session, error) {
	for _, s := range r.Sessions {
		if s.TokenHash == token {
			return &s, nil
		}
	}
	return nil, ErrSessionNotFound
}

func (r *mockResolver) Access(s *Session) error {
	r.AccessedSession = append(r.AccessedSession, s.ID)
	return nil
}

func TestMiddleware(t *testing.T) {
	Convey("Middleware", t, func() {
		config := CookieConfiguration{
			Secure: true,
			MaxAge: nil,
			Domain: "app.test",
		}
		resolver := &mockResolver{}
		userStore := authinfo.NewMockStore()

		userStore.AuthInfoMap["user-id"] = authinfo.AuthInfo{
			ID:       "user-id",
			Verified: true,
		}
		resolver.Sessions = []Session{
			{
				ID: "session-id",
				Attrs: Attrs{
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
		var ctx *Context
		handler := m.Handle(http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
			ctx = GetContext(r.Context())
		}))

		Convey("resolve without session cookie", func() {
			rw := httptest.NewRecorder()
			r, _ := http.NewRequest("POST", "/", nil)
			handler.ServeHTTP(rw, r)

			So(ctx, ShouldBeNil)
			So(rw.Result().Cookies(), ShouldBeEmpty)
		})

		Convey("resolve with invalid session cookie", func() {
			rw := httptest.NewRecorder()
			r, _ := http.NewRequest("POST", "/", nil)
			r.AddCookie(&http.Cookie{Name: CookieName, Value: "invalid"})
			handler.ServeHTTP(rw, r)

			So(ctx, ShouldResemble, &Context{User: nil, Session: nil})
			So(rw.Result().Cookies(), ShouldHaveLength, 1)
			So(rw.Result().Cookies()[0].Raw, ShouldEqual, "session=; Path=/; Domain=app.test; Expires=Thu, 01 Jan 1970 00:00:00 GMT; HttpOnly; Secure; SameSite=Lax")
		})

		Convey("resolve with valid session cookie", func() {
			rw := httptest.NewRecorder()
			r, _ := http.NewRequest("POST", "/", nil)
			r.AddCookie(&http.Cookie{Name: CookieName, Value: "token"})
			handler.ServeHTTP(rw, r)

			So(ctx, ShouldNotBeNil)
			So(ctx.User.ID, ShouldEqual, "user-id")
			So(ctx.Session.ID, ShouldEqual, "session-id")
			So(rw.Result().Cookies(), ShouldBeEmpty)
		})
	})
}
