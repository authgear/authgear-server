package sms

import (
	"encoding/json"
	"errors"
	"io"
	"os"
	"strings"
	"testing"

	"github.com/kelseyhightower/envconfig"
	. "github.com/smartystreets/goconvey/convey"
	goyaml "go.yaml.in/yaml/v2"
	"sigs.k8s.io/yaml"

	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/lib/infra/sms/custom"
)

// nolint: gocognit
func TestClientResolver(t *testing.T) {
	Convey("resolve client", t, func() {
		f, err := os.Open("testdata/client_resolver_tests.yaml")
		if err != nil {
			panic(err)
		}
		defer f.Close()

		type AuthgearYAML struct {
			Messaging *config.MessagingConfig `json:"messaging"`
		}

		type AuthgesrSecretsYAML struct {
			Nexmo  *config.NexmoCredentials        `json:"nexmo"`
			Twilio *config.TwilioCredentials       `json:"twilio"`
			Custom *config.CustomSMSProviderConfig `json:"custom"`
		}

		type EnvConfig struct {
			SMSGateway config.SMSGatewayEnvironmentConfig `envconfig:"SMS_GATEWAY"`
		}

		type TestCase struct {
			Name                 string  `yaml:"name"`
			AuthgearYAML         any     `yaml:"authgear.yaml"`
			AuthgearSecretsYAML  any     `yaml:"authgear.secrets.yaml"`
			EnvironmentVariables *string `yaml:"environment_variables"`
			Result               any     `yaml:"result"`
			Error                string  `yaml:"error"`
		}

		decoder := goyaml.NewDecoder(f)
		for {
			var testCase TestCase
			err := decoder.Decode(&testCase)
			if errors.Is(err, io.EOF) {
				break
			} else if err != nil {
				panic(err)
			}

			Convey(testCase.Name, func() {
				authgearYAMLData, err := goyaml.Marshal(testCase.AuthgearYAML)
				if err != nil {
					panic(err)
				}
				authgearYAMLData, err = yaml.YAMLToJSON(authgearYAMLData)
				if err != nil {
					panic(err)
				}
				var authgearYAML *AuthgearYAML
				err = json.Unmarshal(authgearYAMLData, &authgearYAML)
				if err != nil {
					panic(err)
				}

				authgearSecretsYAMLData, err := goyaml.Marshal(testCase.AuthgearSecretsYAML)
				if err != nil {
					panic(err)
				}
				authgearSecretsYAMLData, err = yaml.YAMLToJSON(authgearSecretsYAMLData)
				if err != nil {
					panic(err)
				}
				var authgearSecretsYAML *AuthgesrSecretsYAML
				err = json.Unmarshal(authgearSecretsYAMLData, &authgearSecretsYAML)
				if err != nil {
					panic(err)
				}

				var messagingConfig *config.MessagingConfig
				if authgearYAML != nil {
					messagingConfig = authgearYAML.Messaging
				}
				var authgearSecretsYAMLNexmo *config.NexmoCredentials
				var authgearSecretsYAMLTwilio *config.TwilioCredentials
				var authgearSecretsYAMLCustom *config.CustomSMSProviderConfig
				if authgearSecretsYAML != nil {
					authgearSecretsYAMLNexmo = authgearSecretsYAML.Nexmo
					authgearSecretsYAMLTwilio = authgearSecretsYAML.Twilio
					authgearSecretsYAMLCustom = authgearSecretsYAML.Custom
				}

				var smsGatewayEnvironmentConfig config.SMSGatewayEnvironmentConfig
				if testCase.EnvironmentVariables != nil {
					for ln := range strings.SplitSeq(*testCase.EnvironmentVariables, "\n") {
						var keyval = strings.Split(ln, "=")
						if len(keyval) < 2 {
							continue
						}
						t.Setenv(keyval[0], keyval[1])
					}

					cfg := &EnvConfig{}
					err = envconfig.Process("", cfg)
					if err != nil {
						panic(err)
					}

					smsGatewayEnvironmentConfig = cfg.SMSGateway
				}

				resultData, err := goyaml.Marshal(testCase.Result)
				if err != nil {
					panic(err)
				}
				resultData, err = yaml.YAMLToJSON(resultData)
				if err != nil {
					panic(err)
				}
				var result any
				err = json.Unmarshal(resultData, &result)
				if err != nil {
					panic(err)
				}

				var authgearYAMLSMSProvider config.SMSProvider
				var authgearYAMLSMSGateway *config.SMSGatewayConfig
				if messagingConfig != nil {
					authgearYAMLSMSProvider = messagingConfig.Deprecated_SMSProvider
					authgearYAMLSMSGateway = messagingConfig.SMSGateway
				}

				var environmentDefaultProvider config.SMSGatewayEnvironmentDefaultProvider
				var environmentDefaultUseConfigFrom config.SMSGatewayEnvironmentDefaultUseConfigFrom
				var environmentNexmoCredentials config.SMSGatewayEnvironmentNexmoCredentials
				var environmentTwilioCredentials config.SMSGatewayEnvironmentTwilioCredentials
				var environmentCustomSMSProviderConfig config.SMSGatewayEnvironmentCustomSMSProviderConfig
				environmentDefaultProvider = smsGatewayEnvironmentConfig.Default.Provider
				environmentDefaultUseConfigFrom = smsGatewayEnvironmentConfig.Default.UseConfigFrom
				environmentNexmoCredentials = smsGatewayEnvironmentConfig.Nexmo
				environmentTwilioCredentials = smsGatewayEnvironmentConfig.Twilio
				environmentCustomSMSProviderConfig = smsGatewayEnvironmentConfig.Custom

				clientResolver := ClientResolver{
					AuthgearYAMLSMSProvider:                    authgearYAMLSMSProvider,
					AuthgearYAMLSMSGateway:                     authgearYAMLSMSGateway,
					AuthgearSecretsYAMLNexmoCredentials:        authgearSecretsYAMLNexmo,
					AuthgearSecretsYAMLTwilioCredentials:       authgearSecretsYAMLTwilio,
					AuthgearSecretsYAMLCustomSMSProviderConfig: authgearSecretsYAMLCustom,
					EnvironmentDefaultProvider:                 environmentDefaultProvider,
					EnvironmentDefaultUseConfigFrom:            environmentDefaultUseConfigFrom,
					EnvironmentNexmoCredentials:                environmentNexmoCredentials,
					EnvironmentTwilioCredentials:               environmentTwilioCredentials,
					EnvironmentCustomSMSProviderConfig:         environmentCustomSMSProviderConfig,
				}
				_, cred, err := clientResolver.ResolveClient()
				if testCase.Error != "" {
					So(err.Error(), ShouldEqual, testCase.Error)
				} else {
					So(err, ShouldBeNil)
				}
				if result != nil {
					So(toMap(cred), ShouldEqual, result)
				} else {
					So(cred, ShouldBeNil)
				}

			})
		}
	})
}

