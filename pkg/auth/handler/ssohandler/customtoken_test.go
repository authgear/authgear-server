package ssohandler

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/dgrijalva/jwt-go"
	. "github.com/smartystreets/goconvey/convey"

	"github.com/skygeario/skygear-server/pkg/auth/dependency/provider/customtoken"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/userprofile"
	"github.com/skygeario/skygear-server/pkg/core/audit"
	"github.com/skygeario/skygear-server/pkg/core/auth/authinfo"
	"github.com/skygeario/skygear-server/pkg/core/auth/authtoken"
	"github.com/skygeario/skygear-server/pkg/core/auth/role"
	"github.com/skygeario/skygear-server/pkg/core/db"
	"github.com/skygeario/skygear-server/pkg/core/handler"
	. "github.com/skygeario/skygear-server/pkg/server/skytest"
)

func TestCustomTokenLoginHandler(t *testing.T) {
	realTime := timeNow
	timeNow = func() time.Time { return zeroTime }
	defer func() {
		timeNow = realTime
	}()

	Convey("Test CustomTokenLoginHandler", t, func() {
		mockTokenStore := authtoken.NewMockStore()
		lh := &CustomTokenLoginHandler{}
		lh.TxContext = db.NewMockTxContext()
		lh.CustomTokenAuthProvider = customtoken.NewMockProvider("ssosecret")
		lh.AuthInfoStore = authinfo.NewMockStore()
		lh.UserProfileStore = userprofile.NewMockUserProfileStore()
		lh.TokenStore = mockTokenStore
		lh.RoleStore = role.NewMockStore()
		lh.AuditTrail = audit.NewMockTrail(t)
		h := handler.APIHandlerToHandler(lh, lh.TxContext)

		Convey("create user account with custom token", func(c C) {
			tokenString, err := jwt.NewWithClaims(
				jwt.SigningMethodHS256,
				customtoken.SSOCustomTokenClaims{
					StandardClaims: jwt.StandardClaims{
						IssuedAt:  time.Now().Unix(),
						ExpiresAt: time.Now().Add(time.Hour * 1).Unix(),
						Subject:   "otherid1",
					},
					RawProfile: map[string]interface{}{
						"name":  "John Doe",
						"email": "John@skygear.io",
					},
				},
			).SignedString([]byte("ssosecret"))
			So(err, ShouldBeNil)

			req, _ := http.NewRequest("POST", "", strings.NewReader(fmt.Sprintf(`
			{
				"token": "%s"
			}`, tokenString)))
			resp := httptest.NewRecorder()
			h.ServeHTTP(resp, req)

			p, _ := lh.CustomTokenAuthProvider.GetPrincipalByTokenPrincipalID("otherid1")
			token := mockTokenStore.GetTokensByAuthInfoID(p.UserID)[0]

			So(resp.Code, ShouldEqual, 200)
			So(resp.Body.Bytes(), ShouldEqualJSON, fmt.Sprintf(`{
				"result": {
					"user_id": "%s",
					"profile": {
						"_access": null,
						"_created_at": "0001-01-01T00:00:00Z",
						"_created_by": "%s",
						"_id": "user/%s",
						"_ownerID": "%s",
						"_recordID": "%s",
						"_recordType": "user",
						"_type": "record",
						"_updated_at": "0001-01-01T00:00:00Z",
						"_updated_by": "%s",
						"email": "John@skygear.io",
						"name": "John Doe"
					},
					"access_token": "%s"
				}
			}`,
				p.UserID,
				p.UserID,
				p.UserID,
				p.UserID,
				p.UserID,
				p.UserID,
				token.AccessToken))

			mockTrail, _ := lh.AuditTrail.(*audit.MockTrail)
			So(mockTrail.Hook.LastEntry().Message, ShouldEqual, "audit_trail")
			So(mockTrail.Hook.LastEntry().Data["event"], ShouldEqual, "signup")
		})

		Convey("update user account with custom token", func(c C) {
			tokenString, err := jwt.NewWithClaims(
				jwt.SigningMethodHS256,
				customtoken.SSOCustomTokenClaims{
					StandardClaims: jwt.StandardClaims{
						IssuedAt:  time.Now().Unix(),
						ExpiresAt: time.Now().Add(time.Hour * 1).Unix(),
						Subject:   "otherid1",
					},
					RawProfile: map[string]interface{}{
						"name":  "John Doe",
						"email": "John@skygear.io",
					},
				},
			).SignedString([]byte("ssosecret"))
			So(err, ShouldBeNil)

			req, _ := http.NewRequest("POST", "", strings.NewReader(fmt.Sprintf(`
			{
				"token": "%s"
			}`, tokenString)))
			resp := httptest.NewRecorder()
			h.ServeHTTP(resp, req)

			p, _ := lh.CustomTokenAuthProvider.GetPrincipalByTokenPrincipalID("otherid1")
			profile, _ := lh.UserProfileStore.GetUserProfile(p.UserID, "")

			So(profile.Data, ShouldResemble, userprofile.Data{
				"name":  "John Doe",
				"email": "John@skygear.io",
			})
		})

		Convey("check whether token is invalid", func(c C) {
			tokenString, err := jwt.NewWithClaims(
				jwt.SigningMethodHS256,
				customtoken.SSOCustomTokenClaims{
					StandardClaims: jwt.StandardClaims{
						IssuedAt:  time.Now().Add(-time.Hour * 1).Unix(),
						ExpiresAt: time.Now().Add(-time.Minute * 30).Unix(),
						Subject:   "otherid1",
					},
				},
			).SignedString([]byte("ssosecret"))
			So(err, ShouldBeNil)

			req, _ := http.NewRequest("POST", "", strings.NewReader(fmt.Sprintf(`
			{
				"token": "%s"
			}`, tokenString)))
			resp := httptest.NewRecorder()
			h.ServeHTTP(resp, req)

			c.Printf("Response: %s", string(resp.Body.Bytes()))
			So(resp.Code, ShouldEqual, 400)

			mockTrail, _ := lh.AuditTrail.(*audit.MockTrail)
			So(mockTrail.Hook.LastEntry().Message, ShouldEqual, "audit_trail")
			So(mockTrail.Hook.LastEntry().Data["event"], ShouldEqual, "login_failure")
		})
	})
}
