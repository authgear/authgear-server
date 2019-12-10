package config

import (
	"bytes"
	"encoding/json"
	"reflect"
	"strings"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
	"gopkg.in/yaml.v2"

	. "github.com/skygeario/skygear-server/pkg/core/skytest"
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
		"database_schema": "app",
		"hook":{}
	},
	"user_config": {
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
`

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
				AppName:          "myapp",
				SecureMatch:      true,
				Sender:           "myforgotpasswordsender",
				Subject:          "myforgotpasswordsubject",
				ReplyTo:          "myforgotpasswordreplyto",
				ResetURLLifetime: 60,
				SuccessRedirect:  "http://localhost:3000/forgotpassword/success",
				ErrorRedirect:    "http://localhost:3000/forgotpassword/error",
			},
			WelcomeEmail: &WelcomeEmailConfiguration{
				Enabled:     true,
				Sender:      "welcomeemailsender",
				Subject:     "welcomeemailsubject",
				ReplyTo:     "welcomeemailreplyto",
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
						Sender:          "userverificationsender",
						ReplyTo:         "userverificationreplyto",
					},
				},
			},
			Hook: &HookUserConfiguration{
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

func copySet(input map[string]interface{}) map[string]interface{} {
	output := map[string]interface{}{}
	for k := range input {
		output[k] = input[k]
	}

	return output
}

func shouldNotHaveDuplicatedTypeInSamePath(i interface{}, pathSet map[string]interface{}) bool {
	t := reflect.TypeOf(i).Elem()
	v := reflect.ValueOf(i).Elem()

	if t.Kind() != reflect.Struct {
		return true
	}
	numField := t.NumField()
	for i := 0; i < numField; i++ {
		zerovalueTag := t.Field(i).Tag.Get("default_zero_value")
		if zerovalueTag != "true" {
			continue
		}

		field := v.Field(i)
		ft := t.Field(i)
		if field.Kind() == reflect.Ptr {
			ele := field.Elem()
			if !ele.IsValid() {
				ele = reflect.New(ft.Type.Elem())
				field.Set(ele)
			}
			typeName := ft.Type.String()
			if _, ok := pathSet[typeName]; ok {
				return false
			}
			newSet := copySet(pathSet)
			newSet[ft.Type.String()] = struct{}{}
			pass := shouldNotHaveDuplicatedTypeInSamePath(field.Interface(), newSet)
			if !pass {
				return false
			}
		}
	}

	return true
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
		Convey("should omit empty", func() {
			config, _ := NewTenantConfigurationFromJSON(strings.NewReader(inputMinimalJSON), true)
			bodyBytes, _ := json.Marshal(config)
			So(string(bodyBytes), ShouldEqualJSON, inputMinimalJSON)
		})
	})

	Convey("Test updateNilFieldsWithZeroValue", t, func() {
		Convey("should update nil fields with tag", func() {
			type ChildStruct struct {
				Num1 *int
				Num2 *int `default_zero_value:"true"`
			}

			type TestStruct struct {
				ChildNode1 *ChildStruct `default_zero_value:"true"`
				ChildNode2 *ChildStruct
			}

			s := &TestStruct{}
			updateNilFieldsWithZeroValue(s)

			So(s.ChildNode1, ShouldNotBeNil)
			So(s.ChildNode2, ShouldBeNil)

			So(s.ChildNode1.Num1, ShouldBeNil)
			So(s.ChildNode1.Num2, ShouldNotBeNil)
		})

		Convey("should update nil fields in user config", func() {
			userConfig := &UserConfiguration{}
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

			updateNilFieldsWithZeroValue(userConfig)

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

	Convey("Test shouldNotHaveDuplicatedTypeInSamePath", t, func() {
		Convey("should pass for normal struct", func() {
			type SubConfigItem struct {
				Num1 *int `default_zero_value:"true"`
			}

			type ConfigItem struct {
				SubItem *SubConfigItem `default_zero_value:"true"`
			}

			type RootConfig struct {
				Item *ConfigItem `default_zero_value:"true"`
			}

			pathSet := map[string]interface{}{}
			pass := shouldNotHaveDuplicatedTypeInSamePath(&RootConfig{}, pathSet)
			So(pass, ShouldBeTrue)
		})

		Convey("should fail for struct with self reference", func() {
			type ConfigItem struct {
				SubItem *ConfigItem `default_zero_value:"true"`
			}

			type RootConfig struct {
				Item *ConfigItem `default_zero_value:"true"`
			}

			pathSet := map[string]interface{}{}
			pass := shouldNotHaveDuplicatedTypeInSamePath(&RootConfig{}, pathSet)
			So(pass, ShouldBeFalse)
		})

		Convey("should pass for user config", func() {
			pathSet := map[string]interface{}{}
			pass := shouldNotHaveDuplicatedTypeInSamePath(&UserConfiguration{}, pathSet)
			So(pass, ShouldBeTrue)
		})

	})
}
