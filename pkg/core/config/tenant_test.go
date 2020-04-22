package config

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strings"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
	"gopkg.in/yaml.v2"

	"github.com/skygeario/skygear-server/pkg/core/apiversion"
	"github.com/skygeario/skygear-server/pkg/core/marshal"
	. "github.com/skygeario/skygear-server/pkg/core/skytest"
	"github.com/skygeario/skygear-server/pkg/core/validation"
)

var inputMinimalYAML = fmt.Sprintf(`api_version: %s
app_id: 66EAFE32-BF5C-4878-8FC8-DD0EEA440981
app_name: myapp
database_config:
  database_url: postgres://
  database_schema: app
app_config:
  api_version: %s
  clients: []
  master_key: masterkey
  asset:
    secret: assetsecret
  authentication:
      secret: authnsessionsecret
  hook:
    secret: hooksecret
  identity:
    login_id:
      keys:
      - key: email
        type: email
      - key: phone
        type: phone
      - key: username
        type: raw
    oauth:
      state_jwt_secret: statejwtsecret
`, apiversion.APIVersion, apiversion.APIVersion)

var inputMinimalJSON = fmt.Sprintf(`
{
	"api_version": "%s",
	"app_id": "66EAFE32-BF5C-4878-8FC8-DD0EEA440981",
	"app_name": "myapp",
	"database_config": {
		"database_url": "postgres://",
		"database_schema": "app"
	},
	"app_config": {
		"api_version": "%s",
		"master_key": "masterkey",
		"asset": {
			"secret": "assetsecret"
		},
		"authentication": {
			"secret": "authnsessionsecret",
			"secondary_authenticators" : null
		},
		"hook": {
			"secret": "hooksecret"
		},
		"identity": {
			"login_id": {
				"keys": [
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
				]
			},
			"oauth": {
				"state_jwt_secret": "statejwtsecret"
			}
		}
	}
}
`, apiversion.APIVersion, apiversion.APIVersion)

func newInt(i int) *int {
	return &i
}

