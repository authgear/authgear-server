package sso

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
	"github.com/skygeario/skygear-server/pkg/auth/task"
	"github.com/skygeario/skygear-server/pkg/core/async"
	"github.com/skygeario/skygear-server/pkg/core/audit"
	"github.com/skygeario/skygear-server/pkg/core/auth/authinfo"
	"github.com/skygeario/skygear-server/pkg/core/auth/authtoken"
	"github.com/skygeario/skygear-server/pkg/core/config"
	"github.com/skygeario/skygear-server/pkg/core/db"
	"github.com/skygeario/skygear-server/pkg/core/handler"
	. "github.com/skygeario/skygear-server/pkg/core/skytest"
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
		lh.CustomTokenAuthProvider = customtoken.NewMockProviderWithPrincipalMap("ssosecret", map[string]customtoken.Principal{
			"uuid-chima-token": customtoken.Principal{
				ID:               "uuid-chima-token",
				TokenPrincipalID: "chima.customtoken.id",
				UserID:           "chima",
			},
		})
		lh.AuthInfoStore = authinfo.NewMockStoreWithAuthInfoMap(
			map[string]authinfo.AuthInfo{
				"chima": authinfo.AuthInfo{
					ID: "chima",
				},
			},
		)
		userProfileStore := userprofile.NewMockUserProfileStore()
		userProfileStore.Data = map[string]map[string]interface{}{}
		userProfileStore.Data["chima"] = map[string]interface{}{
			"name":  "chima",
			"email": "chima@skygear.io",
		}
		userProfileStore.TimeNowfunc = timeNow
		lh.UserProfileStore = userProfileStore
		lh.TokenStore = mockTokenStore
		lh.AuditTrail = audit.NewMockTrail(t)
		lh.UserVerifyKeys = []config.UserVerificationKeyConfiguration{
			config.UserVerificationKeyConfiguration{
				Key: "email",
			},
		}
		lh.WelcomeEmailEnabled = true
		mockTaskQueue := async.NewMockQueue()
		lh.TaskQueue = mockTaskQueue
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
					"metadata": {
						"email":"John@skygear.io",
						"name":"John Doe"
					},
					"access_token": "%s",
					"verified": false,
					"verify_info": {},
					"created_at": "0001-01-01T00:00:00Z",
					"created_by": "%s",
					"updated_at": "0001-01-01T00:00:00Z",
					"updated_by": "%s"
				}
			}`,
				p.UserID,
				token.AccessToken,
				p.UserID,
				p.UserID))

			mockTrail, _ := lh.AuditTrail.(*audit.MockTrail)
			So(mockTrail.Hook.LastEntry().Message, ShouldEqual, "audit_trail")
			So(mockTrail.Hook.LastEntry().Data["event"], ShouldEqual, "signup")

			So(mockTaskQueue.TasksParam, ShouldHaveLength, 1)
			param, _ := mockTaskQueue.TasksParam[0].(task.WelcomeEmailSendTaskParam)
			So(param.Email, ShouldEqual, "John@skygear.io")
			So(param.User, ShouldNotBeNil)
			So(param.User.Metadata["name"], ShouldEqual, "John Doe")
			So(param.User.Metadata["email"], ShouldEqual, "John@skygear.io")
		})

		Convey("update user account with custom token", func(c C) {
			tokenString, err := jwt.NewWithClaims(
				jwt.SigningMethodHS256,
				customtoken.SSOCustomTokenClaims{
					StandardClaims: jwt.StandardClaims{
						IssuedAt:  time.Now().Unix(),
						ExpiresAt: time.Now().Add(time.Hour * 1).Unix(),
						Subject:   "chima.customtoken.id",
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

			p, _ := lh.CustomTokenAuthProvider.GetPrincipalByTokenPrincipalID("chima.customtoken.id")
			So(p.UserID, ShouldEqual, "chima")

			profile, _ := lh.UserProfileStore.GetUserProfile(p.UserID)
			So(profile.Data, ShouldResemble, userprofile.Data{
				"name":  "John Doe",
				"email": "John@skygear.io",
			})

			So(mockTaskQueue.TasksParam, ShouldHaveLength, 0)
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
