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
- name: signup_identify
  type: identify
  one_of:
  - identification: email
    steps:
    - target_step: signup_identify
      type: verify
    - name: authenticate_primary_email
      type: create_authenticator
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
- name: signup_identify
  type: identify
  one_of:
  - identification: email
    steps:
    - target_step: signup_identify
      type: verify
    - name: authenticate_primary_email
      type: create_authenticator
      one_of:
      - authentication: primary_oob_otp_email
        target_step: signup_identify
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
- name: signup_identify
  type: identify
  one_of:
  - identification: phone
    steps:
    - target_step: signup_identify
      type: verify
    - name: authenticate_primary_phone
      type: create_authenticator
      one_of:
      - authentication: primary_oob_otp_sms
        target_step: signup_identify
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
- name: signup_identify
  type: identify
  one_of:
  - identification: username
    steps:
    - name: authenticate_primary_username
      type: create_authenticator
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
- name: signup_identify
  type: identify
  one_of:
  - identification: email
    steps:
    - target_step: signup_identify
      type: verify
    - name: authenticate_primary_email
      type: create_authenticator
      one_of:
      - authentication: primary_password
      - authentication: primary_oob_otp_email
        target_step: signup_identify
  - identification: phone
    steps:
    - target_step: signup_identify
      type: verify
    - name: authenticate_primary_phone
      type: create_authenticator
      one_of:
      - authentication: primary_password
      - authentication: primary_oob_otp_sms
        target_step: signup_identify
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
- name: signup_identify
  type: identify
  one_of:
  - identification: email
    steps:
    - target_step: signup_identify
      type: verify
    - name: authenticate_primary_email
      type: create_authenticator
      one_of:
      - authentication: primary_password
    - name: authenticate_secondary_email
      type: create_authenticator
      one_of:
      - authentication: secondary_totp
        steps:
        - type: view_recovery_code
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
- name: signup_identify
  type: identify
  one_of:
  - identification: email
    steps:
    - target_step: signup_identify
      type: verify
    - name: authenticate_primary_email
      type: create_authenticator
      one_of:
      - authentication: primary_password
    - name: authenticate_secondary_email
      type: create_authenticator
      one_of:
      - authentication: secondary_oob_otp_sms
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
- name: signup_identify
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
- name: signup_identify
  type: identify
  one_of:
  - identification: email
    steps:
    - target_step: signup_identify
      type: verify
    - name: authenticate_primary_email
      type: create_authenticator
      one_of:
      - authentication: primary_password
    - name: authenticate_secondary_email
      type: create_authenticator
      one_of:
      - authentication: secondary_totp
  - identification: oauth
`)

		// ldap
		test(`
authentication:
  identities:
  - ldap
identity:
  ldap:
    servers:
    - name: ldap
`, `
name: default
steps:
- name: signup_identify
  type: identify
  one_of:
  - identification: ldap
`)
		// ldap, otp
		test(`
authentication:
  identities:
  - ldap
  secondary_authenticators:
  - oob_otp_sms
  secondary_authentication_mode: required
  recovery_code:
    disabled: true
identity:
  ldap:
    servers:
    - name: ldap
`, `
name: default
steps:
- name: signup_identify
  type: identify
  one_of:
  - identification: ldap
    steps:
    - name: authenticate_secondary_ldap
      type: create_authenticator
      one_of:
      - authentication: secondary_oob_otp_sms
`)
		// ldap, totp,recovery_code
		test(`
authentication:
  identities:
  - ldap
  secondary_authenticators:
  - totp
  secondary_authentication_mode: required
identity:
  ldap:
    servers:
    - name: ldap
`, `
name: default
steps:
- name: signup_identify
  type: identify
  one_of:
  - identification: ldap
    steps:
    - name: authenticate_secondary_ldap
      type: create_authenticator
      one_of:
      - authentication: secondary_totp
        steps:
        - type: view_recovery_code
`)

		// bot_protection, 3 branches
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
  secondary_authentication_mode: required
identity:
  login_id:
    keys:
    - type: email
    - type: phone
    - type: username
bot_protection:
  enabled: true
  provider:
    type: recaptchav2
    site_key: recaptchav2-site-key
  requirements:
    signup_or_login:
      mode: always
    oob_otp_email:
      mode: always
    oob_otp_sms:
      mode: always
`, `
name: default
steps:
- name: signup_identify
  one_of:
  - identification: email
    bot_protection:
      mode: always
    steps:
    - target_step: signup_identify
      type: verify
    - name: authenticate_primary_email
      one_of:
      - authentication: primary_password
      - authentication: primary_oob_otp_email
        bot_protection:
          mode: always
        target_step: signup_identify
      type: create_authenticator
    - name: authenticate_secondary_email
      one_of:
      - authentication: secondary_oob_otp_email
        bot_protection:
          mode: always
        steps:
        - type: view_recovery_code
      - authentication: secondary_oob_otp_sms
        bot_protection:
          mode: always
        steps:
        - type: view_recovery_code
      type: create_authenticator
  - identification: phone
    bot_protection:
      mode: always
    steps:
    - target_step: signup_identify
      type: verify
    - name: authenticate_primary_phone
      type: create_authenticator
      one_of:
      - authentication: primary_password
      - authentication: primary_oob_otp_sms
        target_step: signup_identify
        bot_protection:
          mode: always
    - name: authenticate_secondary_phone
      one_of:
      - authentication: secondary_oob_otp_email
        bot_protection:
          mode: always
        steps:
        - type: view_recovery_code
      - authentication: secondary_oob_otp_sms
        bot_protection:
          mode: always
        steps:
        - type: view_recovery_code
      type: create_authenticator
  - identification: username
    bot_protection:
      mode: always
    steps:
    - name: authenticate_primary_username
      one_of:
      - authentication: primary_password
      type: create_authenticator
    - name: authenticate_secondary_username
      one_of:
      - authentication: secondary_oob_otp_email
        bot_protection:
          mode: always
        steps:
        - type: view_recovery_code
      - authentication: secondary_oob_otp_sms
        bot_protection:
          mode: always
        steps:
        - type: view_recovery_code
      type: create_authenticator
  type: identify
`)
		// bot_protection, stricter risk mode overrides
		// Note
		// - identify > email      : mode=always
		// - identify > phone      : mode=always
		// - identify > username   : mode=never
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
    - type: username
bot_protection:
  enabled: true
  provider:
    type: recaptchav2
    site_key: recaptchav2-site-key
  requirements:
    signup_or_login:
      mode: never
    oob_otp_email:
      mode: always
    oob_otp_sms:
      mode: always
`, `
name: default
steps:
- name: signup_identify
  one_of:
  - identification: email
    bot_protection:
      mode: always
    steps:
    - target_step: signup_identify
      type: verify
    - name: authenticate_primary_email
      one_of:
      - authentication: primary_password
      - authentication: primary_oob_otp_email
        bot_protection:
          mode: always
        target_step: signup_identify
      type: create_authenticator
  - identification: phone
    bot_protection:
      mode: always
    steps:
    - target_step: signup_identify
      type: verify
    - name: authenticate_primary_phone
      type: create_authenticator
      one_of:
      - authentication: primary_password
      - authentication: primary_oob_otp_sms
        target_step: signup_identify
        bot_protection:
          mode: always
  - identification: username
    bot_protection:
      mode: never
    steps:
    - name: authenticate_primary_username
      one_of:
      - authentication: primary_password
      type: create_authenticator
  type: identify
`)
	})
}
