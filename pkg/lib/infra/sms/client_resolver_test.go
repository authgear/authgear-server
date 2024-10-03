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
	goyaml "gopkg.in/yaml.v2"
	"sigs.k8s.io/yaml"

	"github.com/authgear/authgear-server/pkg/lib/config"
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
			Name                 string      `yaml:"name"`
			AuthgearYAML         interface{} `yaml:"authgear.yaml"`
			AuthgearSecretsYAML  interface{} `yaml:"authgear.secrets.yaml"`
			EnvironmentVariables *string     `yaml:"environment_variables"`
			Result               interface{} `yaml:"result"`
			Error                string      `yaml:"error"`
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
					for _, ln := range strings.Split(*testCase.EnvironmentVariables, "\n") {
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
				var result interface{}
				err = json.Unmarshal(resultData, &result)
				if err != nil {
					panic(err)
				}

				var authgearYAMLSMSProvider config.SMSProvider
				var authgearYAMLSMSGateway *config.SMSGatewayConfig
				if messagingConfig != nil {
					authgearYAMLSMSProvider = messagingConfig.SMSProvider
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
