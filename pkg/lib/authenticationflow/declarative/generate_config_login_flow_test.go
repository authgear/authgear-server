package declarative

import (
	"bytes"
	"encoding/json"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
	"sigs.k8s.io/yaml"

	"github.com/authgear/authgear-server/pkg/lib/config"
	_ "github.com/authgear/authgear-server/pkg/lib/oauthrelyingparty/google"
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
- name: login_identify
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
          - authentication: secondary_totp
          - authentication: recovery_code
          - authentication: device_token
        - type: change_password
          target_step: authenticate_primary_email
- type: check_account_status
- type: terminate_other_sessions
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
- name: login_identify
  type: identify
  one_of:
  - identification: email
    steps:
    - name: authenticate_primary_email
      type: authenticate
      one_of:
      - authentication: primary_oob_otp_email
        target_step: login_identify
        steps:
        - name: authenticate_secondary_email
          type: authenticate
          optional: true
          one_of:
          - authentication: secondary_totp
          - authentication: recovery_code
          - authentication: device_token
- type: check_account_status
- type: terminate_other_sessions
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
- name: login_identify
  type: identify
  one_of:
  - identification: phone
    steps:
    - name: authenticate_primary_phone
      type: authenticate
      one_of:
      - authentication: primary_oob_otp_sms
        target_step: login_identify
        steps:
        - name: authenticate_secondary_phone
          type: authenticate
          optional: true
          one_of:
          - authentication: secondary_totp
          - authentication: recovery_code
          - authentication: device_token
- type: check_account_status
- type: terminate_other_sessions
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
- name: login_identify
  type: identify
  one_of:
  - identification: username
    steps:
    - name: authenticate_primary_username
      type: authenticate
      one_of:
      - authentication: primary_password
        steps:
        - name: authenticate_secondary_username
          type: authenticate
          optional: true
          one_of:
          - authentication: secondary_totp
          - authentication: recovery_code
          - authentication: device_token
        - type: change_password
          target_step: authenticate_primary_username
- type: check_account_status
- type: terminate_other_sessions
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
- name: login_identify
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
          - authentication: secondary_totp
          - authentication: recovery_code
          - authentication: device_token
        - type: change_password
          target_step: authenticate_primary_email
      - authentication: primary_oob_otp_email
        target_step: login_identify
        steps:
        - name: authenticate_secondary_email
          type: authenticate
          optional: true
          one_of:
          - authentication: secondary_totp
          - authentication: recovery_code
          - authentication: device_token
  - identification: phone
    steps:
    - name: authenticate_primary_phone
      type: authenticate
      one_of:
      - authentication: primary_password
        steps:
        - name: authenticate_secondary_phone
          type: authenticate
          optional: true
          one_of:
          - authentication: secondary_totp
          - authentication: recovery_code
          - authentication: device_token
        - type: change_password
          target_step: authenticate_primary_phone
      - authentication: primary_oob_otp_sms
        target_step: login_identify
        steps:
        - name: authenticate_secondary_phone
          type: authenticate
          optional: true
          one_of:
          - authentication: secondary_totp
          - authentication: recovery_code
          - authentication: device_token
- type: check_account_status
- type: terminate_other_sessions
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
- name: login_identify
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
          one_of:
          - authentication: secondary_totp
          - authentication: recovery_code
          - authentication: device_token
        - type: change_password
          target_step: authenticate_primary_email
- type: check_account_status
- type: terminate_other_sessions
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
- name: login_identify
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
          - authentication: secondary_totp
        - type: change_password
          target_step: authenticate_primary_email
- type: check_account_status
- type: terminate_other_sessions
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
- name: login_identify
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
          - authentication: secondary_totp
          - authentication: recovery_code
          - authentication: device_token
        - type: change_password
          target_step: authenticate_primary_email
- type: check_account_status
- type: terminate_other_sessions
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
- name: login_identify
  type: identify
  one_of:
  - identification: oauth
- type: check_account_status
- type: terminate_other_sessions
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
- name: login_identify
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
          one_of:
          - authentication: secondary_totp
        - type: change_password
          target_step: authenticate_primary_email
  - identification: oauth
- type: check_account_status
- type: terminate_other_sessions
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
- name: login_identify
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
          - authentication: secondary_totp
          - authentication: recovery_code
          - authentication: device_token
        - type: change_password
          target_step: authenticate_primary_email
      - authentication: primary_passkey
  - identification: passkey
- type: check_account_status
- type: terminate_other_sessions
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
- name: login_identify
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
          one_of:
          - authentication: secondary_totp
        - type: change_password
          target_step: authenticate_primary_email
      - authentication: primary_passkey
  - identification: oauth
  - identification: passkey
- type: check_account_status
- type: terminate_other_sessions
- type: prompt_create_passkey
`)
		// bot_protection.requirements.signup_or_login=always
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
bot_protection:
  enabled: true
  provider:
    type: recaptchav2
    site_key: some-site-key
  requirements:
    signup_or_login: 
      mode: always
`, `
name: default
steps:
- name: login_identify
  type: identify
  one_of:
  - identification: email
    bot_protection:
      mode: always
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
          - authentication: secondary_totp
          - authentication: recovery_code
          - authentication: device_token
        - type: change_password
          target_step: authenticate_primary_email
- type: check_account_status
- type: terminate_other_sessions
`)
		// bot_protection, all always
		test(`
authentication:
  identities:
  - login_id
  primary_authenticators:
  - password
  - oob_otp_email
  - oob_otp_sms
  secondary_authenticators:
  - oob_otp_email
  - oob_otp_sms
  - password
  secondary_authentication_mode: required
identity:
  login_id:
    keys:
    - type: email
    - type: phone
bot_protection:
  enabled: true
  provider:
    type: recaptchav2
    site_key: some-site-key
  requirements:
    signup_or_login: 
      mode: always
    password:
      mode: never
    oob_otp_email:
      mode: always
    oob_otp_sms:
      mode: always
`, `
name: default
steps:
- name: login_identify
  type: identify
  one_of:
  - bot_protection:
      mode: always
    identification: email
    steps:
    - name: authenticate_primary_email
      type: authenticate
      one_of:
      - authentication: primary_password
        bot_protection:
          mode: never
        steps:
        - name: authenticate_secondary_email
          one_of:
          - authentication: secondary_oob_otp_email
            bot_protection:
              mode: always
          - authentication: secondary_oob_otp_sms
            bot_protection:
              mode: always
          - authentication: secondary_password
            bot_protection:
              mode: never
          - authentication: recovery_code
          - authentication: device_token
          type: authenticate
        - target_step: authenticate_primary_email
          type: change_password
      - authentication: primary_oob_otp_email
        target_step: login_identify
        bot_protection:
          mode: always
        steps:
        - name: authenticate_secondary_email
          type: authenticate
          one_of:
          - authentication: secondary_oob_otp_email
            bot_protection:
              mode: always
          - authentication: secondary_oob_otp_sms
            bot_protection:
              mode: always
          - authentication: secondary_password
            bot_protection:
              mode: never
          - authentication: recovery_code
          - authentication: device_token
  - identification: phone
    bot_protection:
      mode: always
    steps:
    - name: authenticate_primary_phone
      type: authenticate
      one_of:
      - authentication: primary_password
        bot_protection:
          mode: never
        steps:
        - name: authenticate_secondary_phone
          type: authenticate
          one_of:
          - authentication: secondary_oob_otp_email
            bot_protection:
              mode: always
          - authentication: secondary_oob_otp_sms
            bot_protection:
              mode: always
          - authentication: secondary_password
            bot_protection:
              mode: never
          - authentication: recovery_code
          - authentication: device_token
        - target_step: authenticate_primary_phone
          type: change_password
      - authentication: primary_oob_otp_sms
        target_step: login_identify
        bot_protection:
          mode: always
        steps:
        - name: authenticate_secondary_phone
          type: authenticate
          one_of:
          - authentication: secondary_oob_otp_email
            bot_protection:
              mode: always
          - authentication: secondary_oob_otp_sms
            bot_protection:
              mode: always
          - authentication: secondary_password
            bot_protection:
              mode: never
          - authentication: recovery_code
          - authentication: device_token
- type: check_account_status
- type: terminate_other_sessions
`)
	})
}
