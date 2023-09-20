package declarative

import (
	"bytes"
	"encoding/json"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
	"sigs.k8s.io/yaml"

	"github.com/authgear/authgear-server/pkg/lib/config"
)

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

			flow := GenerateLoginFlowConfig(&appConfig)
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
    - name: authenticate_primary_email
      type: authenticate
      one_of:
      - authentication: primary_password
        steps:
        - type: change_password
          target_step: authenticate_primary_email
        - name: authenticate_secondary_email
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
name: default
steps:
- name: identify
  type: identify
  one_of:
  - identification: email
    steps:
    - name: authenticate_primary_email
      type: authenticate
      one_of:
      - authentication: primary_oob_otp_email
        target_step: identify
        steps:
        - name: authenticate_secondary_email
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
name: default
steps:
- name: identify
  type: identify
  one_of:
  - identification: phone
    steps:
    - name: authenticate_primary_phone
      type: authenticate
      one_of:
      - authentication: primary_oob_otp_sms
        target_step: identify
        steps:
        - name: authenticate_secondary_phone
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
        steps:
        - type: change_password
          target_step: authenticate_primary_username
        - name: authenticate_secondary_username
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
name: default
steps:
- name: identify
  type: identify
  one_of:
  - identification: email
    steps:
    - name: authenticate_primary_email
      type: authenticate
      one_of:
      - authentication: primary_password
        steps:
        - type: change_password
          target_step: authenticate_primary_email
        - name: authenticate_secondary_email
          type: authenticate
          optional: true
          one_of:
          - authentication: device_token
          - authentication: recovery_code
          - authentication: secondary_totp
      - authentication: primary_oob_otp_email
        target_step: identify
        steps:
        - name: authenticate_secondary_email
          type: authenticate
          optional: true
          one_of:
          - authentication: device_token
          - authentication: recovery_code
          - authentication: secondary_totp
  - identification: phone
    steps:
    - name: authenticate_primary_phone
      type: authenticate
      one_of:
      - authentication: primary_password
        steps:
        - type: change_password
          target_step: authenticate_primary_phone
        - name: authenticate_secondary_phone
          type: authenticate
          optional: true
          one_of:
          - authentication: device_token
          - authentication: recovery_code
          - authentication: secondary_totp
      - authentication: primary_oob_otp_sms
        target_step: identify
        steps:
        - name: authenticate_secondary_phone
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
name: default
steps:
- name: identify
  type: identify
  one_of:
  - identification: email
    steps:
    - name: authenticate_primary_email
      type: authenticate
      one_of:
      - authentication: primary_password
        steps:
        - type: change_password
          target_step: authenticate_primary_email
        - name: authenticate_secondary_email
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
name: default
steps:
- name: identify
  type: identify
  one_of:
  - identification: email
    steps:
    - name: authenticate_primary_email
      type: authenticate
      one_of:
      - authentication: primary_password
        steps:
        - type: change_password
          target_step: authenticate_primary_email
        - name: authenticate_secondary_email
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
name: default
steps:
- name: identify
  type: identify
  one_of:
  - identification: email
    steps:
    - name: authenticate_primary_email
      type: authenticate
      one_of:
      - authentication: primary_password
        steps:
        - name: authenticate_secondary_email
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
    - name: authenticate_primary_email
      type: authenticate
      one_of:
      - authentication: primary_password
        steps:
        - type: change_password
          target_step: authenticate_primary_email
        - name: authenticate_secondary_email
          type: authenticate
          one_of:
          - authentication: secondary_totp
  - identification: oauth
`)

		// passkey
		test(`
authentication:
  identities:
  - login_id
  - passkey
  primary_authenticators:
  - password
  - passkey
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
    - name: authenticate_primary_email
      type: authenticate
      one_of:
      - authentication: primary_password
        steps:
        - type: change_password
          target_step: authenticate_primary_email
        - name: authenticate_secondary_email
          type: authenticate
          optional: true
          one_of:
          - authentication: device_token
          - authentication: recovery_code
          - authentication: secondary_totp
      - authentication: primary_passkey
  - identification: passkey
- type: prompt_create_passkey
`)

		// passkey and oauth and 2fa
		test(`
authentication:
  identities:
  - login_id
  - oauth
  - passkey
  primary_authenticators:
  - password
  - passkey
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
    - name: authenticate_primary_email
      type: authenticate
      one_of:
      - authentication: primary_password
        steps:
        - type: change_password
          target_step: authenticate_primary_email
        - name: authenticate_secondary_email
          type: authenticate
          one_of:
          - authentication: secondary_totp
      - authentication: primary_passkey
  - identification: oauth
  - identification: passkey
- type: prompt_create_passkey
`)
	})
}
