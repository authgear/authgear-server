package authn

import (
	"testing"
	"time"

	. "github.com/smartystreets/goconvey/convey"

	"github.com/skygeario/skygear-server/pkg/auth/dependency/loginid"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/principal"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/principal/oauth"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/principal/password"
	"github.com/skygeario/skygear-server/pkg/core/auth/metadata"
	"github.com/skygeario/skygear-server/pkg/core/config"
	"github.com/skygeario/skygear-server/pkg/core/handler"
	. "github.com/skygeario/skygear-server/pkg/core/skytest"
	coreTime "github.com/skygeario/skygear-server/pkg/core/time"
)

func TestAuthenticateWithLoginID(t *testing.T) {
	Convey("AuthenticateProcess.AuthenticateWithLoginID", t, func() {
		impl := &AuthenticateProcess{}
		one := 1
		loginIDsKeys := []config.LoginIDKeyConfiguration{
			config.LoginIDKeyConfiguration{
				Key:     "email",
				Type:    config.LoginIDKeyType(metadata.Email),
				Maximum: &one,
			},
			config.LoginIDKeyConfiguration{
				Key:     "username",
				Type:    config.LoginIDKeyTypeRaw,
				Maximum: &one,
			},
		}
		allowedRealms := []string{password.DefaultRealm}
		passwordProvider := password.NewMockProviderWithPrincipalMap(
			loginIDsKeys,
			allowedRealms,
			map[string]password.Principal{
				"john.doe.principal.id1": password.Principal{
					ID:             "john.doe.principal.id1",
					UserID:         "john.doe.id",
					LoginIDKey:     "email",
					LoginID:        "john.doe@example.com",
					Realm:          password.DefaultRealm,
					HashedPassword: []byte("$2a$10$/jm/S1sY6ldfL6UZljlJdOAdJojsJfkjg/pqK47Q8WmOLE19tGWQi"), // 123456
					ClaimsValue: map[string]interface{}{
						"email": "john.doe@example.com",
					},
				},
				"john.doe.principal.id2": password.Principal{
					ID:             "john.doe.principal.id2",
					UserID:         "john.doe.id",
					LoginIDKey:     "username",
					LoginID:        "john.doe",
					Realm:          password.DefaultRealm,
					HashedPassword: []byte("$2a$10$/jm/S1sY6ldfL6UZljlJdOAdJojsJfkjg/pqK47Q8WmOLE19tGWQi"), // 123456
					ClaimsValue:    map[string]interface{}{},
				},
				"john.doe.principal.id3": password.Principal{
					ID:             "john.doe.principal.id3",
					UserID:         "john.doe.id",
					LoginIDKey:     "email",
					LoginID:        "john.doe+1@example.com",
					Realm:          "admin",
					HashedPassword: []byte("$2a$10$/jm/S1sY6ldfL6UZljlJdOAdJojsJfkjg/pqK47Q8WmOLE19tGWQi"), // 123456
					ClaimsValue: map[string]interface{}{
						"email": "john.doe+1@example.com",
					},
				},
			},
		)
		oauthProvider := oauth.NewMockProvider(nil)

		timeProvider := &coreTime.MockProvider{TimeNowUTC: time.Date(2006, 1, 2, 15, 4, 5, 0, time.UTC)}
		identityProvider := principal.NewMockIdentityProvider(passwordProvider)

		impl.OAuthProvider = oauthProvider
		impl.IdentityProvider = identityProvider
		impl.TimeProvider = timeProvider
		impl.PasswordProvider = passwordProvider

		checkErr := func(err error, errJSON string) {
			So(err, ShouldNotBeNil)
			b, _ := handler.APIResponse{Error: err}.MarshalJSON()
			So(b, ShouldEqualJSON, errJSON)
		}

		Convey("without login ID key", func() {
			_, err := impl.AuthenticateWithLoginID(loginid.LoginID{
				Value: "john.doe@example.com",
			}, "123456")
			So(err, ShouldBeNil)
		})

		Convey("with correct login ID key", func() {
			_, err := impl.AuthenticateWithLoginID(loginid.LoginID{
				Key:   "email",
				Value: "john.doe@example.com",
			}, "123456")
			So(err, ShouldBeNil)
		})

		Convey("with incorrect login ID key", func() {
			_, err := impl.AuthenticateWithLoginID(loginid.LoginID{
				Key:   "phone",
				Value: "john.doe@example.com",
			}, "123456")
			checkErr(err, `{
				"error": {
					"name": "Unauthorized",
					"reason": "InvalidCredentials",
					"message": "invalid credentials",
					"code": 401
				}
			}`)
		})

		Convey("with incorrect password", func() {
			_, err := impl.AuthenticateWithLoginID(loginid.LoginID{
				Value: "john.doe@example.com",
			}, "wrong_password")
			checkErr(err, `{
				"error": {
					"name": "Unauthorized",
					"reason": "InvalidCredentials",
					"message": "invalid credentials",
					"code": 401
				}
			}`)
		})

		Convey("with unknown login id", func() {
			_, err := impl.AuthenticateWithLoginID(loginid.LoginID{
				Value: "foobar",
			}, "123456")
			checkErr(err, `{
				"error": {
					"name": "Unauthorized",
					"reason": "InvalidCredentials",
					"message": "invalid credentials",
					"code": 401
				}
			}`)
		})
	})
}
