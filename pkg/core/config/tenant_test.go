package config

import (
	"bytes"
	"encoding/json"
	"io"
	"os"
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
    "welcome_email": {
      "enabled": true
    }
  }
}
`

func newInt(i int) *int {
	return &i
}

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
			AllowedRealms: []string{"default"},
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
			SenderName:          "myforgotpasswordsendername",
			Sender:              "myforgotpasswordsender",
			Subject:             "myforgotpasswordsubject",
			ReplyToName:         "myforgotpasswordreplytoname",
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
			SenderName:  "welcomeemailsendername",
			Sender:      "welcomeemailsender",
			Subject:     "welcomeemailsubject",
			ReplyToName: "welcomeemailreplytoname",
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
				JSSDKCDNURL:    "http://localhost:3000/oauth/sdk.js",
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
						Subject:     "userverificationsubject",
						Sender:      "userverificationsender",
						SenderName:  "userverificationsendername",
						ReplyTo:     "userverificationreplyto",
						ReplyToName: "userverificationreplytoname",
						TextURL:     "http://localhost:3000/userverification/text",
						HTMLURL:     "http://localhost:3000/userverification/html",
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
		Convey("should use master key as token secret as fallback in YAML", func() {
			c, err := NewTenantConfigurationFromYAML(strings.NewReader(inputMinimalYAML))
			So(err, ShouldBeNil)
			So(c.UserConfig.MasterKey, ShouldEqual, "masterkey")
			So(c.UserConfig.TokenStore.Secret, ShouldEqual, c.UserConfig.MasterKey)
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
			c, err := NewTenantConfigurationFromJSON(strings.NewReader(inputMinimalJSON))
			So(err, ShouldBeNil)
			So(c.Version, ShouldEqual, "1")
			So(c.AppName, ShouldEqual, "myapp")
			So(c.AppConfig.DatabaseURL, ShouldEqual, "postgres://")
			So(c.UserConfig.APIKey, ShouldEqual, "apikey")
			So(c.UserConfig.MasterKey, ShouldEqual, "masterkey")
			So(c.UserConfig.WelcomeEmail.Enabled, ShouldEqual, true)
			So(c.UserConfig.CORS.Origin, ShouldEqual, "*")
			So(c.AppConfig.SMTP.Port, ShouldEqual, 25)
		})
		Convey("should use master key as token secret as fallback in JSON", func() {
			c, err := NewTenantConfigurationFromJSON(strings.NewReader(inputMinimalJSON))
			So(err, ShouldBeNil)
			So(err, ShouldBeNil)
			So(c.UserConfig.MasterKey, ShouldEqual, "masterkey")
			So(c.UserConfig.TokenStore.Secret, ShouldEqual, c.UserConfig.MasterKey)
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
			_, err := NewTenantConfigurationFromJSON(strings.NewReader(invalidInput))
			So(err, ShouldBeError, "Only version 1 is supported")
		})
		// Conversion
		Convey("should be losslessly converted between Go and msgpack", func() {
			c := &fullTenantConfig
			base64msgpack, err := c.StdBase64Msgpack()
			So(err, ShouldBeNil)
			cc, err := NewTenantConfigurationFromStdBase64Msgpack(base64msgpack)
			So(err, ShouldBeNil)
			So(c, ShouldResemble, cc)
		})
		Convey("should be losslessly converted between Go and JSON", func() {
			c := &fullTenantConfig
			b, err := json.Marshal(c)
			So(err, ShouldBeNil)
			cc, err := NewTenantConfigurationFromJSON(bytes.NewReader(b))
			So(err, ShouldBeNil)
			So(c, ShouldResemble, cc)
		})
		Convey("should be losslessly converted between Go and YAML", func() {
			c := &fullTenantConfig
			b, err := yaml.Marshal(c)
			So(err, ShouldBeNil)
			cc, err := NewTenantConfigurationFromYAML(bytes.NewReader(b))
			So(err, ShouldBeNil)
			So(c, ShouldResemble, cc)
		})
		// Env
		Convey("should load tenant config from env", func() {
			os.Clearenv()
			_, err := NewTenantConfigurationFromEnv()
			So(err, ShouldBeError, "DATABASE_URL is not set")

			os.Setenv("DATABASE_URL", "postgres://")
			os.Setenv("APP_NAME", "myapp")
			os.Setenv("API_KEY", "api_key")
			os.Setenv("MASTER_KEY", "master_key")
			c, err := NewTenantConfigurationFromEnv()

			So(err, ShouldBeNil)
			So(c.AppName, ShouldEqual, "myapp")
			So(c.AppConfig.DatabaseURL, ShouldEqual, "postgres://")
			So(c.UserConfig.APIKey, ShouldEqual, "api_key")
			So(c.UserConfig.MasterKey, ShouldEqual, "master_key")
		})
		Convey("should load tenant config from yaml and env", func() {
			os.Clearenv()

			os.Setenv("DATABASE_URL", "postgres://remote")
			os.Setenv("APP_NAME", "yourapp")
			os.Setenv("API_KEY", "your_api_key")
			os.Setenv("MASTER_KEY", "your_master_key")

			c, err := NewTenantConfigurationFromYAMLAndEnv(func() (io.Reader, error) {
				return strings.NewReader(inputMinimalYAML), nil
			})

			So(err, ShouldBeNil)
			So(c.AppName, ShouldEqual, "yourapp")
			So(c.AppConfig.DatabaseURL, ShouldEqual, "postgres://remote")
			So(c.UserConfig.APIKey, ShouldEqual, "your_api_key")
			So(c.UserConfig.MasterKey, ShouldEqual, "your_master_key")
		})
		Convey("should set OAuth provider id and default scope", func() {
			c := fullTenantConfig
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
		Convey("should validate Custom Token", func() {
			c := fullTenantConfig
			c.UserConfig.SSO.CustomToken.Enabled = true
			c.UserConfig.SSO.CustomToken.Issuer = ""
			c.UserConfig.SSO.CustomToken.Secret = ""

			err := c.Validate()
			So(err, ShouldBeError, "Must set Custom Token Issuer")

			c.UserConfig.SSO.CustomToken.Issuer = "customtokenissuer"
			err = c.Validate()
			So(err, ShouldBeError, "Must set Custom Token Secret")
		})
		Convey("should validate OAuth Provider", func() {
			c := fullTenantConfig
			c.UserConfig.SSO.OAuth.Providers = []OAuthProviderConfiguration{
				OAuthProviderConfiguration{
					Type: OAuthProviderTypeAzureADv2,
				},
				OAuthProviderConfiguration{
					Type: OAuthProviderTypeAzureADv2,
				},
			}

			So(c.Validate(), ShouldBeError, "Missing OAuth Provider ID")

			c.UserConfig.SSO.OAuth.Providers[0].ID = "azure"

			So(c.Validate(), ShouldBeError, "Must set Azure Tenant")

			c.UserConfig.SSO.OAuth.Providers[0].Tenant = "mytenant"

			So(c.Validate(), ShouldBeError, "OAuth Provider azure: missing client id")
			c.UserConfig.SSO.OAuth.Providers[0].ClientID = "clientid"

			So(c.Validate(), ShouldBeError, "OAuth Provider azure: missing client secret")

			c.UserConfig.SSO.OAuth.Providers[0].ClientSecret = "clientsecret"

			So(c.Validate(), ShouldBeError, "OAuth Provider azure: missing scope")

			c.UserConfig.SSO.OAuth.Providers[0].Scope = "profile email"
			c.UserConfig.SSO.OAuth.Providers[1] = c.UserConfig.SSO.OAuth.Providers[0]

			So(c.Validate(), ShouldBeError, "Duplicate OAuth Provider: azure")

			c.UserConfig.SSO.OAuth.Providers[1].ID = "azure1"

			So(c.Validate(), ShouldBeNil)

			c.UserConfig.SSO.OAuth.AllowedCallbackURLs = nil

			So(c.Validate(), ShouldBeError, "Must specify OAuth callback URLs")
		})
	})
}
