package idpsession

import (
	"net/http"
	"net/http/httptest"
	"testing"

	. "github.com/smartystreets/goconvey/convey"

	"github.com/authgear/authgear-server/pkg/lib/session"
	"github.com/authgear/authgear-server/pkg/util/clock"
	"github.com/authgear/authgear-server/pkg/util/httputil"
)

type mockResolverProvider struct {
	Sessions []IDPSession
}

func (r *mockResolverProvider) GetByToken(token string) (*IDPSession, error) {
	for _, s := range r.Sessions {
		if s.TokenHash == token {
			return &s, nil
		}
	}
	return nil, ErrSessionNotFound
}

func (r *mockResolverProvider) Update(session *IDPSession) error {
	for i, s := range r.Sessions {
		if s.ID == session.ID {
			r.Sessions[i] = *session
			break
		}
	}
	return nil
}

func TestResolver(t *testing.T) {
	Convey("Resolver", t, func() {
		cookie := session.CookieDef{
			Def: &httputil.CookieDef{
				Name:   "session",
				Path:   "/",
				Domain: "app.test",
				MaxAge: nil,
			},
		}
		provider := &mockResolverProvider{}
		provider.Sessions = []IDPSession{
			{
				ID: "session-id",
				Attrs: session.Attrs{
					UserID: "user-id",
				},
				TokenHash: "token",
			},
		}

		resolver := Resolver{
			Cookie:     cookie,
			Provider:   provider,
			TrustProxy: true,
			Clock:      clock.NewMockClock(),
		}

		Convey("resolve without session cookie", func() {
			rw := httptest.NewRecorder()
			r, _ := http.NewRequest("POST", "/", nil)
			session, err := resolver.Resolve(rw, r)

			So(session, ShouldBeNil)
			So(err, ShouldBeNil)
		})

		Convey("resolve with invalid session cookie", func() {
			rw := httptest.NewRecorder()
			r, _ := http.NewRequest("POST", "/", nil)
			r.AddCookie(&http.Cookie{Name: "session", Value: "invalid"})
			s, err := resolver.Resolve(rw, r)

			So(s, ShouldBeNil)
			So(err, ShouldBeError, session.ErrInvalidSession)
		})

		Convey("resolve with valid session cookie", func() {
			rw := httptest.NewRecorder()
			r, _ := http.NewRequest("POST", "/", nil)
			r.AddCookie(&http.Cookie{Name: "session", Value: "token"})
			session, err := resolver.Resolve(rw, r)

			So(session, ShouldNotBeNil)
			So(session.SessionID(), ShouldEqual, "session-id")
			So(err, ShouldBeNil)
		})
	})
}
