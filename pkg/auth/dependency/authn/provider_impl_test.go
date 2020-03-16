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
	"github.com/skygeario/skygear-server/pkg/auth/event"
	"github.com/skygeario/skygear-server/pkg/auth/model"
	"github.com/skygeario/skygear-server/pkg/auth/task/spec"
	"github.com/skygeario/skygear-server/pkg/core/async"
	"github.com/skygeario/skygear-server/pkg/core/auth/authinfo"
	"github.com/skygeario/skygear-server/pkg/core/auth/metadata"
	"github.com/skygeario/skygear-server/pkg/core/config"
	"github.com/skygeario/skygear-server/pkg/core/handler"
	. "github.com/skygeario/skygear-server/pkg/core/skytest"
	coreTime "github.com/skygeario/skygear-server/pkg/core/time"
	. "github.com/smartystreets/goconvey/convey"
)

func TestCreateUserWithLoginIDs(t *testing.T) {
	Convey("CreateUserWithLoginIDs", t, func() {
		impl := &ProviderImpl{}
		two := 2
		loginIDsKeys := []config.LoginIDKeyConfiguration{
			config.LoginIDKeyConfiguration{
				Key:     "email",
				Type:    config.LoginIDKeyType(metadata.Email),
				Maximum: &two,
			},
			config.LoginIDKeyConfiguration{
				Key:     "username",
				Type:    config.LoginIDKeyTypeRaw,
				Maximum: &two,
			},
		}
		allowedRealms := []string{password.DefaultRealm}
		authInfoStore := authinfo.NewMockStore()
		passwordProvider := password.NewMockProvider(loginIDsKeys, allowedRealms)
		oauthProvider := oauth.NewMockProvider([]*oauth.Principal{
			&oauth.Principal{
				ID:           "john.doe.id",
				UserID:       "john.doe.id",
				ProviderType: "google",
				ProviderKeys: map[string]interface{}{},
				ClaimsValue: map[string]interface{}{
					"email": "john.doe@example.com",
				},
			},
		})

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
		authConfiguration := &config.AuthConfiguration{
			LoginIDKeys: loginIDsKeys,
		}
		loginIDChecker := &loginid.MockLoginIDChecker{}
		urlPrefixProvider := urlprefix.Provider{
			Prefix: url.URL{
				Scheme: "http",
				Host:   "example.com",
			},
		}
		taskQueue := async.NewMockQueue()

		impl.PasswordChecker = passwordChecker
		impl.OAuthProvider = oauthProvider
		impl.LoginIDChecker = loginIDChecker
		impl.IdentityProvider = identityProvider
		impl.TimeProvider = timeProvider
		impl.AuthInfoStore = authInfoStore
		impl.UserProfileStore = userProfileStore
		impl.PasswordProvider = passwordProvider
		impl.HookProvider = hookProvider
		impl.WelcomeEmailConfiguration = welcomeEmailConfiguration
		impl.UserVerificationConfiguration = userVerificationConfiguration
		impl.AuthConfiguration = authConfiguration
		impl.URLPrefixProvider = urlPrefixProvider
		impl.TaskQueue = taskQueue

		checkErr := func(err error, errJSON string) {
			So(err, ShouldNotBeNil)
			b, _ := handler.APIResponse{Error: err}.MarshalJSON()
			So(b, ShouldEqualJSON, errJSON)
		}

		Convey("detect duplicated login ID", func() {
			_, _, _, err := impl.CreateUserWithLoginIDs(
				[]loginid.LoginID{
					{Key: "username", Value: "john.doe"},
					{Key: "email", Value: "john.doe"},
				},
				"pass",
				map[string]interface{}{},
				model.OnUserDuplicateAbort,
			)
			checkErr(err, `{
				"error": {
					"name": "Invalid",
					"reason": "ValidationFailed",
					"message": "invalid request body",
					"code": 400,
					"info": {
						"causes": [
							{
								"kind": "General",
								"pointer": "/login_ids/1/value",
								"message": "duplicated login ID"
							}
						]
					}
				}
			}`)
		})

		Convey("abort if user duplicate with oauth", func() {
			impl.IdentityProvider = principal.NewMockIdentityProvider(
				passwordProvider,
				oauthProvider,
			)
			_, _, _, err := impl.CreateUserWithLoginIDs(
				[]loginid.LoginID{
					{Key: "email", Value: "john.doe@example.com"},
					{Key: "username", Value: "john.doe"},
				},
				"123456",
				map[string]interface{}{},
				model.OnUserDuplicateAbort,
			)
			checkErr(err, `{
				"error": {
					"name": "AlreadyExists",
					"reason": "LoginIDAlreadyUsed",
					"message": "login ID is already used",
					"code": 409
				}
			}`)
		})

		Convey("create new even duplicate", func() {
			impl.IdentityProvider = principal.NewMockIdentityProvider(
				passwordProvider,
				oauthProvider,
			)
			impl.AuthConfiguration.OnUserDuplicateAllowCreate = true
			_, _, _, err := impl.CreateUserWithLoginIDs(
				[]loginid.LoginID{
					{Key: "email", Value: "john.doe@example.com"},
					{Key: "username", Value: "john.doe"},
				},
				"123456",
				map[string]interface{}{},
				model.OnUserDuplicateCreate,
			)
			So(err, ShouldBeNil)
		})

		Convey("weak password", func() {
			_, _, _, err := impl.CreateUserWithLoginIDs(
				[]loginid.LoginID{
					{Key: "username", Value: "john.doe"},
				},
				"weak",
				map[string]interface{}{},
				model.OnUserDuplicateAbort,
			)
			checkErr(err, `{
				"error": {
					"name": "Invalid",
					"reason": "PasswordPolicyViolated",
					"message": "password policy violated",
					"code": 400,
					"info": {
						"causes": [
							{ "kind": "PasswordTooShort", "min_length": 6, "pw_length": 4 }
						]
					}
				}
			}`)
		})

		Convey("hook", func() {
			now := timeProvider.NowUTC()
			authInfo, _, _, err := impl.CreateUserWithLoginIDs(
				[]loginid.LoginID{
					{Key: "email", Value: "john.doe@example.com"},
					{Key: "username", Value: "john.doe"},
				},
				"123456",
				map[string]interface{}{},
				model.OnUserDuplicateAbort,
			)
			So(err, ShouldBeNil)

			var p password.Principal
			err = passwordProvider.GetPrincipalByLoginIDWithRealm("email", "john.doe@example.com", password.DefaultRealm, &p)
			So(err, ShouldBeNil)
			var p2 password.Principal
			err = passwordProvider.GetPrincipalByLoginIDWithRealm("username", "john.doe", password.DefaultRealm, &p2)
			So(err, ShouldBeNil)

			So(hookProvider.DispatchedEvents, ShouldResemble, []event.Payload{
				event.UserCreateEvent{
					User: model.User{
						ID:          authInfo.ID,
						LastLoginAt: &now,
						Verified:    false,
						Disabled:    false,
						VerifyInfo:  map[string]bool{},
						Metadata:    userprofile.Data{},
					},
					Identities: []model.Identity{
						model.Identity{
							ID:   p.ID,
							Type: "password",
							Attributes: principal.Attributes{
								"login_id_key": "email",
								"login_id":     "john.doe@example.com",
							},
							Claims: principal.Claims{
								"email": "john.doe@example.com",
							},
						},
						model.Identity{
							ID:   p2.ID,
							Type: "password",
							Attributes: principal.Attributes{
								"login_id_key": "username",
								"login_id":     "john.doe",
							},
							Claims: principal.Claims{},
						},
					},
				},
			})
		})

		Convey("send welcome email to the first login ID", func() {
			impl.WelcomeEmailConfiguration.Enabled = true
			impl.WelcomeEmailConfiguration.Destination = config.WelcomeEmailDestinationFirst

			_, _, _, err := impl.CreateUserWithLoginIDs(
				[]loginid.LoginID{
					{Key: "email", Value: "john.doe+1@example.com"},
					{Key: "username", Value: "john.doe"},
					{Key: "email", Value: "john.doe+2@example.com"},
				},
				"123456",
				map[string]interface{}{},
				model.OnUserDuplicateAbort,
			)
			So(err, ShouldBeNil)

			So(len(taskQueue.TasksName), ShouldEqual, 1)
			So(len(taskQueue.TasksParam), ShouldEqual, 1)
			So(taskQueue.TasksName[0], ShouldEqual, spec.WelcomeEmailSendTaskName)
			param := taskQueue.TasksParam[0].(spec.WelcomeEmailSendTaskParam)
			So(param.Email, ShouldEqual, "john.doe+1@example.com")
		})

		Convey("send welcome email to all login IDs", func() {
			impl.WelcomeEmailConfiguration.Enabled = true
			impl.WelcomeEmailConfiguration.Destination = config.WelcomeEmailDestinationAll

			_, _, _, err := impl.CreateUserWithLoginIDs(
				[]loginid.LoginID{
					{Key: "email", Value: "john.doe+1@example.com"},
					{Key: "username", Value: "john.doe"},
					{Key: "email", Value: "john.doe+2@example.com"},
				},
				"123456",
				map[string]interface{}{},
				model.OnUserDuplicateAbort,
			)
			So(err, ShouldBeNil)

			So(taskQueue.TasksName, ShouldHaveLength, 2)
			So(taskQueue.TasksParam, ShouldHaveLength, 2)

			So(taskQueue.TasksName[0], ShouldEqual, spec.WelcomeEmailSendTaskName)
			So(taskQueue.TasksName[1], ShouldEqual, spec.WelcomeEmailSendTaskName)

			So(taskQueue.TasksParam[0].(spec.WelcomeEmailSendTaskParam).Email, ShouldEqual, "john.doe+1@example.com")
			So(taskQueue.TasksParam[1].(spec.WelcomeEmailSendTaskParam).Email, ShouldEqual, "john.doe+2@example.com")
		})

		Convey("send verification code to all login IDs", func() {
			impl.UserVerificationConfiguration.AutoSendOnSignup = true

			_, _, _, err := impl.CreateUserWithLoginIDs(
				[]loginid.LoginID{
					{Key: "email", Value: "john.doe+1@example.com"},
					{Key: "username", Value: "john.doe"},
					{Key: "email", Value: "john.doe+2@example.com"},
				},
				"123456",
				map[string]interface{}{},
				model.OnUserDuplicateAbort,
			)
			So(err, ShouldBeNil)

			So(taskQueue.TasksName, ShouldHaveLength, 2)
			So(taskQueue.TasksParam, ShouldHaveLength, 2)

			So(taskQueue.TasksName[0], ShouldEqual, spec.VerifyCodeSendTaskName)
			So(taskQueue.TasksName[1], ShouldEqual, spec.VerifyCodeSendTaskName)

			So(taskQueue.TasksParam[0].(spec.VerifyCodeSendTaskParam).LoginID, ShouldEqual, "john.doe+1@example.com")
			So(taskQueue.TasksParam[1].(spec.VerifyCodeSendTaskParam).LoginID, ShouldEqual, "john.doe+2@example.com")
		})
	})
}

