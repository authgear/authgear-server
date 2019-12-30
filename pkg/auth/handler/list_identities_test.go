package handler

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	. "github.com/skygeario/skygear-server/pkg/core/skytest"
	. "github.com/smartystreets/goconvey/convey"

	"github.com/skygeario/skygear-server/pkg/auth/dependency/principal"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/principal/oauth"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/principal/password"
	authtest "github.com/skygeario/skygear-server/pkg/core/auth/testing"
	"github.com/skygeario/skygear-server/pkg/core/config"
	"github.com/skygeario/skygear-server/pkg/core/db"
)

func TestListIdentitiesHandler(t *testing.T) {
	Convey("Test ListIdentitiesHandler", t, func() {
		h := &ListIdentitiesHandler{}
		h.TxContext = db.NewMockTxContext()
		authContext := authtest.NewMockContext().
			UseUser("user-id-1", "principal-id-1")
		h.AuthContext = authContext
		passwordAuthProvider := password.NewMockProviderWithPrincipalMap(
			[]config.LoginIDKeyConfiguration{},
			[]string{password.DefaultRealm},
			map[string]password.Principal{
				"principal-id-1": password.Principal{
					ID:         "principal-id-1",
					UserID:     "user-id-1",
					LoginIDKey: "email",
					LoginID:    "user1@example.com",
					Realm:      password.DefaultRealm,
					ClaimsValue: map[string]interface{}{
						"email": "user1@example.com",
					},
				},
				"principal-id-2": password.Principal{
					ID:         "principal-id-2",
					UserID:     "user-id-1",
					LoginIDKey: "username",
					LoginID:    "user1",
					Realm:      password.DefaultRealm,
					ClaimsValue: map[string]interface{}{
						"username": "user1",
					},
				},
			},
		)
		oauthProvider := oauth.NewMockProvider([]*oauth.Principal{
			&oauth.Principal{
				ID:             "principal-id-3",
				UserID:         "user-id-1",
				ProviderType:   "google",
				ProviderKeys:   map[string]interface{}{},
				ProviderUserID: "google-user-id",
				UserProfile: map[string]interface{}{
					"name":  "User 1",
					"email": "user1@example.com",
				},
				ClaimsValue: map[string]interface{}{
					"email": "user1@example.com",
				},
			},
		})
		h.IdentityProvider = principal.NewMockIdentityProvider(passwordAuthProvider, oauthProvider)

		Convey("should return list of identities", func() {
			r, _ := http.NewRequest("POST", "", strings.NewReader("{}"))
			r.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()
			h.ServeHTTP(w, r)

			So(w.Body.Bytes(), ShouldEqualJSON, `{
				"result": {
					"identities": [
						{
							"id": "principal-id-1",
							"type": "password",
							"login_id_key": "email",
							"login_id": "user1@example.com",
							"claims": {
								"email": "user1@example.com"
							}
						},
						{
							"id": "principal-id-2",
							"type": "password",
							"login_id_key": "username",
							"login_id": "user1",
							"claims": {
								"username": "user1"
							}
						},
						{
							"id": "principal-id-3",
							"type": "oauth",
							"provider_type": "google",
							"provider_user_id": "google-user-id",
							"provider_keys": {},
							"raw_profile": {
								"name": "User 1",
								"email": "user1@example.com"
							},
							"claims": {
								"email": "user1@example.com"
							}
						}
					]
				}
			}`)
		})
	})
}
