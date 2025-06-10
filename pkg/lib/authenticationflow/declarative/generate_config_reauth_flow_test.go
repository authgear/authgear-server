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
		test := func(name string, cfgStr string, expected string) {
			Convey(name, func() {
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
			})
		}

		// email, password
		test("test-1", `
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
- name: reauthenticate_amr_constraints
  show_until_amr_constraints_fulfilled: true
  type: authenticate
`)

		// email, otp
		test("test-2", `
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
- name: reauthenticate_amr_constraints
  show_until_amr_constraints_fulfilled: true
  type: authenticate
`)

		// phone, otp
		test("test-3", `
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
- name: reauthenticate_amr_constraints
  show_until_amr_constraints_fulfilled: true
  type: authenticate
`)

		// username, password
		test("test-4", `
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
- name: reauthenticate_amr_constraints
  show_until_amr_constraints_fulfilled: true
  type: authenticate
`)

		// email,phone, password,otp
		test("test-5", `
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
- name: reauthenticate_amr_constraints
  show_until_amr_constraints_fulfilled: true
  type: authenticate
`)

		// email,password, totp,recovery_code
		test("test-6", `
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
- name: reauthenticate_amr_constraints
  one_of:
  - authentication: secondary_totp
  show_until_amr_constraints_fulfilled: true
  type: authenticate
`)

		// oauth
		test("test-7", `
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
- name: reauthenticate_amr_constraints
  show_until_amr_constraints_fulfilled: true
  type: authenticate
`)

		// passkey
		test("test-8", `
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
- name: reauthenticate_amr_constraints
  one_of:
  - authentication: primary_passkey
  show_until_amr_constraints_fulfilled: true
  type: authenticate
`)

		// passkey and oauth and 2fa
		test("test-9", `
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
- name: reauthenticate_amr_constraints
  one_of:
  - authentication: primary_passkey
  - authentication: secondary_totp
  show_until_amr_constraints_fulfilled: true
  type: authenticate
`)
		// bot_protection should have no effect on reauth
		test("test-10", `
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
bot_protection:
  enabled: true
  provider:
    type: recaptchav2
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
- name: reauthenticate_amr_constraints
  show_until_amr_constraints_fulfilled: true
  type: authenticate
`)
		// bot_protection should have no effect on reauth
		test("test-11", `
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
bot_protection:
  enabled: true
  provider:
    type: recaptchav2
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
  - authentication: primary_oob_otp_email
  - authentication: primary_oob_otp_sms
- name: reauthenticate_amr_constraints
  show_until_amr_constraints_fulfilled: true
  type: authenticate
`)
	})
}
