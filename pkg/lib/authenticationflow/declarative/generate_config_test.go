package declarative

import (
	"bytes"
	"encoding/json"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
	"sigs.k8s.io/yaml"

	"github.com/authgear/authgear-server/pkg/lib/config"
)

func TestGenerateSignupFlowConfig(t *testing.T) {
	Convey("GenerateSignupFlowConfig", t, func() {
		test := func(cfgStr string, expected string) {
			jsonData, err := yaml.YAMLToJSON([]byte(cfgStr))
			So(err, ShouldBeNil)

			var appConfig config.AppConfig
			decoder := json.NewDecoder(bytes.NewReader(jsonData))
			err = decoder.Decode(&appConfig)
			So(err, ShouldBeNil)

			config.PopulateDefaultValues(&appConfig)

			signupFlow := GenerateSignupFlowConfig(&appConfig)
			signupFlowJSON, err := json.Marshal(signupFlow)
			So(err, ShouldBeNil)

			expectedJSON, err := yaml.YAMLToJSON([]byte(expected))
			So(err, ShouldBeNil)

			So(string(signupFlowJSON), ShouldEqualJSON, string(expectedJSON))
		}

		// email, password
		test(`
authentication:
  identities:
  - login_id
  primary_authenticators:
  - password
identity:
  login_id:
    keys:
    - type: email
`, `
id: default
steps:
- id: identify
  type: identify
  one_of:
  - identification: email
    steps:
    - target_step: identify
      type: verify
    - id: authenticate_primary_email
      type: authenticate
      one_of:
      - authentication: primary_password
`)

		// email, otp
		test(`
authentication:
  identities:
  - login_id
  primary_authenticators:
  - oob_otp_email
identity:
  login_id:
    keys:
    - type: email
`, `
id: default
steps:
- id: identify
  type: identify
  one_of:
  - identification: email
    steps:
    - target_step: identify
      type: verify
    - id: authenticate_primary_email
      type: authenticate
      one_of:
      - authentication: primary_oob_otp_email
        target_step: identify
        steps:
        - target_step: authenticate_primary_email
          type: verify
`)

		// phone, otp
		test(`
authentication:
  identities:
  - login_id
  primary_authenticators:
  - oob_otp_sms
identity:
  login_id:
    keys:
    - type: phone
`, `
id: default
steps:
- id: identify
  type: identify
  one_of:
  - identification: phone
    steps:
    - target_step: identify
      type: verify
    - id: authenticate_primary_phone
      type: authenticate
      one_of:
      - authentication: primary_oob_otp_sms
        target_step: identify
        steps:
        - target_step: authenticate_primary_phone
          type: verify
`)

		// username, password
		test(`
authentication:
  identities:
  - login_id
  primary_authenticators:
  - password
identity:
  login_id:
    keys:
    - type: username
`, `
id: default
steps:
- id: identify
  type: identify
  one_of:
  - identification: username
    steps:
    - id: authenticate_primary_username
      type: authenticate
      one_of:
      - authentication: primary_password
`)

		// email,phone, password,otp
		test(`
authentication:
  identities:
  - login_id
  primary_authenticators:
  - password
  - oob_otp_email
  - oob_otp_sms
identity:
  login_id:
    keys:
    - type: email
    - type: phone
`, `
id: default
steps:
- id: identify
  type: identify
  one_of:
  - identification: email
    steps:
    - target_step: identify
      type: verify
    - id: authenticate_primary_email
      type: authenticate
      one_of:
      - authentication: primary_password
      - authentication: primary_oob_otp_email
        target_step: identify
        steps:
        - target_step: authenticate_primary_email
          type: verify
  - identification: phone
    steps:
    - target_step: identify
      type: verify
    - id: authenticate_primary_phone
      type: authenticate
      one_of:
      - authentication: primary_password
      - authentication: primary_oob_otp_sms
        target_step: identify
        steps:
        - target_step: authenticate_primary_phone
          type: verify
`)

		// email,password, totp,recovery_code
		test(`
authentication:
  identities:
  - login_id
  primary_authenticators:
  - password
  secondary_authenticators:
  - totp
  secondary_authentication_mode: required
identity:
  login_id:
    keys:
    - type: email
`, `
id: default
steps:
- id: identify
  type: identify
  one_of:
  - identification: email
    steps:
    - target_step: identify
      type: verify
    - id: authenticate_primary_email
      type: authenticate
      one_of:
      - authentication: primary_password
    - id: authenticate_secondary_email
      type: authenticate
      one_of:
      - authentication: secondary_totp
        steps:
        - type: recovery_code
`)

		// oauth
		test(`
authentication:
  identities:
  - oauth
identity:
  oauth:
    providers:
    - alias: google
      type: google
`, `
id: default
steps:
- id: identify
  type: identify
  one_of:
  - identification: oauth
`)

		// oauth does not require 2fa.
		test(`
authentication:
  identities:
  - login_id
  - oauth
  primary_authenticators:
  - password
  secondary_authenticators:
  - totp
  secondary_authentication_mode: required
  device_token:
    disabled: true
  recovery_code:
    disabled: true
identity:
  login_id:
    keys:
    - type: email
  oauth:
    providers:
    - alias: google
      type: google
`, `
id: default
steps:
- id: identify
  type: identify
  one_of:
  - identification: email
    steps:
    - target_step: identify
      type: verify
    - id: authenticate_primary_email
      type: authenticate
      one_of:
      - authentication: primary_password
    - id: authenticate_secondary_email
      type: authenticate
      one_of:
      - authentication: secondary_totp
  - identification: oauth
`)

	})
}

