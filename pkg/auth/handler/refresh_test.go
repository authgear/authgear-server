package handler

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/skygeario/skygear-server/pkg/core/db"
	. "github.com/skygeario/skygear-server/pkg/core/skytest"
	. "github.com/smartystreets/goconvey/convey"

	"github.com/skygeario/skygear-server/pkg/core/auth"
	"github.com/skygeario/skygear-server/pkg/core/auth/session"
)

func TestRefreshHandler(t *testing.T) {
	Convey("RefreshHandler", t, func() {
		h := &RefreshHandler{}

		h.TxContext = db.NewMockTxContext()
		sessionProvider := session.NewMockProvider()
		h.SessionProvider = sessionProvider
		h.SessionWriter = session.NewMockWriter()

		now := time.Date(2006, 1, 1, 0, 0, 0, 0, time.UTC)
		session := &auth.Session{
			ID:                   "session-id",
			ClientID:             "web-app",
			UserID:               "user-id",
			PrincipalID:          "principal-id",
			CreatedAt:            now,
			AccessedAt:           now,
			AccessTokenHash:      "access-token",
			RefreshTokenHash:     "refresh-token",
			AccessTokenCreatedAt: now,
		}
		sessionProvider.Sessions[session.ID] = *session

		Convey("should reject invalid token", func() {
			req, _ := http.NewRequest("POST", "", strings.NewReader(`
			{
				"refresh_token": ""
			}`))
			req.Header.Set("Content-Type", "application/json")
			resp := httptest.NewRecorder()
			h.ServeHTTP(resp, req)

			So(resp.Code, ShouldEqual, 401)
			So(resp.Body.Bytes(), ShouldEqualJSON, `{
				"error": {
					"name": "Unauthorized",
					"reason": "NotAuthenticated",
					"message": "authentication required",
					"code": 401
				}
			}`)
		})

		Convey("should refresh access token", func() {
			req, _ := http.NewRequest("POST", "", strings.NewReader(`
			{
				"refresh_token": "refresh-token"
			}`))
			req.Header.Set("Content-Type", "application/json")
			resp := httptest.NewRecorder()
			h.ServeHTTP(resp, req)

			So(resp.Code, ShouldEqual, 200)
			So(resp.Body.Bytes(), ShouldEqualJSON, `{
				"result": {
					"access_token": "access-token-session-id-0"
				}
			}`)

			So(sessionProvider.Sessions[session.ID].AccessTokenHash, ShouldEqual, "access-token-session-id-0")
		})
	})
}
