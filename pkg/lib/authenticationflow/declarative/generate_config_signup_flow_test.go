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

			flow := GenerateSignupFlowConfig(&appConfig)
			flowJSON, err := json.Marshal(flow)
			So(err, ShouldBeNil)

			expectedJSON, err := yaml.YAMLToJSON([]byte(expected))
			So(err, ShouldBeNil)

			So(string(flowJSON), ShouldEqualJSON, string(expectedJSON))
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
name: default
steps:
- name: identify
  type: identify
  one_of:
  - identification: email
    steps:
    - target_step: identify
      type: verify
    - name: authenticate_primary_email
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
name: default
steps:
- name: identify
  type: identify
  one_of:
  - identification: email
    steps:
    - target_step: identify
      type: verify
    - name: authenticate_primary_email
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
name: default
steps:
- name: identify
  type: identify
  one_of:
  - identification: phone
    steps:
    - target_step: identify
      type: verify
    - name: authenticate_primary_phone
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
name: default
steps:
- name: identify
  type: identify
  one_of:
  - identification: username
    steps:
    - name: authenticate_primary_username
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
name: default
steps:
- name: identify
  type: identify
  one_of:
  - identification: email
    steps:
    - target_step: identify
      type: verify
    - name: authenticate_primary_email
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
    - name: authenticate_primary_phone
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
name: default
steps:
- name: identify
  type: identify
  one_of:
  - identification: email
    steps:
    - target_step: identify
      type: verify
    - name: authenticate_primary_email
      type: authenticate
      one_of:
      - authentication: primary_password
    - name: authenticate_secondary_email
      type: authenticate
      one_of:
      - authentication: secondary_totp
        steps:
        - type: recovery_code
`)

		// email,password, phone
		test(`
authentication:
  identities:
  - login_id
  primary_authenticators:
  - password
  secondary_authenticators:
  - oob_otp_sms
  secondary_authentication_mode: required
  recovery_code:
    disabled: true
identity:
  login_id:
    keys:
    - type: email
`, `
name: default
steps:
- name: identify
  type: identify
  one_of:
  - identification: email
    steps:
    - target_step: identify
      type: verify
    - name: authenticate_primary_email
      type: authenticate
      one_of:
      - authentication: primary_password
    - name: authenticate_secondary_email
      type: authenticate
      one_of:
      - authentication: secondary_oob_otp_sms
        steps:
        - type: verify
          target_step: authenticate_secondary_email
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
name: default
steps:
- name: identify
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
name: default
steps:
- name: identify
  type: identify
  one_of:
  - identification: email
    steps:
    - target_step: identify
      type: verify
    - name: authenticate_primary_email
      type: authenticate
      one_of:
      - authentication: primary_password
    - name: authenticate_secondary_email
      type: authenticate
      one_of:
      - authentication: secondary_totp
  - identification: oauth
`)

	})
}
