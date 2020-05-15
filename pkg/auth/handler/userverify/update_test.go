package userverify

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/skygeario/skygear-server/pkg/auth/dependency/hook"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/identity/loginid"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/userprofile"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/userverify"
	"github.com/skygeario/skygear-server/pkg/auth/event"
	"github.com/skygeario/skygear-server/pkg/auth/model"
	"github.com/skygeario/skygear-server/pkg/core/auth/authinfo"
	"github.com/skygeario/skygear-server/pkg/core/config"
	"github.com/skygeario/skygear-server/pkg/core/db"
	. "github.com/skygeario/skygear-server/pkg/core/skytest"
	"github.com/skygeario/skygear-server/pkg/core/time"
	"github.com/skygeario/skygear-server/pkg/core/validation"
	. "github.com/smartystreets/goconvey/convey"
)

func TestUpdateHandler(t *testing.T) {
	Convey("TestUpdateHandler", t, func() {
		vh := &UpdateHandler{}
		validator := validation.NewValidator("http://v2.skygear.io")
		validator.AddSchemaFragments(
			UpdateVerifyStateRequestSchema,
		)
		vh.Validator = validator
		vh.TxContext = db.NewMockTxContext()

		vh.LoginIDProvider = &mockLoginIDProvider{
			Identities: []loginid.Identity{
				{
					ID:         "principal-id-1",
					UserID:     "user-id-1",
					LoginIDKey: "email",
					LoginID:    "user+1@example.com",
				},
				{
					ID:         "principal-id-2",
					UserID:     "user-id-1",
					LoginIDKey: "email",
					LoginID:    "user+2@example.com",
				},
			},
		}

		authInfo := authinfo.AuthInfo{
			ID:               "user-id-1",
			Verified:         false,
			ManuallyVerified: false,
			VerifyInfo: map[string]bool{
				"user+1@example.com": true,
			},
		}
		authInfoStore := authinfo.NewMockStoreWithAuthInfoMap(
			map[string]authinfo.AuthInfo{
				"user-id-1": authInfo,
			},
		)
		vh.AuthInfoStore = authInfoStore
		vh.UserProfileStore = userprofile.NewMockUserProfileStore()
		hookProvider := hook.NewMockProvider()
		vh.HookProvider = hookProvider

		verifyConfig := &config.UserVerificationConfiguration{
			Criteria: config.UserVerificationCriteriaAll,
			LoginIDKeys: []config.UserVerificationKeyConfiguration{
				config.UserVerificationKeyConfiguration{
					Key:    "email",
					Expiry: 12 * 60 * 60,
				},
			},
		}
		store := userverify.MockStore{}
		vh.UserVerificationProvider = userverify.NewProvider(nil, &store, verifyConfig, &time.MockProvider{})

		Convey("should require user ID", func() {
			req, _ := http.NewRequest("POST", "", strings.NewReader(`{}`))
			req.Header.Set("Content-Type", "application/json")
			resp := httptest.NewRecorder()
			vh.ServeHTTP(resp, req)
			So(resp.Body.Bytes(), ShouldEqualJSON, `{
				"error":{
					"name": "Invalid",
					"reason": "ValidationFailed",
					"message": "invalid request body",
					"code": 400,
					"info": {
						"causes": [
							{ "kind": "Required", "pointer": "/user_id", "message": "user_id is required" }
						]
					}
				}
			}`)
		})

		Convey("should set verify info", func() {
			req, _ := http.NewRequest("POST", "", strings.NewReader(`{
				"user_id": "user-id-1",
				"verify_info": {
					"user+1@example.com": true,
					"user+2@example.com": true
				}
			}`))
			req.Header.Set("Content-Type", "application/json")
			resp := httptest.NewRecorder()
			vh.ServeHTTP(resp, req)
			So(resp.Body.Bytes(), ShouldEqualJSON, `{
				"result": {
					"user": {
						"id": "user-id-1",
						"created_at": "0001-01-01T00:00:00Z",
						"is_disabled": false,
						"is_manually_verified": false,
						"is_verified": true,
						"verify_info": {
							"user+1@example.com": true,
							"user+2@example.com": true
						},
						"metadata": {}
					}
				}
			}`)

			isVerified := true
			So(hookProvider.DispatchedEvents, ShouldResemble, []event.Payload{
				event.UserUpdateEvent{
					Reason:     event.UserUpdateReasonAdministrative,
					IsVerified: &isVerified,
					VerifyInfo: &map[string]bool{
						"user+1@example.com": true,
						"user+2@example.com": true,
					},
					User: model.User{
						ID:         "user-id-1",
						Verified:   false,
						Disabled:   false,
						VerifyInfo: map[string]bool{"user+1@example.com": true},
						Metadata:   userprofile.Data{},
					},
				},
			})
		})

		Convey("should ignore false verify entry", func() {
			req, _ := http.NewRequest("POST", "", strings.NewReader(`{
				"user_id": "user-id-1",
				"verify_info": {
					"user+1@example.com": false,
					"user+2@example.com": true
				}
			}`))
			req.Header.Set("Content-Type", "application/json")
			resp := httptest.NewRecorder()
			vh.ServeHTTP(resp, req)
			So(resp.Body.Bytes(), ShouldEqualJSON, `{
				"result": {
					"user": {
						"id": "user-id-1",
						"created_at": "0001-01-01T00:00:00Z",
						"is_disabled": false,
						"is_manually_verified": false,
						"is_verified": false,
						"verify_info": {
							"user+2@example.com": true
						},
						"metadata": {}
					}
				}
			}`)
		})

		Convey("should set manually verified", func() {
			req, _ := http.NewRequest("POST", "", strings.NewReader(`{
				"user_id": "user-id-1",
				"is_manually_verified": true
			}`))
			req.Header.Set("Content-Type", "application/json")
			resp := httptest.NewRecorder()
			vh.ServeHTTP(resp, req)
			So(resp.Body.Bytes(), ShouldEqualJSON, `{
				"result": {
					"user": {
						"id": "user-id-1",
						"created_at": "0001-01-01T00:00:00Z",
						"is_disabled": false,
						"is_manually_verified": true,
						"is_verified": true,
						"verify_info": {
							"user+1@example.com": true
						},
						"metadata": {}
					}
				}
			}`)
		})

		Convey("should set verify info and manually verified", func() {
			req, _ := http.NewRequest("POST", "", strings.NewReader(`{
				"user_id": "user-id-1",
				"verify_info": {},
				"is_manually_verified": true
			}`))
			req.Header.Set("Content-Type", "application/json")
			resp := httptest.NewRecorder()
			vh.ServeHTTP(resp, req)
			So(resp.Body.Bytes(), ShouldEqualJSON, `{
				"result": {
					"user": {
						"id": "user-id-1",
						"created_at": "0001-01-01T00:00:00Z",
						"is_disabled": false,
						"is_manually_verified": true,
						"is_verified": true,
						"verify_info": {},
						"metadata": {}
					}
				}
			}`)
		})
	})
}
