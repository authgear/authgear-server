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

func TestGenerateSignupLoginFlowConfig(t *testing.T) {
	Convey("GenerateSignupLoginFlowConfig", t, func() {
		test := func(cfgStr string, expected string) {
			jsonData, err := yaml.YAMLToJSON([]byte(cfgStr))
			So(err, ShouldBeNil)

			var appConfig config.AppConfig
			decoder := json.NewDecoder(bytes.NewReader(jsonData))
			err = decoder.Decode(&appConfig)
			So(err, ShouldBeNil)

			config.PopulateDefaultValues(&appConfig)

			flow := GenerateSignupLoginFlowConfig(&appConfig)
			flowJSON, err := json.Marshal(flow)
			So(err, ShouldBeNil)

			expectedJSON, err := yaml.YAMLToJSON([]byte(expected))
			So(err, ShouldBeNil)

			So(string(flowJSON), ShouldEqualJSON, string(expectedJSON))
		}

		// email
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
- name: signup_login_identify
  type: identify
  one_of:
  - identification: email
    signup_flow: default
    login_flow: default
`)

		// phone
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
- name: signup_login_identify
  type: identify
  one_of:
  - identification: phone
    signup_flow: default
    login_flow: default
`)

		// username
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
- name: signup_login_identify
  type: identify
  one_of:
  - identification: username
    signup_flow: default
    login_flow: default
`)

		// email,phone
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
- name: signup_login_identify
  type: identify
  one_of:
  - identification: email
    signup_flow: default
    login_flow: default
  - identification: phone
    signup_flow: default
    login_flow: default
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
- name: signup_login_identify
  type: identify
  one_of:
  - identification: oauth
    signup_flow: default
    login_flow: default
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
- name: signup_login_identify
  type: identify
  one_of:
  - identification: email
    signup_flow: default
    login_flow: default
  - identification: passkey
    login_flow: default
`)

		// all
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
    - type: phone
    - type: username
  oauth:
    providers:
    - alias: google
      type: google
`, `
name: default
steps:
- name: signup_login_identify
  type: identify
  one_of:
  - identification: email
    signup_flow: default
    login_flow: default
  - identification: phone
    signup_flow: default
    login_flow: default
  - identification: username
    signup_flow: default
    login_flow: default
  - identification: oauth
    signup_flow: default
    login_flow: default
  - identification: passkey
    login_flow: default
`)
		// bot_protection, 1 branch
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
- name: signup_login_identify
  type: identify
  one_of:
  - identification: email
    bot_protection:
      mode: always
      provider: 
        type: recaptchav2
    signup_flow: default
    login_flow: default
`)
		// bot_protection, all loginID branches except oauth & passkey
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
    - type: phone
    - type: username
  oauth:
    providers:
    - alias: google
      type: google
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
- name: signup_login_identify
  type: identify
  one_of:
  - identification: email
    bot_protection:
      mode: always
      provider: 
        type: recaptchav2
    signup_flow: default
    login_flow: default
  - identification: phone
    bot_protection:
      mode: always
      provider: 
        type: recaptchav2
    signup_flow: default
    login_flow: default
  - identification: username
    bot_protection:
      mode: always
      provider: 
        type: recaptchav2
    signup_flow: default
    login_flow: default
  - identification: oauth
    signup_flow: default
    login_flow: default
  - identification: passkey
    login_flow: default
`)
	})
}