// TestClientResolverWebHookSelection pins which webhook a resolved custom
// client carries, because that is what decides whether the fetch address
// policy applies. See docs/specs/ssrf-protection.md.
func TestClientResolverWebHookSelection(t *testing.T) {
	Convey("the webhook depends on where the URL came from", t, func() {
		envCustom := config.SMSGatewayEnvironmentCustomSMSProviderConfig{
			URL:     "http://authgear-sms-gateway.default.svc.cluster.local:8080/send",
			Timeout: "30",
		}
		secretsCustom := &config.CustomSMSProviderConfig{
			URL: "https://sms.example.com/send",
		}

		resolveWebHook := func(r ClientResolver) custom.SMSWebHookCaller {
			client, _, err := r.ResolveClient()
			So(err, ShouldBeNil)
			customClient, ok := client.(*custom.CustomClient)
			So(ok, ShouldBeTrue)
			return customClient.SMSWebHook
		}

		Convey("SMS_GATEWAY_CUSTOM_URL is not bound by the address policy", func() {
			webhook := resolveWebHook(ClientResolver{
				EnvironmentDefaultUseConfigFrom:    config.SMSGatewayEnvironmentDefaultUseConfigFromEnvironmentVariable,
				EnvironmentDefaultProvider:         config.SMSGatewayEnvironmentDefaultProviderCustom,
				EnvironmentCustomSMSProviderConfig: envCustom,
			})

			So(webhook, ShouldHaveSameTypeAs, &custom.EnvSMSWebHook{})
		})

		Convey("a project selecting environment_variable is not bound by it either", func() {
			// The project only says "use what the operator configured"; it
			// never names the destination.
			webhook := resolveWebHook(ClientResolver{
				AuthgearYAMLSMSGateway: &config.SMSGatewayConfig{
					UseConfigFrom: config.SMSGatewayUseConfigFromEnvironmentVariable,
					Provider:      config.SMSProviderCustom,
				},
				EnvironmentCustomSMSProviderConfig: envCustom,
			})

			So(webhook, ShouldHaveSameTypeAs, &custom.EnvSMSWebHook{})
		})

		Convey("messaging.sms_provider is bound by the address policy", func() {
			webhook := resolveWebHook(ClientResolver{
				AuthgearYAMLSMSProvider:                    config.SMSProviderCustom,
				AuthgearSecretsYAMLCustomSMSProviderConfig: secretsCustom,
			})

			So(webhook, ShouldHaveSameTypeAs, &custom.SMSWebHook{})
		})

		Convey("a project selecting authgear.secrets.yaml is bound by it", func() {
			webhook := resolveWebHook(ClientResolver{
				AuthgearYAMLSMSGateway: &config.SMSGatewayConfig{
					UseConfigFrom: config.SMSGatewayUseConfigFromAuthgearSecretsYAML,
					Provider:      config.SMSProviderCustom,
				},
				AuthgearSecretsYAMLCustomSMSProviderConfig: secretsCustom,
			})

			So(webhook, ShouldHaveSameTypeAs, &custom.SMSWebHook{})
		})
	})
}

func toMap(c SMSClientCredentials) map[string]any {
	switch v := c.(type) {
	case *TwilioClientCredentials:
		return map[string]any{
			"account_sid":         v.AccountSID,
			"auth_token":          v.AuthToken,
			"message_service_sid": v.MessagingServiceSID,
		}
	case *NexmoClientCredentials:
		return map[string]any{
			"api_key":    v.APIKey,
			"api_secret": v.APISecret,
		}
	case *CustomClientCredentials:
		return map[string]any{
			"url":     v.URL,
			"timeout": float64(*v.Timeout),
		}
	}
	return nil
}
