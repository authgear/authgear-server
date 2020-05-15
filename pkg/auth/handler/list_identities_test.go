package handler

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	. "github.com/smartystreets/goconvey/convey"

	authtesting "github.com/skygeario/skygear-server/pkg/auth/dependency/auth/testing"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/identity"
	"github.com/skygeario/skygear-server/pkg/core/authn"
	"github.com/skygeario/skygear-server/pkg/core/db"
	. "github.com/skygeario/skygear-server/pkg/core/skytest"
)

type MockListIdentityProvider struct{}

func (m *MockListIdentityProvider) ListByUser(userID string) ([]*identity.Info, error) {
	return []*identity.Info{
		&identity.Info{
			Type: authn.IdentityTypeLoginID,
			Claims: map[string]interface{}{
				"email":                            "user1@example.com",
				identity.IdentityClaimLoginIDKey:   "email",
				identity.IdentityClaimLoginIDValue: "user1@example.com",
			},
		},
		&identity.Info{
			Type: authn.IdentityTypeLoginID,
			Claims: map[string]interface{}{
				"username":                         "user1",
				identity.IdentityClaimLoginIDKey:   "username",
				identity.IdentityClaimLoginIDValue: "user1",
			},
		},
		&identity.Info{
			Type: authn.IdentityTypeOAuth,
			Claims: map[string]interface{}{
				"email":                              "user1@example.com",
				identity.IdentityClaimOAuthProvider:  map[string]interface{}{"type": "google"},
				identity.IdentityClaimOAuthSubjectID: "google-user-id",
				identity.IdentityClaimOAuthProfile: map[string]interface{}{
					"email": "user1@example.com",
					"name":  "User 1",
				},
			},
		},
	}, nil
}

func TestListIdentitiesHandler(t *testing.T) {
	Convey("Test ListIdentitiesHandler", t, func() {
		h := &ListIdentitiesHandler{}
		h.TxContext = db.NewMockTxContext()
		h.IdentityProvider = &MockListIdentityProvider{}

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
							"type": "login_id",
							"claims": {
								"email": "user1@example.com",
								"https://auth.skygear.io/claims/login_id/key":   "email",
								"https://auth.skygear.io/claims/login_id/value": "user1@example.com"
							}
						},
						{
							"type": "login_id",
							"claims": {
								"username": "user1",
								"https://auth.skygear.io/claims/login_id/key":   "username",
								"https://auth.skygear.io/claims/login_id/value": "user1"
							}
						},
						{
							"type": "oauth",
							"claims": {
								"email": "user1@example.com",
								"https://auth.skygear.io/claims/oauth/provider": {"type": "google"},
								"https://auth.skygear.io/claims/oauth/subject_id": "google-user-id",
								"https://auth.skygear.io/claims/oauth/profile": {
									"email": "user1@example.com",
									"name": "User 1"
								}
							}
						}
					]
				}
			}`)
		})
	})
}
