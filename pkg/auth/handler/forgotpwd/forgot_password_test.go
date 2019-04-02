package forgotpwd

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
	. "github.com/skygeario/skygear-server/pkg/core/skytest"
	. "github.com/smartystreets/goconvey/convey"
)

func TestForgotPasswordHandler(t *testing.T) {
	getSender := func() *MockForgotPasswordEmailSender {
		return &MockForgotPasswordEmailSender{}
	}

	Convey("Test ForgotPasswordHandler", t, func() {
		fh := &ForgotPasswordHandler{}
		fh.TxContext = db.NewMockTxContext()
		loginIDsKeyWhitelist := []string{}
		fh.PasswordAuthProvider = password.NewMockProviderWithPrincipalMap(
			loginIDsKeyWhitelist,
			map[string]password.Principal{
				"john.doe.principal.id": password.Principal{
					ID:             "john.doe.principal.id",
					UserID:         "john.doe.id",
					LoginIDKey:     "email",
					LoginID:        "john.doe@example.com",
					HashedPassword: []byte("$2a$10$/jm/S1sY6ldfL6UZljlJdOAdJojsJfkjg/pqK47Q8WmOLE19tGWQi"), // 123456
				},
				"john.doe2.principal.id": password.Principal{
					ID:             "john.doe2.principal.id",
					UserID:         "john.doe2.id",
					LoginIDKey:     "email",
					LoginID:        "john.doe@example.com",
					HashedPassword: []byte("$2a$10$/jm/S1sY6ldfL6UZljlJdOAdJojsJfkjg/pqK47Q8WmOLE19tGWQi"), // 123456
				},
				"chima.principal.id": password.Principal{
					ID:             "chima.principal.id",
					UserID:         "chima.id",
					LoginIDKey:     "email",
					LoginID:        "chima@example.com",
					HashedPassword: []byte("$2a$10$/jm/S1sY6ldfL6UZljlJdOAdJojsJfkjg/pqK47Q8WmOLE19tGWQi"), // 123456
				},
			},
		)
		fh.AuthInfoStore = authinfo.NewMockStoreWithAuthInfoMap(
			map[string]authinfo.AuthInfo{
				"john.doe.id": authinfo.AuthInfo{
					ID: "john.doe.id",
				},
				"john.doe2.id": authinfo.AuthInfo{
					ID: "john.doe2.id",
				},
				"chima.id": authinfo.AuthInfo{
					ID: "chima.id",
				},
			},
		)
		userProfileStore := userprofile.NewMockUserProfileStore()
		userProfileStore.Data["john.doe.id"] = map[string]interface{}{
			"username": "john.doe",
			"email":    "john.doe@example.com",
		}
		userProfileStore.Data["john.doe2.id"] = map[string]interface{}{
			"username": "john.doe2",
			"email":    "john.doe@example.com",
		}
		userProfileStore.Data["chima.id"] = map[string]interface{}{
			"username": "chima",
			"email":    "chima@example.com",
		}
		fh.UserProfileStore = userProfileStore
		sender := getSender()
		fh.ForgotPasswordEmailSender = sender

		Convey("send email to user", func() {
			req, _ := http.NewRequest("POST", "", strings.NewReader(`{
				"email": "chima@example.com"
			}`))
			resp := httptest.NewRecorder()
			h := handler.APIHandlerToHandler(fh, fh.TxContext)
			h.ServeHTTP(resp, req)
			So(resp.Body.Bytes(), ShouldEqualJSON, `{
				"result": "OK"
			}`)
			So(sender.emails, ShouldResemble, []string{"chima@example.com"})
			So(sender.userProfiles, ShouldHaveLength, 1)
			So(sender.userProfileIDs, ShouldContain, "chima.id")
			So(sender.authInfos, ShouldHaveLength, 1)
			So(sender.authInfoIDs, ShouldContain, "chima.id")
			So(sender.hashedPasswords, ShouldResemble, [][]byte{
				[]byte("$2a$10$/jm/S1sY6ldfL6UZljlJdOAdJojsJfkjg/pqK47Q8WmOLE19tGWQi"),
			})
		})

		Convey("send email to users with the same email", func() {
			req, _ := http.NewRequest("POST", "", strings.NewReader(`{
				"email": "john.doe@example.com"
			}`))
			resp := httptest.NewRecorder()
			h := handler.APIHandlerToHandler(fh, fh.TxContext)
			h.ServeHTTP(resp, req)
			So(resp.Body.Bytes(), ShouldEqualJSON, `{
				"result": "OK"
			}`)
			So(sender.emails, ShouldResemble, []string{"john.doe@example.com", "john.doe@example.com"})
			So(sender.userProfiles, ShouldHaveLength, 2)
			So(sender.userProfileIDs, ShouldContain, "john.doe.id")
			So(sender.userProfileIDs, ShouldContain, "john.doe2.id")
			So(sender.authInfos, ShouldHaveLength, 2)
			So(sender.authInfoIDs, ShouldContain, "john.doe.id")
			So(sender.authInfoIDs, ShouldContain, "john.doe2.id")
			So(sender.hashedPasswords, ShouldResemble, [][]byte{
				[]byte("$2a$10$/jm/S1sY6ldfL6UZljlJdOAdJojsJfkjg/pqK47Q8WmOLE19tGWQi"),
				[]byte("$2a$10$/jm/S1sY6ldfL6UZljlJdOAdJojsJfkjg/pqK47Q8WmOLE19tGWQi"),
			})
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
	emails          []string
	authInfos       []authinfo.AuthInfo
	authInfoIDs     []string
	userProfiles    []userprofile.UserProfile
	userProfileIDs  []string
	hashedPasswords [][]byte
}

func (m *MockForgotPasswordEmailSender) Send(
	email string,
	authInfo authinfo.AuthInfo,
	userProfile userprofile.UserProfile,
	hashedPassword []byte,
) (err error) {
	m.emails = append(m.emails, email)
	m.authInfos = append(m.authInfos, authInfo)
	m.authInfoIDs = append(m.authInfoIDs, authInfo.ID)
	m.userProfiles = append(m.userProfiles, userProfile)
	m.userProfileIDs = append(m.userProfileIDs, userProfile.ID)
	m.hashedPasswords = append(m.hashedPasswords, hashedPassword)
	return nil
}