func makeFullTenantConfig() TenantConfiguration {
	newFalse := func() *bool {
		b := false
		return &b
	}
	newTrue := func() *bool {
		b := true
		return &b
	}
	newStr := func(s string) *string {
		return &s
	}
	var fullTenantConfig = TenantConfiguration{
		APIVersion: apiversion.APIVersion,
		AppName:    "myapp",
		AppID:      "66EAFE32-BF5C-4878-8FC8-DD0EEA440981",
		DatabaseConfig: &DatabaseConfiguration{
			DatabaseURL:    "postgres://user:password@localhost:5432/db?sslmode=disable",
			DatabaseSchema: "app",
		},
		Hook: &HookTenantConfiguration{
			SyncHookTimeout:      10,
			SyncHookTotalTimeout: 60,
		},
		AppConfig: &AppConfiguration{
			APIVersion:     apiversion.APIVersion,
			DisplayAppName: "MyApp",
			Session: &SessionConfiguration{
				Lifetime:            86400,
				IdleTimeoutEnabled:  true,
				IdleTimeout:         3600,
				CookieDomain:        newStr("example.com"),
				CookieNonPersistent: true,
			},
			Clients: []OAuthClientConfiguration{
				OAuthClientConfiguration{
					"client_name": "Web App",
					"client_id":   "web_app",
					"redirect_uris": []interface{}{
						"http://localhost:8081/oauth2/continue",
					},
					"access_token_lifetime":  1800.0,
					"refresh_token_lifetime": 86400.0,
				},
			},
			MasterKey: "mymasterkey",
			CORS: &CORSConfiguration{
				Origin: "localhost:3000",
			},
			Asset: &AssetConfiguration{
				Secret: "assetsecret",
			},
			OIDC: &OIDCConfiguration{
				Keys: []OIDCSigningKeyConfiguration{
					OIDCSigningKeyConfiguration{
						KID:        "k1",
						PublicKey:  "content of .pem",
						PrivateKey: "content of .pem",
					},
				},
			},
			AuthUI: &AuthUIConfiguration{
				CSS: "a { color: red; }",
				CountryCallingCode: &AuthUICountryCallingCodeConfiguration{
					Values:  []string{"852"},
					Default: "852",
				},
			},
			AuthAPI: &AuthAPIConfiguration{
				Enabled: true,
				OnIdentityConflict: &AuthAPIIdentityConflictConfiguration{
					LoginID: &AuthAPILoginIDConflictConfiguration{
						AllowCreateNewUser: true,
					},
					OAuth: &AuthAPIOAuthConflictConfiguration{
						AllowCreateNewUser: true,
						AllowAutoMergeUser: true,
					},
				},
			},
			Authentication: &AuthenticationConfiguration{
				Secret:                      "authnsessionsecret",
				Identities:                  []string{"login_id", "oauth"},
				PrimaryAuthenticators:       []string{"oauth", "password"},
				SecondaryAuthenticators:     []string{"otp", "bearer_token"},
				SecondaryAuthenticationMode: SecondaryAuthenticationModeIfExists,
			},
			Authenticator: &AuthenticatorConfiguration{
				Password: &AuthenticatorPasswordConfiguration{
					Policy: &PasswordPolicyConfiguration{
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
				},
				TOTP: &AuthenticatorTOTPConfiguration{
					Maximum: newInt(99),
				},
				OOB: &AuthenticatorOOBConfiguration{
					SMS: &AuthenticatorOOBSMSConfiguration{
						Maximum: newInt(99),
						Message: SMSMessageConfiguration{
							"sender": "+85212345678",
						},
					},
					Email: &AuthenticatorOOBEmailConfiguration{
						Maximum: newInt(99),
						Message: EmailMessageConfiguration{
							"sender":   `"MFA Sender" <mfaoobsender@example.com>`,
							"subject":  "mfaoobsubject",
							"reply_to": `"MFA Reply To" <mfaoobreplyto@example.com>`,
						},
					},
				},
				BearerToken: &AuthenticatorBearerTokenConfiguration{
					ExpireInDays: 60,
				},
				RecoveryCode: &AuthenticatorRecoveryCodeConfiguration{
					Count:       24,
					ListEnabled: true,
				},
			},
			ForgotPassword: &ForgotPasswordConfiguration{
				SecureMatch: true,
				EmailMessage: EmailMessageConfiguration{
					"sender":   `"Forgot Password Sender" <myforgotpasswordsender@example.com>`,
					"subject":  "myforgotpasswordsubject",
					"reply_to": `"Forgot Password Reply To" <myforgotpasswordreplyto@example.com>`,
				},
				ResetURLLifetime: 60,
				SuccessRedirect:  "http://localhost:3000/forgotpassword/success",
				ErrorRedirect:    "http://localhost:3000/forgotpassword/error",
			},
			WelcomeEmail: &WelcomeEmailConfiguration{
				Enabled: true,
				Message: EmailMessageConfiguration{
					"sender":   `"Welcome Email Sender" <welcomeemailsender@example.com>`,
					"subject":  "welcomeemailsubject",
					"reply_to": `"Welcome Email Reply To" <welcomeemailreplyto@example.com>`,
				},
				Destination: "first",
			},
			Identity: &IdentityConfiguration{
				LoginID: &LoginIDConfiguration{
					Keys: []LoginIDKeyConfiguration{
						LoginIDKeyConfiguration{
							Key:     "email",
							Type:    LoginIDKeyType("email"),
							Maximum: newInt(1),
						},
						LoginIDKeyConfiguration{
							Key:     "phone",
							Type:    LoginIDKeyType("phone"),
							Maximum: newInt(1),
						},
						LoginIDKeyConfiguration{
							Key:     "username",
							Type:    LoginIDKeyTypeRaw,
							Maximum: newInt(1),
						},
					},
					Types: &LoginIDTypesConfiguration{
						Email: &LoginIDTypeEmailConfiguration{
							CaseSensitive: newFalse(),
							BlockPlusSign: newFalse(),
							IgnoreDotSign: newFalse(),
						},
						Username: &LoginIDTypeUsernameConfiguration{
							BlockReservedUsernames: newTrue(),
							ExcludedKeywords:       []string{"skygear"},
							ASCIIOnly:              newFalse(),
							CaseSensitive:          newFalse(),
						},
					},
				},
				OAuth: &OAuthConfiguration{
					StateJWTSecret:                 "oauthstatejwtsecret",
					ExternalAccessTokenFlowEnabled: true,
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
			UserVerification: &UserVerificationConfiguration{
				AutoSendOnSignup: true,
				Criteria:         "any",
				LoginIDKeys: []UserVerificationKeyConfiguration{
					UserVerificationKeyConfiguration{
						Key:             "email",
						CodeFormat:      "complex",
						Expiry:          3600,
						SuccessRedirect: "http://localhost:3000/userverification/success",
						ErrorRedirect:   "http://localhost:3000/userverification/error",
						EmailMessage: EmailMessageConfiguration{
							"sender":   `"Verify Sender" <userverificationsender@example.com>`,
							"subject":  "userverificationsubject",
							"reply_to": `"Verify Reply To" <userverificationreplyto@example.com>`,
						},
						SMSMessage: SMSMessageConfiguration{
							"sender": "+85212345678",
						},
					},
				},
			},
			Hook: &HookAppConfiguration{
				Secret: "hook-secret",
			},
			Messages: &MessagesConfiguration{
				Email: EmailMessageConfiguration{
					"sender":   `"Default Sender" <defaultsender@example.com>`,
					"subject":  "subject",
					"reply_to": `"Default Reply To" <defaultreplyto@example.com>`,
				},
				SMS: SMSMessageConfiguration{
					"sender": "+85212345678",
				},
			},
			SMTP: &SMTPConfiguration{
				Host:     "localhost",
				Port:     465,
				Mode:     "ssl",
				Login:    "user",
				Password: "password",
			},
			Twilio: &TwilioConfiguration{
				AccountSID: "mytwilioaccountsid",
				AuthToken:  "mytwilioauthtoken",
			},
			Nexmo: &NexmoConfiguration{
				APIKey:    "mynexmoapikey",
				APISecret: "mynexmoapisecret",
			},
		},
		TemplateItems: []TemplateItem{
			TemplateItem{
				Type:        "templatetype",
				LanguageTag: "en",
				Key:         "templatekey",
				URI:         "file:///template.html",
				Digest:      "base64",
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

			So(c.APIVersion, ShouldEqual, apiversion.APIVersion)
			So(c.AppName, ShouldEqual, "myapp")
			So(c.DatabaseConfig.DatabaseURL, ShouldEqual, "postgres://")
			So(c.AppConfig.Clients, ShouldBeEmpty)
			So(c.AppConfig.MasterKey, ShouldEqual, "masterkey")
		})
		Convey("should have default value when load from YAML", func() {
			c, err := NewTenantConfigurationFromYAML(strings.NewReader(inputMinimalYAML))
			So(err, ShouldBeNil)
			So(c.AppConfig.SMTP.Port, ShouldEqual, 25)
		})
		// JSON
		Convey("should have default value when load from JSON", func() {
			c, err := NewTenantConfigurationFromJSON(strings.NewReader(inputMinimalJSON), false)
			So(err, ShouldBeNil)
			So(c.APIVersion, ShouldEqual, apiversion.APIVersion)
			So(c.AppName, ShouldEqual, "myapp")
			So(c.DatabaseConfig.DatabaseURL, ShouldEqual, "postgres://")
			So(c.AppConfig.Clients, ShouldBeEmpty)
			So(c.AppConfig.MasterKey, ShouldEqual, "masterkey")
			So(c.AppConfig.SMTP.Port, ShouldEqual, 25)
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
			c.AppConfig.Identity.OAuth.Providers = []OAuthProviderConfiguration{
				OAuthProviderConfiguration{
					Type:         OAuthProviderTypeGoogle,
					ClientID:     "googleclientid",
					ClientSecret: "googleclientsecret",
				},
			}
			c.AfterUnmarshal()

			google := c.AppConfig.Identity.OAuth.Providers[0]

			So(google.ID, ShouldEqual, OAuthProviderTypeGoogle)
			So(google.Scope, ShouldEqual, "openid profile email")
		})

		testValidation := func(c *TenantConfiguration, causes []validation.ErrorCause) {
			b, err := json.Marshal(c)
			So(err, ShouldBeNil)
			_, err = NewTenantConfigurationFromJSON(bytes.NewReader(b), false)
			So(err, ShouldNotBeNil)
			So(validation.ErrorCauses(err), ShouldResemble, causes)
		}

		Convey("should validate client_id != master key", func() {
			c := makeFullTenantConfig()
			for i := range c.AppConfig.Clients {
				if c.AppConfig.Clients[i].ClientID() == "web_app" {
					c.AppConfig.Clients[i]["client_id"] = c.AppConfig.MasterKey
				}
			}

			testValidation(&c, []validation.ErrorCause{{
				Kind:    validation.ErrorGeneral,
				Message: "master key must not be same as client_id",
				Pointer: "/user_config/master_key",
			}})
		})
		Convey("UserVerification.LoginIDKeys is subset of Auth.LoginIDKeys", func() {
			c := makeFullTenantConfig()
			c.AppConfig.UserVerification.LoginIDKeys = append(
				c.AppConfig.UserVerification.LoginIDKeys,
				UserVerificationKeyConfiguration{
					Key:          "invalid",
					SMSMessage:   SMSMessageConfiguration{},
					EmailMessage: EmailMessageConfiguration{},
				},
			)

			testValidation(&c, []validation.ErrorCause{{
				Kind:    validation.ErrorGeneral,
				Message: "cannot verify disallowed login ID key",
				Pointer: "/user_config/user_verification/login_id_keys/invalid",
			}})
		})
		Convey("should validate OAuth Provider", func() {
			c := makeFullTenantConfig()
			c.AppConfig.Identity.OAuth.Providers = []OAuthProviderConfiguration{
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

			testValidation(&c, []validation.ErrorCause{{
				Kind:    validation.ErrorGeneral,
				Message: "duplicated OAuth provider",
				Pointer: "/user_config/identity/oauth/providers/1",
			}})
		})
		Convey("validate default country calling code", func() {
			c := makeFullTenantConfig()
			c.AppConfig.AuthUI.CountryCallingCode.Values = []string{"852"}
			c.AppConfig.AuthUI.CountryCallingCode.Default = "1"
			testValidation(&c, []validation.ErrorCause{{
				Kind:    validation.ErrorGeneral,
				Message: "default country calling code is unlisted",
				Pointer: "/user_config/auth_ui/country_calling_code/default",
			}})
		})
		Convey("should omit empty", func() {
			config, _ := NewTenantConfigurationFromJSON(strings.NewReader(inputMinimalJSON), true)
			bodyBytes, _ := json.Marshal(config)
			So(string(bodyBytes), ShouldEqualJSON, inputMinimalJSON)
		})

		Convey("ShouldNotHaveDuplicatedTypeInSamePath", func() {
			pathSet := map[string]interface{}{}
			pass := marshal.ShouldNotHaveDuplicatedTypeInSamePath(&AppConfiguration{}, pathSet)
			So(pass, ShouldBeTrue)
		})

		Convey("UpdateNilFieldsWithZeroValue", func() {
			userConfig := &AppConfiguration{}
			So(userConfig.CORS, ShouldBeNil)
			So(userConfig.Authentication, ShouldBeNil)
			So(userConfig.Authenticator, ShouldBeNil)
			So(userConfig.ForgotPassword, ShouldBeNil)
			So(userConfig.WelcomeEmail, ShouldBeNil)
			So(userConfig.Identity, ShouldBeNil)
			So(userConfig.UserVerification, ShouldBeNil)
			So(userConfig.Hook, ShouldBeNil)
			So(userConfig.SMTP, ShouldBeNil)
			So(userConfig.Twilio, ShouldBeNil)
			So(userConfig.Nexmo, ShouldBeNil)
			So(userConfig.Asset, ShouldBeNil)

			marshal.UpdateNilFieldsWithZeroValue(userConfig)

			So(userConfig.CORS, ShouldNotBeNil)
			So(userConfig.Authentication, ShouldNotBeNil)
			So(userConfig.Authenticator, ShouldNotBeNil)
			So(userConfig.ForgotPassword, ShouldNotBeNil)
			So(userConfig.WelcomeEmail, ShouldNotBeNil)
			So(userConfig.Identity, ShouldNotBeNil)
			So(userConfig.UserVerification, ShouldNotBeNil)
			So(userConfig.Hook, ShouldNotBeNil)
			So(userConfig.SMTP, ShouldNotBeNil)
			So(userConfig.Twilio, ShouldNotBeNil)
			So(userConfig.Nexmo, ShouldNotBeNil)
			So(userConfig.Asset, ShouldNotBeNil)
			So(userConfig.Messages, ShouldNotBeNil)
			So(userConfig.Messages.Email, ShouldNotBeNil)
		})
	})

}
