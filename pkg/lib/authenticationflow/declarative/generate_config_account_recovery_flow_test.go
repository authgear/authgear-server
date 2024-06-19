package declarative

import (
	"bytes"
	"encoding/json"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
	"sigs.k8s.io/yaml"

	"github.com/authgear/authgear-server/pkg/lib/config"
)

func TestGenerateAccountRecoveryFlowConfig(t *testing.T) {
	Convey("GenerateAccountRecoveryFlowConfig", t, func() {
		test := func(cfgStr string, expected string) {

			jsonData, err := yaml.YAMLToJSON([]byte(cfgStr))
			So(err, ShouldBeNil)

			var appConfig config.AppConfig
			decoder := json.NewDecoder(bytes.NewReader(jsonData))
			err = decoder.Decode(&appConfig)
			So(err, ShouldBeNil)

			config.PopulateDefaultValues(&appConfig)

			flow := GenerateAccountRecoveryFlowConfig(&appConfig)
			flowJSON, err := json.Marshal(flow)
			So(err, ShouldBeNil)

			expectedJSON, err := yaml.YAMLToJSON([]byte(expected))
			So(err, ShouldBeNil)

			So(string(flowJSON), ShouldEqualJSON, string(expectedJSON))
		}

		// email, phone
		test(
			`
identity:
  login_id:
    keys:
    - type: email
    - type: phone
`,
			`
name: default
steps:
- type: identify
  one_of:
  - identification: email
    on_failure: ignore
    steps:
      - type: select_destination
        allowed_channels:
          - channel: email
            otp_form: link
  - identification: phone
    on_failure: ignore
    steps:
      - type: select_destination
        allowed_channels:
          - channel: sms
            otp_form: code
- type: verify_account_recovery_code
- type: reset_password
`)
		// phone
		test(
			`
identity:
  login_id:
    keys:
      - type: phone
`,
			`
name: default
steps:
- type: identify
  one_of:
  - identification: phone
    on_failure: ignore
    steps:
      - type: select_destination
        allowed_channels:
          - channel: sms
            otp_form: code
- type: verify_account_recovery_code
- type: reset_password
`)
		// email
		test(
			`
identity:
  login_id:
    keys:
      - type: email
`,
			`
name: default
steps:
- type: identify
  one_of:
  - identification: email
    on_failure: ignore
    steps:
      - type: select_destination
        allowed_channels:
          - channel: email
            otp_form: link
- type: verify_account_recovery_code
- type: reset_password
`)
		// email, phone, custom ui
		test(
			`
identity:
  login_id:
    keys:
    - type: email
    - type: phone
ui:
  forgot_password:
    phone:
      - channel: sms
        otp_form: link
    email:
      - channel: email
        otp_form: code
`,
			`
name: default
steps:
- type: identify
  one_of:
  - identification: email
    on_failure: ignore
    steps:
      - type: select_destination
        allowed_channels:
          - channel: email
            otp_form: code
  - identification: phone
    on_failure: ignore
    steps:
      - type: select_destination
        allowed_channels:
          - channel: sms
            otp_form: link
- type: verify_account_recovery_code
- type: reset_password
`)
		// captcha, 1 branch
		test(`
identity:
  login_id:
    keys:
      - type: phone
captcha:
  enabled: true
  providers:
  - type: recaptchav2
    alias: recaptchav2-a
    site_key: some-site-key
`,
			`
name: default
steps:
- type: identify
  one_of:
  - identification: phone
    captcha:
      mode: always
      provider: 
        alias: recaptchav2-a
    on_failure: ignore
    steps:
      - type: select_destination
        allowed_channels:
          - channel: sms
            otp_form: code
- type: verify_account_recovery_code
- type: reset_password
`)
		// captcha, 2 branches
		test(
			`
identity:
  login_id:
    keys:
    - type: email
    - type: phone
captcha:
  enabled: true
  providers:
  - type: recaptchav2
    alias: recaptchav2-a
    site_key: some-site-key
`,
			`
name: default
steps:
- type: identify
  one_of:
  - identification: email
    captcha:
      mode: always
      provider: 
        alias: recaptchav2-a
    on_failure: ignore
    steps:
      - type: select_destination
        allowed_channels:
          - channel: email
            otp_form: link
  - identification: phone
    captcha:
      mode: always
      provider: 
        alias: recaptchav2-a
    on_failure: ignore
    steps:
      - type: select_destination
        allowed_channels:
          - channel: sms
            otp_form: code
- type: verify_account_recovery_code
- type: reset_password
`)
	})
}
