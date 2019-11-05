package handler

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/skygeario/skygear-server/pkg/auth/dependency/hook"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/principal"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/principal/password"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/userprofile"
	"github.com/skygeario/skygear-server/pkg/auth/event"
	"github.com/skygeario/skygear-server/pkg/auth/model"
	"github.com/skygeario/skygear-server/pkg/core/auth/authinfo"
	authtest "github.com/skygeario/skygear-server/pkg/core/auth/testing"
	"github.com/skygeario/skygear-server/pkg/core/config"
	"github.com/skygeario/skygear-server/pkg/core/db"
	. "github.com/skygeario/skygear-server/pkg/core/skytest"
	"github.com/skygeario/skygear-server/pkg/core/validation"
	. "github.com/smartystreets/goconvey/convey"
)

func TestUpdateMetadataHandler(t *testing.T) {
	var zeroTime time.Time
	realTime := timeNow
	timeNow = func() time.Time { return zeroTime }
	defer func() {
		timeNow = realTime
	}()

	Convey("Test UpdateMetadataHandler", t, func() {
		// fixture
		userID := "john.doe.id"

		uh := &UpdateMetadataHandler{}
		validator := validation.NewValidator("http://v2.skygear.io")
		validator.AddSchemaFragments(
			UpdateMetadataRequestSchema,
		)
		uh.Validator = validator
		uh.AuthContext = authtest.NewMockContext().
			UseUser(userID, "john.doe.principal.id0").
			MarkVerified()
		uh.AuthInfoStore = authinfo.NewMockStoreWithAuthInfoMap(
			map[string]authinfo.AuthInfo{
				userID: authinfo.AuthInfo{
					ID:         userID,
					Verified:   true,
					VerifyInfo: map[string]bool{},
				},
			},
		)
		profileData := map[string]map[string]interface{}{
			"john.doe.id": map[string]interface{}{
				"username": "john.doe",
				"email":    "john.doe@example.com",
			},
		}
		uh.UserProfileStore = userprofile.NewMockUserProfileStoreByData(profileData)

		loginIDsKeys := []config.LoginIDKeyConfiguration{}
		allowedRealms := []string{password.DefaultRealm}
		passwordAuthProvider := password.NewMockProviderWithPrincipalMap(
			loginIDsKeys,
			allowedRealms,
			map[string]password.Principal{
				"john.doe.principal.id0": password.Principal{
					ID:             "john.doe.principal.id0",
					UserID:         "john.doe.id",
					LoginIDKey:     "username",
					LoginID:        "john.doe",
					Realm:          "default",
					HashedPassword: []byte("$2a$10$/jm/S1sY6ldfL6UZljlJdOAdJojsJfkjg/pqK47Q8WmOLE19tGWQi"), // 123456
				},
				"john.doe.principal.id1": password.Principal{
					ID:             "john.doe.principal.id1",
					UserID:         "john.doe.id",
					LoginIDKey:     "email",
					LoginID:        "john.doe@example.com",
					Realm:          "default",
					HashedPassword: []byte("$2a$10$/jm/S1sY6ldfL6UZljlJdOAdJojsJfkjg/pqK47Q8WmOLE19tGWQi"), // 123456
				},
			},
		)
		uh.PasswordAuthProvider = passwordAuthProvider
		uh.IdentityProvider = principal.NewMockIdentityProvider(uh.PasswordAuthProvider)
		hookProvider := hook.NewMockProvider()
		uh.HookProvider = hookProvider
		uh.TxContext = db.NewMockTxContext()

		Convey("should update metadata", func() {
			req, _ := http.NewRequest("POST", "", strings.NewReader(`
			{
				"metadata": {
					"username": "john.doe",
					"email":    "john.doe@example.com",
					"age": 24
				}
			}`))
			req.Header.Set("Content-Type", "application/json")
			resp := httptest.NewRecorder()
			uh.ServeHTTP(resp, req)

			So(resp.Code, ShouldEqual, 200)
			So(resp.Body.Bytes(), ShouldEqualJSON, fmt.Sprintf(`{
				"result": {
					"user": {
						"id": "%s",
						"is_verified": true,
						"is_disabled": false,
						"created_at": "0001-01-01T00:00:00Z",
						"verify_info": {},
						"metadata": {
							"username": "john.doe",
							"email": "john.doe@example.com",
							"age": 24
						}
					}
				}
			}`, userID))

			So(hookProvider.DispatchedEvents, ShouldResemble, []event.Payload{
				event.UserUpdateEvent{
					Reason: event.UserUpdateReasonUpdateMetadata,
					Metadata: &userprofile.Data{
						"username": "john.doe",
						"email":    "john.doe@example.com",
						"age":      float64(24),
					},
					User: model.User{
						ID:         userID,
						Verified:   true,
						Disabled:   false,
						VerifyInfo: map[string]bool{},
						Metadata: userprofile.Data{
							"username": "john.doe",
							"email":    "john.doe@example.com",
						},
					},
				},
			})
		})

		Convey("should allow to delete attributes in metadata", func() {
			req, _ := http.NewRequest("POST", "", strings.NewReader(`
			{
				"metadata": {
					"username": "john.doe",
					"email":    "john.doe@example.com",
					"age":      30,
					"love":     "cat"
				}
			}`))
			req.Header.Set("Content-Type", "application/json")
			resp := httptest.NewRecorder()
			uh.ServeHTTP(resp, req)

			So(resp.Code, ShouldEqual, 200)
			So(resp.Body.Bytes(), ShouldEqualJSON, fmt.Sprintf(`{
				"result": {
					"user": {
						"id": "%s",
						"is_verified": true,
						"is_disabled": false,
						"created_at": "0001-01-01T00:00:00Z",
						"verify_info": {},
						"metadata": {
							"username": "john.doe",
							"email": "john.doe@example.com",
							"age": 30,
							"love": "cat"
						}
					}
				}
			}`, userID))

			req, _ = http.NewRequest("POST", "", strings.NewReader(`
			{
				"metadata": {
					"username": "john.doe",
					"email":    "john.doe@example.com"
				}
			}`))
			req.Header.Set("Content-Type", "application/json")
			resp = httptest.NewRecorder()
			uh.ServeHTTP(resp, req)

			So(resp.Code, ShouldEqual, 200)
			So(resp.Body.Bytes(), ShouldEqualJSON, fmt.Sprintf(`{
				"result": {
					"user": {
						"id": "%s",
						"is_verified": true,
						"is_disabled": false,
						"created_at": "0001-01-01T00:00:00Z",
						"verify_info": {},
						"metadata": {
							"username": "john.doe",
							"email": "john.doe@example.com"
						}
					}
				}
			}`, userID))
		})

		Convey("shouldn't update another user's metadata", func() {
			req, _ := http.NewRequest("POST", "", strings.NewReader(`
			{
				"user_id": "mary.jane",
				"metadata": {
					"username": "mary.jane",
					"email":    "mary.jane@example.com",
					"age": 25
				}
			}`))
			req.Header.Set("Content-Type", "application/json")
			resp := httptest.NewRecorder()
			uh.ServeHTTP(resp, req)

			So(resp.Code, ShouldEqual, 403)
			So(resp.Body.Bytes(), ShouldEqualJSON, `
			{
				"error": {
					"name": "Forbidden",
					"reason": "Forbidden",
					"message": "must not specify user_id",
					"code": 403
				}
			}`)
		})
	})

	Convey("Test UpdateMetadataHandler by MasterKey", t, func() {
		// fixture
		userID := "john.doe.id"

		uh := &UpdateMetadataHandler{}
		validator := validation.NewValidator("http://v2.skygear.io")
		validator.AddSchemaFragments(
			UpdateMetadataRequestSchema,
		)
		uh.Validator = validator
		uh.AuthContext = authtest.NewMockContext().
			UseUser("faseng.cat.id", "faseng.cat.principal.id").
			MarkVerified().
			UseMasterKey()
		uh.AuthInfoStore = authinfo.NewMockStoreWithAuthInfoMap(
			map[string]authinfo.AuthInfo{
				userID: authinfo.AuthInfo{
					ID:         userID,
					Verified:   true,
					VerifyInfo: map[string]bool{},
				},
			},
		)
		profileData := map[string]map[string]interface{}{
			"john.doe.id": map[string]interface{}{
				"username": "john.doe",
				"email":    "john.doe@example.com",
			},
		}
		uh.UserProfileStore = userprofile.NewMockUserProfileStoreByData(profileData)

		loginIDsKeys := []config.LoginIDKeyConfiguration{}
		allowedRealms := []string{password.DefaultRealm}
		passwordAuthProvider := password.NewMockProviderWithPrincipalMap(
			loginIDsKeys,
			allowedRealms,
			map[string]password.Principal{
				"john.doe.principal.id0": password.Principal{
					ID:             "john.doe.principal.id0",
					UserID:         "john.doe.id",
					LoginIDKey:     "username",
					LoginID:        "john.doe",
					HashedPassword: []byte("$2a$10$/jm/S1sY6ldfL6UZljlJdOAdJojsJfkjg/pqK47Q8WmOLE19tGWQi"), // 123456
				},
				"john.doe.principal.id1": password.Principal{
					ID:             "john.doe.principal.id1",
					UserID:         "john.doe.id",
					LoginIDKey:     "email",
					LoginID:        "john.doe@example.com",
					HashedPassword: []byte("$2a$10$/jm/S1sY6ldfL6UZljlJdOAdJojsJfkjg/pqK47Q8WmOLE19tGWQi"), // 123456
				},
			},
		)
		uh.PasswordAuthProvider = passwordAuthProvider
		uh.TxContext = db.NewMockTxContext()
		uh.HookProvider = hook.NewMockProvider()

		Convey("able to update another user's metadata", func() {
			req, _ := http.NewRequest("POST", "", strings.NewReader(fmt.Sprintf(`
			{
				"user_id": "%s",
				"metadata": {
					"username": "john.doe",
					"email": "john.doe@example.com",
					"age": 25
				}
			}`,
				userID)))
			req.Header.Set("Content-Type", "application/json")
			resp := httptest.NewRecorder()
			uh.ServeHTTP(resp, req)

			So(resp.Code, ShouldEqual, 200)
			So(resp.Body.Bytes(), ShouldEqualJSON, fmt.Sprintf(`{
				"result": {
					"user": {
						"id": "%s",
						"is_verified": true,
						"is_disabled": false,
						"created_at": "0001-01-01T00:00:00Z",
						"verify_info": {},
						"metadata": {
							"username": "john.doe",
							"email": "john.doe@example.com",
							"age": 25
						}
					}
				}
			}`, userID))
		})
	})
}
