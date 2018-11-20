package handler

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/skygeario/skygear-server/pkg/auth/dependency/provider/password"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/userprofile"

	"github.com/skygeario/skygear-server/pkg/core/auth/authinfo"
	"github.com/skygeario/skygear-server/pkg/core/db"
	"github.com/skygeario/skygear-server/pkg/core/handler"
	. "github.com/skygeario/skygear-server/pkg/server/skytest"
	. "github.com/smartystreets/goconvey/convey"
)

func TestForgotPasswordHandler(t *testing.T) {
	getSender := func() *MockForgotPasswordEmailSender {
		return &MockForgotPasswordEmailSender{}
	}

	Convey("Test ForgotPasswordHandler", t, func() {
		fh := &ForgotPasswordHandler{}
		fh.TxContext = db.NewMockTxContext()
		authRecordKeys := [][]string{[]string{"email", "username"}}
		fh.PasswordAuthProvider = password.NewMockProviderWithPrincipalMap(
			authRecordKeys,
			map[string]password.Principal{
				"john.doe.principal.id": password.Principal{
					ID:     "john.doe.principal.id",
					UserID: "john.doe.id",
					AuthData: map[string]interface{}{
						"username": "john.doe",
						"email":    "john.doe@example.com",
					},
					HashedPassword: []byte("$2a$10$/jm/S1sY6ldfL6UZljlJdOAdJojsJfkjg/pqK47Q8WmOLE19tGWQi"), // 123456
				},
			},
		)
		fh.AuthInfoStore = authinfo.NewMockStoreWithAuthInfoMap(
			map[string]authinfo.AuthInfo{
				"john.doe.id": authinfo.AuthInfo{
					ID: "john.doe.id",
				},
			},
		)
		userProfileStore := userprofile.NewMockUserProfileStore()
		userProfileStore.Data["john.doe.id"] = map[string]interface{}{
			"username": "john.doe",
			"email":    "john.doe@example.com",
		}
		fh.UserProfileStore = userProfileStore
		sender := getSender()
		fh.ForgotPasswordEmailSender = sender

		Convey("send email to user", func() {
			req, _ := http.NewRequest("POST", "", strings.NewReader(`{
				"email": "john.doe@example.com"
			}`))
			resp := httptest.NewRecorder()
			h := handler.APIHandlerToHandler(fh, fh.TxContext)
			h.ServeHTTP(resp, req)
			So(resp.Body.Bytes(), ShouldEqualJSON, `{
				"result": "OK"
			}`)
			So(sender.lastEmail, ShouldEqual, "john.doe@example.com")
			So(sender.lastUserProfile.ID, ShouldEqual, "user/john.doe.id")
			So(sender.lastAuthInfo.ID, ShouldEqual, "john.doe.id")
			So(sender.lastHashedPassword, ShouldResemble, []byte("$2a$10$/jm/S1sY6ldfL6UZljlJdOAdJojsJfkjg/pqK47Q8WmOLE19tGWQi"))
		})

		Convey("throw error for empty email", func() {
			req, _ := http.NewRequest("POST", "", strings.NewReader(`{
				"email": ""
			}`))
			resp := httptest.NewRecorder()
			h := handler.APIHandlerToHandler(fh, fh.TxContext)
			h.ServeHTTP(resp, req)
			So(resp.Body.Bytes(), ShouldEqualJSON, `{
				"error": {
					"code": 108,
					"info": {
						"arguments": ["email"]
					},
					"message": "empty email",
					"name": "InvalidArgument"
				}
			}`)
		})

		Convey("throw error for unknown email", func() {
			req, _ := http.NewRequest("POST", "", strings.NewReader(`{
				"email": "iamyourfather@example.com"
			}`))
			resp := httptest.NewRecorder()
			h := handler.APIHandlerToHandler(fh, fh.TxContext)
			h.ServeHTTP(resp, req)
			So(resp.Body.Bytes(), ShouldEqualJSON, `{
				"error": {
					"code": 110,
					"message": "user not found",
					"name": "ResourceNotFound"
				}
			}`)
		})
	})
}

type MockForgotPasswordEmailSender struct {
	lastEmail          string
	lastAuthInfo       authinfo.AuthInfo
	lastUserProfile    userprofile.UserProfile
	lastHashedPassword []byte
}

func (m *MockForgotPasswordEmailSender) Send(
	email string,
	authInfo authinfo.AuthInfo,
	userProfile userprofile.UserProfile,
	hashedPassword []byte,
) (err error) {
	m.lastEmail = email
	m.lastAuthInfo = authInfo
	m.lastUserProfile = userProfile
	m.lastHashedPassword = hashedPassword
	return nil
}
