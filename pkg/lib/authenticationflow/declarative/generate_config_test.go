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

func TestGenerateAuthenticationFlowConfig(t *testing.T) {
	Convey("GenerateAuthenticationFlowConfig", t, func() {
		ctx := context.Background()

		// Saving the generated flows unchanged as custom flows must produce a valid config,
		// except where expectedErr says otherwise.
		test := func(cfgYAML string, expectedErr string) {
			appConfig, err := config.Parse(ctx, []byte(cfgYAML))
			So(err, ShouldBeNil)

			flows := GenerateAuthenticationFlowConfig(appConfig)
			So(flows.SignupFlows, ShouldHaveLength, 1)
			So(flows.PromoteFlows, ShouldHaveLength, 1)
			So(flows.LoginFlows, ShouldHaveLength, 1)
			So(flows.SignupLoginFlows, ShouldHaveLength, 1)
			So(flows.ReauthFlows, ShouldHaveLength, 1)
			So(flows.AccountRecoveryFlows, ShouldHaveLength, 1)

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
			if expectedErr == "" {
				So(err, ShouldBeNil)
			} else {
				So(err, ShouldBeError, expectedErr)
			}
		}

		Convey("default settings", func() {
			test(`
id: test
http:
  public_origin: http://test
`, "")
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
`, "")
		})

		// Without an email or phone login ID there is no way to recover an account,
		// so the generated account recovery flow has no options.
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
`, `invalid configuration:
/authentication_flow/account_recovery_flows/0/steps/0: required
  map[actual:[type] expected:[one_of] missing:[one_of]]`)
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
`, "")
		})
	})
}
