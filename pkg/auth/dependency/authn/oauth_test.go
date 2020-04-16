package authn

import (
	"net/url"
	"testing"
	"time"

	authAudit "github.com/skygeario/skygear-server/pkg/auth/dependency/audit"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/hook"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/loginid"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/principal"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/principal/oauth"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/principal/password"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/sso"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/urlprefix"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/userprofile"
	"github.com/skygeario/skygear-server/pkg/auth/model"
	"github.com/skygeario/skygear-server/pkg/core/auth/authinfo"
	"github.com/skygeario/skygear-server/pkg/core/auth/metadata"
	"github.com/skygeario/skygear-server/pkg/core/config"
	"github.com/skygeario/skygear-server/pkg/core/handler"
	. "github.com/skygeario/skygear-server/pkg/core/skytest"
	coreTime "github.com/skygeario/skygear-server/pkg/core/time"
	. "github.com/smartystreets/goconvey/convey"
)

func TestOAuthCoordinator(t *testing.T) {
	Convey("OAuthCoordinator", t, func() {
		authn := &AuthenticateProcess{}
		signup := &SignupProcess{}
		oauthc := &OAuthCoordinator{
			Authn:  authn,
			Signup: signup,
		}

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
		authInfoStore := authinfo.NewMockStoreWithAuthInfoMap(
			map[string]authinfo.AuthInfo{
				"john.doe.id": authinfo.AuthInfo{
					ID: "john.doe.id",
				},
			},
		)
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
			},
		)
		oauthProvider := oauth.NewMockProvider(nil)

		passwordChecker := &authAudit.PasswordChecker{
			PwMinLength: 6,
		}
		timeProvider := &coreTime.MockProvider{TimeNowUTC: time.Date(2006, 1, 2, 15, 4, 5, 0, time.UTC)}
		userProfileStore := userprofile.NewMockUserProfileStore()
		identityProvider := principal.NewMockIdentityProvider(passwordProvider)
		hookProvider := hook.NewMockProvider()
		welcomeEmailConfiguration := &config.WelcomeEmailConfiguration{}
		userVerificationConfiguration := &config.UserVerificationConfiguration{
			LoginIDKeys: []config.UserVerificationKeyConfiguration{
				config.UserVerificationKeyConfiguration{
					Key: "email",
				},
			},
		}
		loginIDChecker := &loginid.MockLoginIDChecker{}
		urlPrefixProvider := urlprefix.Provider{
			Prefix: url.URL{
				Scheme: "http",
				Host:   "example.com",
			},
		}

		signup.PasswordChecker = passwordChecker
		signup.OAuthProvider = oauthProvider
		signup.LoginIDChecker = loginIDChecker
		signup.IdentityProvider = identityProvider
		signup.TimeProvider = timeProvider
		signup.AuthInfoStore = authInfoStore
		signup.UserProfileStore = userProfileStore
		signup.PasswordProvider = passwordProvider
		signup.HookProvider = hookProvider
		signup.WelcomeEmailConfiguration = welcomeEmailConfiguration
		signup.UserVerificationConfiguration = userVerificationConfiguration
		signup.LoginIDConflictConfiguration = &config.AuthAPILoginIDConflictConfiguration{}
		signup.URLPrefixProvider = urlPrefixProvider

		authn.OAuthProvider = oauthProvider
		authn.IdentityProvider = identityProvider
		authn.TimeProvider = timeProvider
		authn.PasswordProvider = passwordProvider

		checkErr := func(err error, errJSON string) {
			So(err, ShouldNotBeNil)
			b, _ := handler.APIResponse{Error: err}.MarshalJSON()
			So(b, ShouldEqualJSON, errJSON)
		}

		Convey("Authenticate", func() {
			Convey("OnUserDuplicateAbort == abort", func() {
				_, _, err := oauthc.AuthenticateCode(sso.AuthInfo{
					ProviderUserInfo: sso.ProviderUserInfo{
						Email: "john.doe@example.com",
					},
				}, "", sso.LoginState{
					OnUserDuplicate: model.OnUserDuplicateAbort,
				})
				checkErr(err, `
			{
				"error": {
					"name": "AlreadyExists",
					"reason": "LoginIDAlreadyUsed",
					"message": "login ID is already used",
					"code": 409
				}
			}
			`)
			})

			Convey("OnUserDuplicateAbort == merge", func() {
				code, _, err := oauthc.AuthenticateCode(sso.AuthInfo{
					ProviderUserInfo: sso.ProviderUserInfo{
						Email: "john.doe@example.com",
					},
				}, "", sso.LoginState{
					OnUserDuplicate: model.OnUserDuplicateMerge,
				})
				So(err, ShouldBeNil)
				So(code.UserID, ShouldEqual, "john.doe.id")
			})

			Convey("OnUserDuplicateAbort == create", func() {
				code, _, err := oauthc.AuthenticateCode(sso.AuthInfo{
					ProviderUserInfo: sso.ProviderUserInfo{
						Email: "john.doe@example.com",
					},
				}, "", sso.LoginState{
					OnUserDuplicate: model.OnUserDuplicateCreate,
				})
				So(err, ShouldBeNil)
				So(code.UserID, ShouldNotEqual, "john.doe.id")
			})
		})

		Convey("Link", func() {
			Convey("never linked before", func() {
				_, _, err := oauthc.LinkCode(sso.AuthInfo{
					ProviderUserInfo: sso.ProviderUserInfo{
						Email: "john.doe@example.com",
					},
				}, "", sso.LinkState{
					UserID: "john.doe.id",
				})
				So(err, ShouldBeNil)
			})

			Convey("already linked", func() {
				providerType := "google"
				providerUserID := "google.a"
				oauthProvider = oauth.NewMockProvider([]*oauth.Principal{
					&oauth.Principal{
						ID:             "a",
						UserID:         "john.doe.id",
						ProviderType:   providerType,
						ProviderKeys:   map[string]interface{}{},
						ProviderUserID: providerUserID,
					},
				})
				signup.OAuthProvider = oauthProvider
				authn.OAuthProvider = oauthProvider

				_, _, err := oauthc.LinkCode(sso.AuthInfo{
					ProviderConfig: config.OAuthProviderConfiguration{
						Type: config.OAuthProviderType(providerType),
					},
					ProviderUserInfo: sso.ProviderUserInfo{
						ID:    providerUserID,
						Email: "john.doe@example.com",
					},
				}, "", sso.LinkState{
					UserID: "john.doe.id",
				})
				checkErr(err, `
				{
					"error": {
						"code": 401,
						"info": {
							"cause": {
								"kind": "AlreadyLinked"
							}
						},
						"message": "user is already linked to this provider",
						"name": "Unauthorized",
						"reason": "SSOFailed"
					}
				}
				`)
			})
		})
	})
}
