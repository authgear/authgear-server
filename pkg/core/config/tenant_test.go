package config

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
	"gopkg.in/yaml.v2"

	"github.com/skygeario/skygear-server/pkg/core/validation"
)

const inputMinimalYAML = `version: '1'
app_id: 66EAFE32-BF5C-4878-8FC8-DD0EEA440981
app_name: myapp
app_config:
  database_url: postgres://
  database_schema: app
user_config:
  clients: []
  master_key: masterkey
  asset:
    secret: assetsecret
  auth:
    authentication_session:
      secret: authnsessionsecret
    login_id_keys:
    - key: email
      type: email
    - key: phone
      type: phone
    - key: username
      type: raw
  hook:
    secret: hooksecret
  sso:
    custom_token:
      secret: customtokensecret
    oauth:
      state_jwt_secret: statejwtsecret
`

const inputMinimalJSON = `
{
	"version": "1",
	"app_id": "66EAFE32-BF5C-4878-8FC8-DD0EEA440981",
	"app_name": "myapp",
	"app_config": {
		"database_url": "postgres://",
		"database_schema": "app"
	},
	"user_config": {
		"clients": [],
		"master_key": "masterkey",
		"asset": {
			"secret": "assetsecret"
		},
		"auth": {
			"authentication_session": {
				"secret": "authnsessionsecret"
			},
			"login_id_keys": [
				{
					"key": "email",
					"type": "email"
				},
				{
					"key": "phone",
					"type": "phone"
				},
				{
					"key": "username",
					"type": "raw"
				}
			],
			"allowed_realms": ["default"]
		},
		"hook": {
			"secret": "hooksecret"
		},
		"sso": {
			"custom_token": {
				"secret": "customtokensecret"
			},
			"oauth": {
				"state_jwt_secret": "statejwtsecret"
			}
		}
	}
}
`

func newInt(i int) *int {
	return &i
}

