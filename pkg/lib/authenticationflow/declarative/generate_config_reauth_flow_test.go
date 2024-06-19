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

func TestGenerateReauthFlowConfig(t *testing.T) {
	Convey("GenerateReauthFlowConfig", t, func() {
		test := func(cfgStr string, expected string) {
			jsonData, err := yaml.YAMLToJSON([]byte(cfgStr))
			So(err, ShouldBeNil)

			var appConfig config.AppConfig
			decoder := json.NewDecoder(bytes.NewReader(jsonData))
			err = decoder.Decode(&appConfig)
			So(err, ShouldBeNil)

			config.PopulateDefaultValues(&appConfig)

			flow := GenerateReauthFlowConfig(&appConfig)
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
  secondary_authenticators: []
identity:
  login_id:
    keys:
    - type: email
`, `
name: default
steps:
- name: reauth_identify
  type: identify
  one_of:
  - identification: id_token
- name: reauthenticate
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
  secondary_authenticators: []
identity:
  login_id:
    keys:
    - type: email
`, `
name: default
steps:
- name: reauth_identify
  type: identify
  one_of:
  - identification: id_token
- name: reauthenticate
  type: authenticate
  one_of:
  - authentication: primary_oob_otp_email
`)

		// phone, otp
		test(`
authentication:
  identities:
  - login_id
  primary_authenticators:
  - oob_otp_sms
  secondary_authenticators: []
identity:
  login_id:
    keys:
    - type: phone
`, `
name: default
steps:
- name: reauth_identify
  type: identify
  one_of:
  - identification: id_token
- name: reauthenticate
  type: authenticate
  one_of:
  - authentication: primary_oob_otp_sms
`)

		// username, password
		test(`
authentication:
  identities:
  - login_id
  primary_authenticators:
  - password
  secondary_authenticators: []
identity:
  login_id:
    keys:
    - type: username
`, `
name: default
steps:
- name: reauth_identify
  type: identify
  one_of:
  - identification: id_token
- name: reauthenticate
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
  secondary_authenticators: []
identity:
  login_id:
    keys:
    - type: email
    - type: phone
`, `
name: default
steps:
- name: reauth_identify
  type: identify
  one_of:
  - identification: id_token
- name: reauthenticate
  type: authenticate
  one_of:
  - authentication: primary_password
  - authentication: primary_oob_otp_email
  - authentication: primary_oob_otp_sms
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
- name: reauth_identify
  type: identify
  one_of:
  - identification: id_token
- name: reauthenticate
  type: authenticate
  one_of:
  - authentication: primary_password
  - authentication: secondary_totp
`)

		// oauth
		test(`
authentication:
  identities:
  - oauth
  primary_authenticators: []
  secondary_authenticators: []
identity:
  oauth:
    providers:
    - alias: google
      type: google
`, `
name: default
steps:
- name: reauth_identify
  type: identify
  one_of:
  - identification: id_token
- name: reauthenticate
  type: authenticate
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
  secondary_authenticators: []
identity:
  login_id:
    keys:
    - type: email
`, `
name: default
steps:
- name: reauth_identify
  type: identify
  one_of:
  - identification: id_token
- name: reauthenticate
  type: authenticate
  one_of:
  - authentication: primary_password
  - authentication: primary_passkey
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
- name: reauth_identify
  type: identify
  one_of:
  - identification: id_token
- name: reauthenticate
  type: authenticate
  one_of:
  - authentication: primary_password
  - authentication: primary_passkey
  - authentication: secondary_totp
`)
		// captcha, 1 branch
		test(`
authentication:
  identities:
  - login_id
  primary_authenticators:
  - password
  secondary_authenticators: []
identity:
  login_id:
    keys:
    - type: email
captcha:
  enabled: true
  providers:
  - type: recaptchav2
    alias: recaptchav2-a
    site_key: some-site-key
`, `
name: default
steps:
- name: reauth_identify
  type: identify
  one_of:
  - identification: id_token
- name: reauthenticate
  type: authenticate
  one_of:
  - authentication: primary_password
    captcha:
      mode: never
`)
		// captcha, 3 branches
		test(`
authentication:
  identities:
  - login_id
  primary_authenticators:
  - password
  - oob_otp_email
  - oob_otp_sms
  secondary_authenticators: []
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
`, `
name: default
steps:
- name: reauth_identify
  type: identify
  one_of:
  - identification: id_token
- name: reauthenticate
  type: authenticate
  one_of:
  - authentication: primary_password
    captcha:
      mode: never
  - authentication: primary_oob_otp_email
    captcha:
      mode: never
  - authentication: primary_oob_otp_sms
    captcha:
      mode: never
`)
	})
}