func TestAuthenticateWithLoginID(t *testing.T) {
	Convey("AuthenticateWithLoginID", t, func() {
		impl := &ProviderImpl{}
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
		authConfiguration := &config.AuthConfiguration{
			LoginIDKeys: loginIDsKeys,
		}
		loginIDChecker := &loginid.MockLoginIDChecker{}
		urlPrefixProvider := urlprefix.Provider{
			Prefix: url.URL{
				Scheme: "http",
				Host:   "example.com",
			},
		}

		impl.PasswordChecker = passwordChecker
		impl.OAuthProvider = oauthProvider
		impl.LoginIDChecker = loginIDChecker
		impl.IdentityProvider = identityProvider
		impl.TimeProvider = timeProvider
		impl.AuthInfoStore = authInfoStore
		impl.UserProfileStore = userProfileStore
		impl.PasswordProvider = passwordProvider
		impl.HookProvider = hookProvider
		impl.WelcomeEmailConfiguration = welcomeEmailConfiguration
		impl.UserVerificationConfiguration = userVerificationConfiguration
		impl.AuthConfiguration = authConfiguration
		impl.URLPrefixProvider = urlPrefixProvider

		checkErr := func(err error, errJSON string) {
			So(err, ShouldNotBeNil)
			b, _ := handler.APIResponse{Error: err}.MarshalJSON()
			So(b, ShouldEqualJSON, errJSON)
		}

		Convey("without login ID key", func() {
			_, _, err := impl.AuthenticateWithLoginID(loginid.LoginID{
				Value: "john.doe@example.com",
			}, "123456")
			So(err, ShouldBeNil)
		})

		Convey("with correct login ID key", func() {
			_, _, err := impl.AuthenticateWithLoginID(loginid.LoginID{
				Key:   "email",
				Value: "john.doe@example.com",
			}, "123456")
			So(err, ShouldBeNil)
		})

		Convey("with incorrect login ID key", func() {
			_, _, err := impl.AuthenticateWithLoginID(loginid.LoginID{
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
			_, _, err := impl.AuthenticateWithLoginID(loginid.LoginID{
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
			_, _, err := impl.AuthenticateWithLoginID(loginid.LoginID{
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

func TestAuthenticateWithOAuth(t *testing.T) {
	Convey("AuthenticateWithOAuth", t, func() {
		impl := &ProviderImpl{}
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
		authConfiguration := &config.AuthConfiguration{
			LoginIDKeys: loginIDsKeys,
		}
		loginIDChecker := &loginid.MockLoginIDChecker{}
		urlPrefixProvider := urlprefix.Provider{
			Prefix: url.URL{
				Scheme: "http",
				Host:   "example.com",
			},
		}

		impl.PasswordChecker = passwordChecker
		impl.OAuthProvider = oauthProvider
		impl.LoginIDChecker = loginIDChecker
		impl.IdentityProvider = identityProvider
		impl.TimeProvider = timeProvider
		impl.AuthInfoStore = authInfoStore
		impl.UserProfileStore = userProfileStore
		impl.PasswordProvider = passwordProvider
		impl.HookProvider = hookProvider
		impl.WelcomeEmailConfiguration = welcomeEmailConfiguration
		impl.UserVerificationConfiguration = userVerificationConfiguration
		impl.AuthConfiguration = authConfiguration
		impl.URLPrefixProvider = urlPrefixProvider

		checkErr := func(err error, errJSON string) {
			So(err, ShouldNotBeNil)
			b, _ := handler.APIResponse{Error: err}.MarshalJSON()
			So(b, ShouldEqualJSON, errJSON)
		}

		Convey("OnUserDuplicateAbort == abort", func() {
			_, err := impl.AuthenticateWithOAuth(sso.AuthInfo{
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
			code, err := impl.AuthenticateWithOAuth(sso.AuthInfo{
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
			code, err := impl.AuthenticateWithOAuth(sso.AuthInfo{
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
}

func TestLinkOAuth(t *testing.T) {
	Convey("LinkOAuth", t, func() {
		impl := &ProviderImpl{}
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
		authConfiguration := &config.AuthConfiguration{
			LoginIDKeys: loginIDsKeys,
		}
		loginIDChecker := &loginid.MockLoginIDChecker{}
		urlPrefixProvider := urlprefix.Provider{
			Prefix: url.URL{
				Scheme: "http",
				Host:   "example.com",
			},
		}

		impl.PasswordChecker = passwordChecker
		impl.OAuthProvider = oauthProvider
		impl.LoginIDChecker = loginIDChecker
		impl.IdentityProvider = identityProvider
		impl.TimeProvider = timeProvider
		impl.AuthInfoStore = authInfoStore
		impl.UserProfileStore = userProfileStore
		impl.PasswordProvider = passwordProvider
		impl.HookProvider = hookProvider
		impl.WelcomeEmailConfiguration = welcomeEmailConfiguration
		impl.UserVerificationConfiguration = userVerificationConfiguration
		impl.AuthConfiguration = authConfiguration
		impl.URLPrefixProvider = urlPrefixProvider

		checkErr := func(err error, errJSON string) {
			So(err, ShouldNotBeNil)
			b, _ := handler.APIResponse{Error: err}.MarshalJSON()
			So(b, ShouldEqualJSON, errJSON)
		}

		Convey("never linked before", func() {
			_, err := impl.LinkOAuth(sso.AuthInfo{
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
			impl.OAuthProvider = oauth.NewMockProvider([]*oauth.Principal{
				&oauth.Principal{
					ID:             "a",
					UserID:         "john.doe.id",
					ProviderType:   providerType,
					ProviderKeys:   map[string]interface{}{},
					ProviderUserID: providerUserID,
				},
			})
			_, err := impl.LinkOAuth(sso.AuthInfo{
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
}
