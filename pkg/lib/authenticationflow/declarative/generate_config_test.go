package declarative

import (
	"context"
	"encoding/json"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
	"sigs.k8s.io/yaml"

	"github.com/authgear/authgear-server/pkg/lib/config"
	_ "github.com/authgear/authgear-server/pkg/lib/oauthrelyingparty/google"
)

func TestGenerateConfigAsCustomFlows(t *testing.T) {
	Convey("Generated flows saved as custom flows", t, func() {
		ctx := context.Background()

		// Saving the generated flows unchanged as custom flows must produce a valid config.
		test := func(cfgYAML string, expectRecoverable bool) {
			appConfig, err := config.Parse(ctx, []byte(cfgYAML))
			So(err, ShouldBeNil)

			flows := &config.AuthenticationFlowConfig{
				SignupFlows:      []*config.AuthenticationFlowSignupFlow{GenerateSignupFlowConfig(appConfig)},
				LoginFlows:       []*config.AuthenticationFlowLoginFlow{GenerateLoginFlowConfig(appConfig)},
				SignupLoginFlows: []*config.AuthenticationFlowSignupLoginFlow{GenerateSignupLoginFlowConfig(appConfig)},
				ReauthFlows:      []*config.AuthenticationFlowReauthFlow{GenerateReauthFlowConfig(appConfig)},
			}
			accountRecoveryFlow := GenerateAccountRecoveryFlowConfigIfRecoverable(appConfig)
			if expectRecoverable {
				So(accountRecoveryFlow, ShouldNotBeNil)
				flows.AccountRecoveryFlows = []*config.AuthenticationFlowAccountRecoveryFlow{accountRecoveryFlow}
			} else {
				So(accountRecoveryFlow, ShouldBeNil)
			}

			flowsJSON, err := json.Marshal(flows)
			So(err, ShouldBeNil)
			var flowsMap map[string]any
			err = json.Unmarshal(flowsJSON, &flowsMap)
			So(err, ShouldBeNil)

			var raw map[string]any
			err = yaml.Unmarshal([]byte(cfgYAML), &raw)
			So(err, ShouldBeNil)
			raw["authentication_flow"] = flowsMap

			roundTripped, err := yaml.Marshal(raw)
			So(err, ShouldBeNil)
			_, err = config.Parse(ctx, roundTripped)
			So(err, ShouldBeNil)
		}

		Convey("default settings", func() {
			test(`
id: test
http:
  public_origin: http://test
`, true)
		})

		Convey("phone with SMS OTP", func() {
			test(`
id: test
http:
  public_origin: http://test
authentication:
  identities:
  - login_id
  primary_authenticators:
  - oob_otp_sms
identity:
  login_id:
    keys:
    - type: phone
`, true)
		})

		// Without an email or phone login ID there is no way to recover an account,
		// so no account recovery flow is generated.
		Convey("username and password with required 2FA", func() {
			test(`
id: test
http:
  public_origin: http://test
authentication:
  identities:
  - login_id
  primary_authenticators:
  - password
  secondary_authenticators:
  - totp
  - oob_otp_email
  secondary_authentication_mode: required
identity:
  login_id:
    keys:
    - type: username
`, false)
		})

		Convey("email with passkey and an OAuth provider", func() {
			test(`
id: test
http:
  public_origin: http://test
authentication:
  identities:
  - login_id
  - oauth
  - passkey
  primary_authenticators:
  - password
  - passkey
identity:
  login_id:
    keys:
    - type: email
  oauth:
    providers:
    - alias: google
      type: google
      client_id: client-id
`, true)
		})
	})
}
