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
			]
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
			CORS: &CORSConfiguration{
				Origin: "localhost:3000",
			},
			Asset: &AssetConfiguration{
				Secret: "assetsecret",
			},
			Auth: &AuthConfiguration{
				AuthenticationSession: &AuthenticationSessionConfiguration{
					Secret: "authnsessionsecret",
				},
				LoginIDKeys: []LoginIDKeyConfiguration{
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
				LoginIDTypes: &LoginIDTypesConfiguration{
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
				AllowedRealms:              []string{"default"},
				OnUserDuplicateAllowCreate: true,
			},
			MFA: &MFAConfiguration{
				Enabled:     true,
				Enforcement: MFAEnforcementOptional,
				Maximum:     newInt(99),
				TOTP: &MFATOTPConfiguration{
					Maximum: newInt(99),
				},
				OOB: &MFAOOBConfiguration{
					Sender:  `"MFA Sender" <mfaoobsender@example.com>`,
					Subject: "mfaoobsubject",
					ReplyTo: `"MFA Reply To" <mfaoobreplyto@example.com>`,
					SMS: &MFAOOBSMSConfiguration{
						Maximum: newInt(99),
					},
					Email: &MFAOOBEmailConfiguration{
						Maximum: newInt(99),
					},
				},
				BearerToken: &MFABearerTokenConfiguration{
					ExpireInDays: 60,
				},
				RecoveryCode: &MFARecoveryCodeConfiguration{
					Count:       24,
					ListEnabled: true,
				},
			},
			UserAudit: &UserAuditConfiguration{
				Enabled:         true,
				TrailHandlerURL: "http://localhost:3000/useraudit",
			},
			PasswordPolicy: &PasswordPolicyConfiguration{
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
			ForgotPassword: &ForgotPasswordConfiguration{
				SecureMatch:      true,
				Sender:           `"Forgot Password Sender" <myforgotpasswordsender@example.com>`,
				Subject:          "myforgotpasswordsubject",
				ReplyTo:          `"Forgot Password Reply To" <myforgotpasswordreplyto@example.com>`,
				ResetURLLifetime: 60,
				SuccessRedirect:  "http://localhost:3000/forgotpassword/success",
				ErrorRedirect:    "http://localhost:3000/forgotpassword/error",
			},
			WelcomeEmail: &WelcomeEmailConfiguration{
				Enabled:     true,
				Sender:      `"Welcome Email Sender" <welcomeemailsender@example.com>`,
				Subject:     "welcomeemailsubject",
				ReplyTo:     `"Welcome Email Reply To" <welcomeemailreplyto@example.com>`,
				Destination: "first",
			},
			SSO: &SSOConfiguration{
				CustomToken: &CustomTokenConfiguration{
					Enabled:                    true,
					Issuer:                     "customtokenissuer",
					Audience:                   "customtokenaudience",
					Secret:                     "customtokensecret",
					OnUserDuplicateAllowMerge:  true,
					OnUserDuplicateAllowCreate: true,
				},
				OAuth: &OAuthConfiguration{
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
						Subject:         "userverificationsubject",
						Sender:          `"Verify Sender" <userverificationsender@example.com>`,
						ReplyTo:         `"Verify Reply To" <userverificationreplyto@example.com>`,
					},
				},
			},
			Hook: &HookAppConfiguration{
				Secret: "hook-secret",
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
				From:       "mytwilio",
			},
			Nexmo: &NexmoConfiguration{
				APIKey:    "mynexmoapikey",
				APISecret: "mynexmoapisecret",
				From:      "mynexmo",
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
			c.AppConfig.SSO.OAuth.Providers = []OAuthProviderConfiguration{
				OAuthProviderConfiguration{
					Type:         OAuthProviderTypeGoogle,
					ClientID:     "googleclientid",
					ClientSecret: "googleclientsecret",
				},
			}
			c.AfterUnmarshal()

			google := c.AppConfig.SSO.OAuth.Providers[0]

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

		Convey("should validate api key != master key", func() {
			c := makeFullTenantConfig()
			for i := range c.AppConfig.Clients {
				if c.AppConfig.Clients[i].ID == "web-app" {
					c.AppConfig.Clients[i].APIKey = c.AppConfig.MasterKey
				}
			}

			testValidation(&c, []validation.ErrorCause{{
				Kind:    validation.ErrorGeneral,
				Message: "master key must not be same as API key",
				Pointer: "/user_config/master_key",
			}})
		})
		Convey("UserVerification.LoginIDKeys is subset of Auth.LoginIDKeys", func() {
			c := makeFullTenantConfig()
			c.AppConfig.UserVerification.LoginIDKeys = append(
				c.AppConfig.UserVerification.LoginIDKeys,
				UserVerificationKeyConfiguration{
					Key: "invalid",
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
			c.AppConfig.SSO.OAuth.Providers = []OAuthProviderConfiguration{
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
				Pointer: "/user_config/sso/oauth/providers/1",
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
			So(userConfig.Auth, ShouldBeNil)
			So(userConfig.MFA, ShouldBeNil)
			So(userConfig.UserAudit, ShouldBeNil)
			So(userConfig.PasswordPolicy, ShouldBeNil)
			So(userConfig.ForgotPassword, ShouldBeNil)
			So(userConfig.WelcomeEmail, ShouldBeNil)
			So(userConfig.SSO, ShouldBeNil)
			So(userConfig.UserVerification, ShouldBeNil)
			So(userConfig.Hook, ShouldBeNil)
			So(userConfig.SMTP, ShouldBeNil)
			So(userConfig.Twilio, ShouldBeNil)
			So(userConfig.Nexmo, ShouldBeNil)
			So(userConfig.Asset, ShouldBeNil)

			marshal.UpdateNilFieldsWithZeroValue(userConfig)

			So(userConfig.CORS, ShouldNotBeNil)
			So(userConfig.Auth, ShouldNotBeNil)
			So(userConfig.MFA, ShouldNotBeNil)
			So(userConfig.UserAudit, ShouldNotBeNil)
			So(userConfig.PasswordPolicy, ShouldNotBeNil)
			So(userConfig.ForgotPassword, ShouldNotBeNil)
			So(userConfig.WelcomeEmail, ShouldNotBeNil)
			So(userConfig.SSO, ShouldNotBeNil)
			So(userConfig.UserVerification, ShouldNotBeNil)
			So(userConfig.Hook, ShouldNotBeNil)
			So(userConfig.SMTP, ShouldNotBeNil)
			So(userConfig.Twilio, ShouldNotBeNil)
			So(userConfig.Nexmo, ShouldNotBeNil)
			So(userConfig.Asset, ShouldNotBeNil)

			So(userConfig.Auth.AuthenticationSession, ShouldNotBeNil)
		})
	})

}