func makeFullTenantConfig() TenantConfiguration {
	var fullTenantConfig = TenantConfiguration{
		Version: "1",
		AppName: "myapp",
		AppID:   "66EAFE32-BF5C-4878-8FC8-DD0EEA440981",
		AppConfig: AppConfiguration{
			DatabaseURL:    "postgres://user:password@localhost:5432/db?sslmode=disable",
			DatabaseSchema: "app",
			Hook: HookAppConfiguration{
				SyncHookTimeout:      10,
				SyncHookTotalTimeout: 60,
			},
		},
		UserConfig: UserConfiguration{
			Clients: []APIClientConfiguration{
				APIClientConfiguration{
					ID:                   "web-app",
					Name:                 "Web App",
					APIKey:               "api_key",
					SessionTransport:     SessionTransportTypeHeader,
					AccessTokenLifetime:  1800,
					SessionIdleTimeout:   300,
					RefreshTokenLifetime: 86400,
					SameSite:             SessionCookieSameSiteLax,
				},
			},
			MasterKey: "mymasterkey",
			CORS: CORSConfiguration{
				Origin: "localhost:3000",
			},
			Asset: AssetConfiguration{
				Secret: "assetsecret",
			},
			Auth: AuthConfiguration{
				AuthenticationSession: AuthenticationSessionConfiguration{
					Secret: "authnsessionsecret",
				},
				LoginIDKeys: []LoginIDKeyConfiguration{
					LoginIDKeyConfiguration{
						Key:     "email",
						Type:    LoginIDKeyType("email"),
						Minimum: newInt(0),
						Maximum: newInt(1),
					},
					LoginIDKeyConfiguration{
						Key:     "phone",
						Type:    LoginIDKeyType("phone"),
						Minimum: newInt(0),
						Maximum: newInt(1),
					},
					LoginIDKeyConfiguration{
						Key:     "username",
						Type:    LoginIDKeyTypeRaw,
						Minimum: newInt(0),
						Maximum: newInt(1),
					},
				},
				AllowedRealms:              []string{"default"},
				OnUserDuplicateAllowCreate: true,
			},
			MFA: MFAConfiguration{
				Enabled:     true,
				Enforcement: MFAEnforcementOptional,
				Maximum:     newInt(3),
				TOTP: MFATOTPConfiguration{
					Maximum: 1,
				},
				OOB: MFAOOBConfiguration{
					SMS: MFAOOBSMSConfiguration{
						Maximum: 1,
					},
					Email: MFAOOBEmailConfiguration{
						Maximum: 1,
					},
				},
				BearerToken: MFABearerTokenConfiguration{
					ExpireInDays: 60,
				},
				RecoveryCode: MFARecoveryCodeConfiguration{
					Count:       24,
					ListEnabled: true,
				},
			},
			UserAudit: UserAuditConfiguration{
				Enabled:         true,
				TrailHandlerURL: "http://localhost:3000/useraudit",
			},
			PasswordPolicy: PasswordPolicyConfiguration{
				MinLength:             8,
				UppercaseRequired:     true,
				LowercaseRequired:     true,
				DigitRequired:         true,
				SymbolRequired:        true,
				MinimumGuessableLevel: 4,
				ExcludedKeywords:      []string{"admin", "password", "secret"},
				HistorySize:           10,
				HistoryDays:           90,
				ExpiryDays:            30,
			},
			ForgotPassword: ForgotPasswordConfiguration{
				AppName:             "myapp",
				SecureMatch:         true,
				Sender:              "myforgotpasswordsender",
				Subject:             "myforgotpasswordsubject",
				ReplyTo:             "myforgotpasswordreplyto",
				ResetURLLifetime:    60,
				SuccessRedirect:     "http://localhost:3000/forgotpassword/success",
				ErrorRedirect:       "http://localhost:3000/forgotpassword/error",
				EmailTextURL:        "http://localhost:3000/forgotpassword/text",
				EmailHTMLURL:        "http://localhost:3000/forgotpassword/html",
				ResetHTMLURL:        "http://localhost:3000/forgotpassword/reset",
				ResetSuccessHTMLURL: "http://localhost:3000/forgotpassword/reset/success",
				ResetErrorHTMLURL:   "http://localhost:3000/forgotpassword/reset/error",
			},
			WelcomeEmail: WelcomeEmailConfiguration{
				Enabled:     true,
				Sender:      "welcomeemailsender",
				Subject:     "welcomeemailsubject",
				ReplyTo:     "welcomeemailreplyto",
				TextURL:     "http://localhost:3000/welcomeemail/text",
				HTMLURL:     "http://localhost:3000/welcomeemail/html",
				Destination: "first",
			},
			SSO: SSOConfiguration{
				CustomToken: CustomTokenConfiguration{
					Enabled:                    true,
					Issuer:                     "customtokenissuer",
					Audience:                   "customtokenaudience",
					Secret:                     "customtokensecret",
					OnUserDuplicateAllowMerge:  true,
					OnUserDuplicateAllowCreate: true,
				},
				OAuth: OAuthConfiguration{
					StateJWTSecret: "oauthstatejwtsecret",
					AllowedCallbackURLs: []string{
						"http://localhost:3000/oauth/callback",
					},
					ExternalAccessTokenFlowEnabled: true,
					OnUserDuplicateAllowMerge:      true,
					OnUserDuplicateAllowCreate:     true,
					Providers: []OAuthProviderConfiguration{
						OAuthProviderConfiguration{
							ID:           "google",
							Type:         "google",
							ClientID:     "googleclientid",
							ClientSecret: "googleclientsecret",
							Scope:        "email profile",
						},
						OAuthProviderConfiguration{
							ID:           "azure-id-1",
							Type:         "azureadv2",
							ClientID:     "azureclientid",
							ClientSecret: "azureclientsecret",
							Scope:        "email",
							Tenant:       "azure-id-1",
						},
					},
				},
			},
			UserVerification: UserVerificationConfiguration{
				AutoSendOnSignup: true,
				Criteria:         "any",
				ErrorRedirect:    "http://localhost:3000/userverification/error",
				ErrorHTMLURL:     "http://localhost:3000/userverification/error.html",
				LoginIDKeys: []UserVerificationKeyConfiguration{
					UserVerificationKeyConfiguration{
						Key:             "email",
						CodeFormat:      "complex",
						Expiry:          3600,
						SuccessRedirect: "http://localhost:3000/userverification/success",
						SuccessHTMLURL:  "http://localhost:3000/userverification/success.html",
						ErrorRedirect:   "http://localhost:3000/userverification/error",
						ErrorHTMLURL:    "http://localhost:3000/userverification/error.html",
						ProviderConfig: UserVerificationProviderConfiguration{
							Subject: "userverificationsubject",
							Sender:  "userverificationsender",
							ReplyTo: "userverificationreplyto",
							TextURL: "http://localhost:3000/userverification/text",
							HTMLURL: "http://localhost:3000/userverification/html",
						},
					},
				},
			},
			Hook: HookUserConfiguration{
				Secret: "hook-secret",
			},
			SMTP: SMTPConfiguration{
				Host:     "localhost",
				Port:     465,
				Mode:     "ssl",
				Login:    "user",
				Password: "password",
			},
			Twilio: TwilioConfiguration{
				AccountSID: "mytwilioaccountsid",
				AuthToken:  "mytwilioauthtoken",
				From:       "mytwilio",
			},
			Nexmo: NexmoConfiguration{
				APIKey:    "mynexmoapikey",
				APISecret: "mynexmoapisecret",
				From:      "mynexmo",
			},
		},
		Hooks: []Hook{
			Hook{
				Event: "after_signup",
				URL:   "http://localhost:3000/after_signup",
			},
			Hook{
				Event: "after_signup",
				URL:   "http://localhost:3000/after_signup",
			},
		},
		DeploymentRoutes: []DeploymentRoute{
			DeploymentRoute{
				Version: "a",
				Path:    "/",
				Type:    "http-service",
				TypeConfig: map[string]interface{}{
					"backend_url": "http://localhost:3000",
				},
			},
			DeploymentRoute{
				Version: "a",
				Path:    "/api",
				Type:    "http-service",
				TypeConfig: map[string]interface{}{
					"backend_url": "http://localhost:3001",
				},
			},
		},
	}

	return fullTenantConfig
}

