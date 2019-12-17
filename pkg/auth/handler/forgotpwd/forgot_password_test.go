package forgotpwd

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	. "github.com/smartystreets/goconvey/convey"

	"github.com/skygeario/skygear-server/pkg/auth/dependency/principal"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/principal/password"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/userprofile"
	"github.com/skygeario/skygear-server/pkg/auth/model"
	"github.com/skygeario/skygear-server/pkg/core/auth/authinfo"
	"github.com/skygeario/skygear-server/pkg/core/auth/metadata"
	"github.com/skygeario/skygear-server/pkg/core/config"
	"github.com/skygeario/skygear-server/pkg/core/db"
	. "github.com/skygeario/skygear-server/pkg/core/skytest"
	"github.com/skygeario/skygear-server/pkg/core/validation"
)

func TestForgotPasswordHandler(t *testing.T) {
	getSender := func() *MockForgotPasswordEmailSender {
		return &MockForgotPasswordEmailSender{}
	}

	Convey("Test ForgotPasswordHandler", t, func() {
		fh := &ForgotPasswordHandler{}
		validator := validation.NewValidator("http://v2.skygear.io")
		validator.AddSchemaFragments(
			ForgotPasswordRequestSchema,
		)
		fh.Validator = validator
		fh.TxContext = db.NewMockTxContext()
		fh.PasswordAuthProvider = password.NewMockProviderWithPrincipalMap(
			[]config.LoginIDKeyConfiguration{
				config.LoginIDKeyConfiguration{Key: "email", Type: config.LoginIDKeyType(metadata.Email)},
				config.LoginIDKeyConfiguration{Key: "phone", Type: config.LoginIDKeyType(metadata.Phone)},
			},
			[]string{password.DefaultRealm},
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
				"chima.principal1.id": password.Principal{
					ID:             "chima.principal1.id",
					UserID:         "chima.id",
					LoginIDKey:     "email",
					LoginID:        "chima@example.com",
					HashedPassword: []byte("$2a$10$/jm/S1sY6ldfL6UZljlJdOAdJojsJfkjg/pqK47Q8WmOLE19tGWQi"), // 123456
				},
				"chima.principal2.id": password.Principal{
					ID:             "chima.principal2.id",
					UserID:         "chima.id",
					LoginIDKey:     "phone",
					LoginID:        "+85299999999",
					HashedPassword: []byte("$2a$10$/jm/S1sY6ldfL6UZljlJdOAdJojsJfkjg/pqK47Q8WmOLE19tGWQi"), // 123456
				},
			},
		)
		fh.IdentityProvider = principal.NewMockIdentityProvider(fh.PasswordAuthProvider)
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
			req.Header.Set("Content-Type", "application/json")
			resp := httptest.NewRecorder()
			fh.ServeHTTP(resp, req)
			So(resp.Body.Bytes(), ShouldEqualJSON, `{
				"result": {}
			}`)
			So(sender.emails, ShouldResemble, []string{"chima@example.com"})
			So(sender.userObjects, ShouldHaveLength, 1)
			So(sender.userObjectIDs, ShouldContain, "chima.id")
			So(sender.authInfos, ShouldHaveLength, 1)
			So(sender.authInfoIDs, ShouldContain, "chima.id")
			So(sender.hashedPasswords, ShouldResemble, [][]byte{
				[]byte("$2a$10$/jm/S1sY6ldfL6UZljlJdOAdJojsJfkjg/pqK47Q8WmOLE19tGWQi"),
			})
		})

		Convey("should send email to correct user email", func() {
			req, _ := http.NewRequest("POST", "", strings.NewReader(`{
				"email": "chIma@example.com"
			}`))
			req.Header.Set("Content-Type", "application/json")
			resp := httptest.NewRecorder()
			fh.ServeHTTP(resp, req)
			So(resp.Body.Bytes(), ShouldEqualJSON, `{
				"result": {}
			}`)
			So(sender.emails, ShouldResemble, []string{"chima@example.com"})
			So(sender.userObjects, ShouldHaveLength, 1)
			So(sender.userObjectIDs, ShouldContain, "chima.id")
		})

		Convey("send email to users with the same email", func() {
			req, _ := http.NewRequest("POST", "", strings.NewReader(`{
				"email": "john.doe@example.com"
			}`))
			req.Header.Set("Content-Type", "application/json")
			resp := httptest.NewRecorder()
			fh.ServeHTTP(resp, req)
			So(resp.Body.Bytes(), ShouldEqualJSON, `{
				"result": {}
			}`)
			So(sender.emails, ShouldResemble, []string{"john.doe@example.com", "john.doe@example.com"})
			So(sender.userObjects, ShouldHaveLength, 2)
			So(sender.userObjectIDs, ShouldContain, "john.doe.id")
			So(sender.userObjectIDs, ShouldContain, "john.doe2.id")
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
			req.Header.Set("Content-Type", "application/json")
			resp := httptest.NewRecorder()
			fh.ServeHTTP(resp, req)
			So(resp.Body.Bytes(), ShouldEqualJSON, `{
				"error": {
					"name": "Invalid",
					"reason": "ValidationFailed",
					"message": "invalid request body",
					"code": 400,
					"info": {
						"causes": [
							{
								"kind": "StringFormat",
								"pointer": "/email",
								"message": "Does not match format 'email'",
								"details": { "format": "email" }
							}
						]
					}
				}
			}`)
		})

		Convey("throw error for unknown email", func() {
			req, _ := http.NewRequest("POST", "", strings.NewReader(`{
				"email": "iamyourfather@example.com"
			}`))
			req.Header.Set("Content-Type", "application/json")
			resp := httptest.NewRecorder()
			fh.ServeHTTP(resp, req)
			So(resp.Body.Bytes(), ShouldEqualJSON, `{
				"error": {
					"name": "NotFound",
					"reason": "UserNotFound",
					"message": "user not found",
					"code": 404
				}
			}`)
		})
	})
}

type MockForgotPasswordEmailSender struct {
	emails          []string
	authInfos       []authinfo.AuthInfo
	authInfoIDs     []string
	userObjects     []model.User
	userObjectIDs   []string
	hashedPasswords [][]byte
}

func (m *MockForgotPasswordEmailSender) Send(
	email string,
	authInfo authinfo.AuthInfo,
	user model.User,
	hashedPassword []byte,
) (err error) {
	m.emails = append(m.emails, email)
	m.authInfos = append(m.authInfos, authInfo)
	m.authInfoIDs = append(m.authInfoIDs, authInfo.ID)
	m.userObjects = append(m.userObjects, user)
	m.userObjectIDs = append(m.userObjectIDs, user.ID)
	m.hashedPasswords = append(m.hashedPasswords, hashedPassword)
	return nil
}
