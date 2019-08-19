package config

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
	"gopkg.in/yaml.v2"
)

const inputMinimalYAML = `version: '1'
app_name: myapp
app_config:
  database_url: postgres://
user_config:
  api_key: apikey
  master_key: masterkey
  auth:
    login_id_keys:
      email:
        type: email
      phone:
        type: phone
      username:
        type: raw
  token_store:
    secret: tokensecret
  hook:
    secret: hooksecret
`

const inputMinimalJSON = `
{
	"version": "1",
	"app_name": "myapp",
	"app_config": {
		"database_url": "postgres://"
	},
	"user_config": {
		"api_key": "apikey",
		"master_key": "masterkey",
		"auth": {
			"login_id_keys": {
				"email": {
					"type": "email"
				},
				"phone": {
					"type": "phone"
				},
				"username": {
					"type": "raw"
				}
			},
			"allowed_realms": ["default"]
		},
		"token_store": {
			"secret": "tokensecret"
		},
		"hook": {
			"secret": "hooksecret"
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
		AppConfig: AppConfiguration{
			DatabaseURL: "postgres://user:password@localhost:5432/db?sslmode=disable",
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
			Hook: HookAppConfiguration{
				SyncHookTimeout:      10,
				SyncHookTotalTimeout: 60,
			},
		},
		UserConfig: UserConfiguration{
			APIKey:    "myapikey",
			MasterKey: "mymasterkey",
			URLPrefix: "http://localhost:3000",
			CORS: CORSConfiguration{
				Origin: "localhost:3000",
			},
			Auth: AuthConfiguration{
				LoginIDKeys: map[string]LoginIDKeyConfiguration{
					"email": LoginIDKeyConfiguration{
						Type:    LoginIDKeyType("email"),
						Minimum: newInt(0),
						Maximum: newInt(1),
					},
					"phone": LoginIDKeyConfiguration{
						Type:    LoginIDKeyType("phone"),
						Minimum: newInt(0),
						Maximum: newInt(1),
					},
					"username": LoginIDKeyConfiguration{
						Type:    LoginIDKeyTypeRaw,
						Minimum: newInt(0),
						Maximum: newInt(1),
					},
				},
				AllowedRealms:              []string{"default"},
				OnUserDuplicateAllowCreate: true,
			},
			TokenStore: TokenStoreConfiguration{
				Secret: "mytokenstoresecret",
				Expiry: 0,
			},
			UserAudit: UserAuditConfiguration{
				Enabled:         true,
				TrailHandlerURL: "http://localhost:3000/useraudit",
				Password: PasswordConfiguration{
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
			ForgotPassword: ForgotPasswordConfiguration{
				AppName:             "myapp",
				URLPrefix:           "http://localhost:3000/forgotpassword",
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
				URLPrefix:   "http://localhost:3000/welcomeemail",
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
					URLPrefix:      "http://localhost:3000/oauth",
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
				URLPrefix:        "http://localhost:3000/userverification",
				AutoSendOnSignup: true,
				Criteria:         "any",
				ErrorRedirect:    "http://localhost:3000/userverification/error",
				ErrorHTMLURL:     "http://localhost:3000/userverification/error.html",
				LoginIDKeys: map[string]UserVerificationKeyConfiguration{
					"email": UserVerificationKeyConfiguration{

						CodeFormat:      "complex",
						Expiry:          3600,
						SuccessRedirect: "http://localhost:3000/userverification/success",
						SuccessHTMLURL:  "http://localhost:3000/userverification/success.html",
						ErrorRedirect:   "http://localhost:3000/userverification/error",
						ErrorHTMLURL:    "http://localhost:3000/userverification/error.html",
						Provider:        "twilio",
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
			So(c.UserConfig.APIKey, ShouldEqual, "apikey")
			So(c.UserConfig.MasterKey, ShouldEqual, "masterkey")
		})
		Convey("should have default value when load from YAML", func() {
			c, err := NewTenantConfigurationFromYAML(strings.NewReader(inputMinimalYAML))
			So(err, ShouldBeNil)
			So(c.UserConfig.CORS.Origin, ShouldEqual, "*")
			So(c.AppConfig.SMTP.Port, ShouldEqual, 25)
		})
		Convey("should validate when load from YAML", func() {
			invalidInput := `app_name: myapp
app_config:
  database_url: postgres://
user_config:
  api_key: apikey
  master_key: masterkey
`
			_, err := NewTenantConfigurationFromYAML(strings.NewReader(invalidInput))
			So(err, ShouldBeError, "Only version 1 is supported")
		})
		// JSON
		Convey("should have default value when load from JSON", func() {
			c, err := NewTenantConfigurationFromJSON(strings.NewReader(inputMinimalJSON), false)
			So(err, ShouldBeNil)
			So(c.Version, ShouldEqual, "1")
			So(c.AppName, ShouldEqual, "myapp")
			So(c.AppConfig.DatabaseURL, ShouldEqual, "postgres://")
			So(c.UserConfig.APIKey, ShouldEqual, "apikey")
			So(c.UserConfig.MasterKey, ShouldEqual, "masterkey")
			So(c.UserConfig.CORS.Origin, ShouldEqual, "*")
			So(c.AppConfig.SMTP.Port, ShouldEqual, 25)
		})
		Convey("should validate when load from JSON", func() {
			invalidInput := `
		{
		  "app_name": "myapp",
		  "app_config": {
		    "database_url": "postgres://"
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
			So(err, ShouldBeError, "Only version 1 is supported")
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
			c.UserConfig.APIKey = c.UserConfig.MasterKey

			err := c.Validate()
			So(err, ShouldBeError, "MASTER_KEY cannot be the same as API_KEY")
		})
		Convey("should validate minimum <= maximum", func() {
			c := makeFullTenantConfig()
			email := c.UserConfig.Auth.LoginIDKeys["email"]
			email.Minimum = newInt(2)
			email.Maximum = newInt(1)
			c.UserConfig.Auth.LoginIDKeys["email"] = email
			err := c.Validate()
			So(err, ShouldBeError, "Invalid LoginIDKeys amount range: email")
		})
		Convey("UserVerification.LoginIDKeys is subset of Auth.LoginIDKeys", func() {
			c := makeFullTenantConfig()
			invalid := c.UserConfig.UserVerification.LoginIDKeys["email"]
			c.UserConfig.UserVerification.LoginIDKeys["invalid"] = invalid
			err := c.Validate()
			So(err, ShouldBeError, "Cannot verify disallowed login ID key: invalid")
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

			So(c.Validate(), ShouldBeError, "Duplicate OAuth Provider: azure")
		})
	})
}