func TestGenerateLoginFlowConfig(t *testing.T) {
	Convey("GenerateLoginFlowConfig", t, func() {
		test := func(cfgStr string, expected string) {
			jsonData, err := yaml.YAMLToJSON([]byte(cfgStr))
			So(err, ShouldBeNil)

			var appConfig config.AppConfig
			decoder := json.NewDecoder(bytes.NewReader(jsonData))
			err = decoder.Decode(&appConfig)
			So(err, ShouldBeNil)

			config.PopulateDefaultValues(&appConfig)

			signupFlow := GenerateLoginFlowConfig(&appConfig)
			signupFlowJSON, err := json.Marshal(signupFlow)
			So(err, ShouldBeNil)

			expectedJSON, err := yaml.YAMLToJSON([]byte(expected))
			So(err, ShouldBeNil)

			So(string(signupFlowJSON), ShouldEqualJSON, string(expectedJSON))
		}

		// email, password
		test(`
authentication:
  identities:
  - login_id
  primary_authenticators:
  - password
identity:
  login_id:
    keys:
    - type: email
`, `
id: default
steps:
- id: identify
  type: identify
  one_of:
  - identification: email
    steps:
    - id: authenticate_primary_email
      type: authenticate
      one_of:
      - authentication: primary_password
        steps:
        - type: change_password
          target_step: authenticate_primary_email
    - id: authenticate_secondary_email
      type: authenticate
      optional: true
      one_of:
      - authentication: device_token
      - authentication: recovery_code
      - authentication: secondary_totp
`)

		// email, otp
		test(`
authentication:
  identities:
  - login_id
  primary_authenticators:
  - oob_otp_email
identity:
  login_id:
    keys:
    - type: email
`, `
id: default
steps:
- id: identify
  type: identify
  one_of:
  - identification: email
    steps:
    - id: authenticate_primary_email
      type: authenticate
      one_of:
      - authentication: primary_oob_otp_email
        target_step: identify
    - id: authenticate_secondary_email
      type: authenticate
      optional: true
      one_of:
      - authentication: device_token
      - authentication: recovery_code
      - authentication: secondary_totp
`)

		// phone, otp
		test(`
authentication:
  identities:
  - login_id
  primary_authenticators:
  - oob_otp_sms
identity:
  login_id:
    keys:
    - type: phone
`, `
id: default
steps:
- id: identify
  type: identify
  one_of:
  - identification: phone
    steps:
    - id: authenticate_primary_phone
      type: authenticate
      one_of:
      - authentication: primary_oob_otp_sms
        target_step: identify
    - id: authenticate_secondary_phone
      type: authenticate
      optional: true
      one_of:
      - authentication: device_token
      - authentication: recovery_code
      - authentication: secondary_totp
`)

		// username, password
		test(`
authentication:
  identities:
  - login_id
  primary_authenticators:
  - password
identity:
  login_id:
    keys:
    - type: username
`, `
id: default
steps:
- id: identify
  type: identify
  one_of:
  - identification: username
    steps:
    - id: authenticate_primary_username
      type: authenticate
      one_of:
      - authentication: primary_password
        steps:
        - type: change_password
          target_step: authenticate_primary_username
    - id: authenticate_secondary_username
      type: authenticate
      optional: true
      one_of:
      - authentication: device_token
      - authentication: recovery_code
      - authentication: secondary_totp
`)

		// email,phone, password,otp
		test(`
authentication:
  identities:
  - login_id
  primary_authenticators:
  - password
  - oob_otp_email
  - oob_otp_sms
identity:
  login_id:
    keys:
    - type: email
    - type: phone
`, `
id: default
steps:
- id: identify
  type: identify
  one_of:
  - identification: email
    steps:
    - id: authenticate_primary_email
      type: authenticate
      one_of:
      - authentication: primary_password
        steps:
        - type: change_password
          target_step: authenticate_primary_email
      - authentication: primary_oob_otp_email
        target_step: identify
    - id: authenticate_secondary_email
      type: authenticate
      optional: true
      one_of:
      - authentication: device_token
      - authentication: recovery_code
      - authentication: secondary_totp
  - identification: phone
    steps:
    - id: authenticate_primary_phone
      type: authenticate
      one_of:
      - authentication: primary_password
        steps:
        - type: change_password
          target_step: authenticate_primary_phone
      - authentication: primary_oob_otp_sms
        target_step: identify
    - id: authenticate_secondary_phone
      type: authenticate
      optional: true
      one_of:
      - authentication: device_token
      - authentication: recovery_code
      - authentication: secondary_totp
`)

		// email,password, totp,recovery_code
		test(`
authentication:
  identities:
  - login_id
  primary_authenticators:
  - password
  secondary_authenticators:
  - totp
  secondary_authentication_mode: required
identity:
  login_id:
    keys:
    - type: email
`, `
id: default
steps:
- id: identify
  type: identify
  one_of:
  - identification: email
    steps:
    - id: authenticate_primary_email
      type: authenticate
      one_of:
      - authentication: primary_password
        steps:
        - type: change_password
          target_step: authenticate_primary_email
    - id: authenticate_secondary_email
      type: authenticate
      one_of:
      - authentication: device_token
      - authentication: recovery_code
      - authentication: secondary_totp
`)

		// Disable device token recovery code.
		test(`
authentication:
  identities:
  - login_id
  primary_authenticators:
  - password
  device_token:
    disabled: true
  recovery_code:
    disabled: true
identity:
  login_id:
    keys:
    - type: email
`, `
id: default
steps:
- id: identify
  type: identify
  one_of:
  - identification: email
    steps:
    - id: authenticate_primary_email
      type: authenticate
      one_of:
      - authentication: primary_password
        steps:
        - type: change_password
          target_step: authenticate_primary_email
    - id: authenticate_secondary_email
      type: authenticate
      optional: true
      one_of:
      - authentication: secondary_totp
`)

		// No password force change
		test(`
authentication:
  identities:
  - login_id
  primary_authenticators:
  - password
identity:
  login_id:
    keys:
    - type: email
authenticator:
  password:
    force_change: false
`, `
id: default
steps:
- id: identify
  type: identify
  one_of:
  - identification: email
    steps:
    - id: authenticate_primary_email
      type: authenticate
      one_of:
      - authentication: primary_password
    - id: authenticate_secondary_email
      type: authenticate
      optional: true
      one_of:
      - authentication: device_token
      - authentication: recovery_code
      - authentication: secondary_totp
`)

		// oauth
		test(`
authentication:
  identities:
  - oauth
identity:
  oauth:
    providers:
    - alias: google
      type: google
`, `
id: default
steps:
- id: identify
  type: identify
  one_of:
  - identification: oauth
`)

		// oauth does not require 2fa.
		test(`
authentication:
  identities:
  - login_id
  - oauth
  primary_authenticators:
  - password
  secondary_authenticators:
  - totp
  secondary_authentication_mode: required
  device_token:
    disabled: true
  recovery_code:
    disabled: true
identity:
  login_id:
    keys:
    - type: email
  oauth:
    providers:
    - alias: google
      type: google
`, `
id: default
steps:
- id: identify
  type: identify
  one_of:
  - identification: email
    steps:
    - id: authenticate_primary_email
      type: authenticate
      one_of:
      - authentication: primary_password
        steps:
        - type: change_password
          target_step: authenticate_primary_email
    - id: authenticate_secondary_email
      type: authenticate
      one_of:
      - authentication: secondary_totp
  - identification: oauth
`)
	})
}
