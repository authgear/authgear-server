package handler

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	. "github.com/smartystreets/goconvey/convey"

	authtesting "github.com/skygeario/skygear-server/pkg/auth/dependency/auth/testing"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/principal"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/principal/oauth"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/principal/password"
	"github.com/skygeario/skygear-server/pkg/core/config"
	"github.com/skygeario/skygear-server/pkg/core/db"
	. "github.com/skygeario/skygear-server/pkg/core/skytest"
)

func TestListIdentitiesHandler(t *testing.T) {
	Convey("Test ListIdentitiesHandler", t, func() {
		h := &ListIdentitiesHandler{}
		h.TxContext = db.NewMockTxContext()
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
			r = authtesting.WithAuthn().
				UserID("user-id-1").
				ToRequest(r)
			r.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()
			h.ServeHTTP(w, r)

			So(w.Body.Bytes(), ShouldEqualJSON, `{
				"result": {
					"identities": [
						{
							"type": "password",
							"claims": {
								"email": "user1@example.com"
							}
						},
						{
							"type": "password",
							"claims": {
								"username": "user1"
							}
						},
						{
							"type": "oauth",
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