func TestTenantConfig(t *testing.T) {
	Convey("Test TenantConfiguration", t, func() {
		// YAML
		Convey("should load tenant config from YAML", func() {
			c, err := NewTenantConfigurationFromYAML(strings.NewReader(inputMinimalYAML))
			So(err, ShouldBeNil)

			So(c.Version, ShouldEqual, "1")
			So(c.AppName, ShouldEqual, "myapp")
			So(c.AppConfig.DatabaseURL, ShouldEqual, "postgres://")
			So(c.UserConfig.Clients, ShouldBeEmpty)
			So(c.UserConfig.MasterKey, ShouldEqual, "masterkey")
		})
		Convey("should have default value when load from YAML", func() {
			c, err := NewTenantConfigurationFromYAML(strings.NewReader(inputMinimalYAML))
			So(err, ShouldBeNil)
			So(c.UserConfig.CORS.Origin, ShouldEqual, "*")
			So(c.UserConfig.SMTP.Port, ShouldEqual, 25)
		})
		Convey("should validate when load from YAML", func() {
			invalidInput := `
app_id: 66EAFE32-BF5C-4878-8FC8-DD0EEA440981
app_name: myapp
app_config:
  database_url: postgres://
  database_schema: app
user_config:
  clients: []
  master_key: masterkey
`
			_, err := NewTenantConfigurationFromYAML(strings.NewReader(invalidInput))
			So(validation.ErrorCauses(err), ShouldResemble, []validation.ErrorCause{{
				Kind:    validation.ErrorGeneral,
				Message: "only version 1 is supported",
				Pointer: "/version",
			}})
		})
		// JSON
		Convey("should have default value when load from JSON", func() {
			c, err := NewTenantConfigurationFromJSON(strings.NewReader(inputMinimalJSON), false)
			So(err, ShouldBeNil)
			So(c.Version, ShouldEqual, "1")
			So(c.AppName, ShouldEqual, "myapp")
			So(c.AppConfig.DatabaseURL, ShouldEqual, "postgres://")
			So(c.UserConfig.Clients, ShouldBeEmpty)
			So(c.UserConfig.MasterKey, ShouldEqual, "masterkey")
			So(c.UserConfig.CORS.Origin, ShouldEqual, "*")
			So(c.UserConfig.SMTP.Port, ShouldEqual, 25)
		})
		Convey("should validate when load from JSON", func() {
			invalidInput := `
		{
		  "app_id": "66EAFE32-BF5C-4878-8FC8-DD0EEA440981",
		  "app_name": "myapp",
		  "app_config": {
		    "database_url": "postgres://",
		    "database_schema": "app"
		  },
		  "user_config": {
		    "api_key": "apikey",
		    "master_key": "masterkey",
		    "welcome_email": {
		      "enabled": true
		    }
		  }
		}
					`
			_, err := NewTenantConfigurationFromJSON(strings.NewReader(invalidInput), false)
			So(validation.ErrorCauses(err), ShouldResemble, []validation.ErrorCause{{
				Kind:    validation.ErrorGeneral,
				Message: "only version 1 is supported",
				Pointer: "/version",
			}})
		})
		// Conversion
		Convey("should be losslessly converted between Go and msgpack", func() {
			c := makeFullTenantConfig()
			base64msgpack, err := c.StdBase64Msgpack()
			So(err, ShouldBeNil)
			cc, err := NewTenantConfigurationFromStdBase64Msgpack(base64msgpack)
			So(err, ShouldBeNil)
			So(c, ShouldResemble, *cc)
		})
		Convey("should be losslessly converted between Go and JSON", func() {
			c := makeFullTenantConfig()
			b, err := json.Marshal(c)
			So(err, ShouldBeNil)
			cc, err := NewTenantConfigurationFromJSON(bytes.NewReader(b), false)
			So(err, ShouldBeNil)
			So(c, ShouldResemble, *cc)
		})
		Convey("should be losslessly converted between Go and YAML", func() {
			c := makeFullTenantConfig()
			b, err := yaml.Marshal(c)
			So(err, ShouldBeNil)
			cc, err := NewTenantConfigurationFromYAML(bytes.NewReader(b))
			So(err, ShouldBeNil)
			So(c, ShouldResemble, *cc)
		})
		Convey("should set OAuth provider id and default scope", func() {
			c := makeFullTenantConfig()
			c.UserConfig.SSO.OAuth.Providers = []OAuthProviderConfiguration{
				OAuthProviderConfiguration{
					Type:         OAuthProviderTypeGoogle,
					ClientID:     "googleclientid",
					ClientSecret: "googleclientsecret",
				},
			}
			c.AfterUnmarshal()

			google := c.UserConfig.SSO.OAuth.Providers[0]

			So(google.ID, ShouldEqual, OAuthProviderTypeGoogle)
			So(google.Scope, ShouldEqual, "profile email")
		})
		Convey("should validate api key != master key", func() {
			c := makeFullTenantConfig()
			for i := range c.UserConfig.Clients {
				if c.UserConfig.Clients[i].ID == "web-app" {
					c.UserConfig.Clients[i].APIKey = c.UserConfig.MasterKey
				}
			}
			err := c.Validate()
			So(validation.ErrorCauses(err), ShouldResemble, []validation.ErrorCause{{
				Kind:    validation.ErrorGeneral,
				Message: "master key must not be same as API key",
				Pointer: "/user_config/master_key",
			}})
		})
		Convey("should validate minimum <= maximum", func() {
			c := makeFullTenantConfig()
			loginIDKeys := c.UserConfig.Auth.LoginIDKeys
			for i := range loginIDKeys {
				if loginIDKeys[i].Key == "email" {
					loginIDKeys[i].Minimum = newInt(2)
					loginIDKeys[i].Maximum = newInt(1)
				}
			}
			c.UserConfig.Auth.LoginIDKeys = loginIDKeys
			err := c.Validate()
			So(validation.ErrorCauses(err), ShouldResemble, []validation.ErrorCause{{
				Kind:    validation.ErrorGeneral,
				Message: "invalid login ID amount range",
				Pointer: "/user_config/auth/login_id_keys/0",
			}})
		})
		Convey("UserVerification.LoginIDKeys is subset of Auth.LoginIDKeys", func() {
			c := makeFullTenantConfig()
			c.UserConfig.UserVerification.LoginIDKeys = append(
				c.UserConfig.UserVerification.LoginIDKeys,
				UserVerificationKeyConfiguration{
					Key: "invalid",
				},
			)
			err := c.Validate()
			So(validation.ErrorCauses(err), ShouldResemble, []validation.ErrorCause{{
				Kind:    validation.ErrorGeneral,
				Message: "cannot verify disallowed login ID key",
				Pointer: "/user_config/user_verification/login_id_keys/invalid",
			}})
		})
		Convey("should validate OAuth Provider", func() {
			c := makeFullTenantConfig()
			c.UserConfig.SSO.OAuth.Providers = []OAuthProviderConfiguration{
				OAuthProviderConfiguration{
					ID:           "azure",
					Type:         OAuthProviderTypeAzureADv2,
					ClientID:     "clientid",
					ClientSecret: "clientsecret",
					Tenant:       "tenant",
				},
				OAuthProviderConfiguration{
					ID:           "azure",
					Type:         OAuthProviderTypeAzureADv2,
					ClientID:     "clientid",
					ClientSecret: "clientsecret",
					Tenant:       "tenant",
				},
			}

			err := c.Validate()
			So(validation.ErrorCauses(err), ShouldResemble, []validation.ErrorCause{{
				Kind:    validation.ErrorGeneral,
				Message: "duplicated OAuth provider",
				Pointer: "/user_config/sso/oauth/providers/1",
			}})
		})
	})
}
