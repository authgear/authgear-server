package declarative_test

import (
	"encoding/json"
	"testing"

	"github.com/authgear/authgear-server/pkg/lib/config"
	. "github.com/smartystreets/goconvey/convey"
	yaml "sigs.k8s.io/yaml"

	"github.com/authgear/authgear-server/pkg/lib/authenticationflow/declarative"
)

func TestIsLastAuthentication(t *testing.T) {
	Convey("IsLastAuthentication", t, func() {
		Convey("login", func() {
			cfgYAML := `
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
`

			var cfg *config.AuthenticationFlowLoginFlow
			jsonBytes, err := yaml.YAMLToJSON([]byte(cfgYAML))
			So(err, ShouldBeNil)
			err = json.Unmarshal(jsonBytes, &cfg)
			So(err, ShouldBeNil)

			Convey("should return false if authentication step exists from skip 0", func() {
				So(declarative.IsLastAuthentication(cfg, 0), ShouldBeFalse)
			})

			Convey("should return true if no authentication step exists after skipping first step", func() {
				So(declarative.IsLastAuthentication(cfg, 1), ShouldBeTrue)
			})

			Convey("should return true if no authentication step exists after skipping first two steps", func() {
				So(declarative.IsLastAuthentication(cfg, 2), ShouldBeTrue)
			})

			Convey("should return true if no authentication step exists after skipping all steps", func() {
				So(declarative.IsLastAuthentication(cfg, 3), ShouldBeTrue)
			})
		})

		Convey("signup", func() {
			cfgYAML := `
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
- type: prompt_create_passkey
`

			var cfg *config.AuthenticationFlowSignupFlow
			jsonBytes, err := yaml.YAMLToJSON([]byte(cfgYAML))
			So(err, ShouldBeNil)
			err = json.Unmarshal(jsonBytes, &cfg)
			So(err, ShouldBeNil)

			Convey("should return false if authentication step exists from skip 0", func() {
				So(declarative.IsLastAuthentication(cfg, 0), ShouldBeFalse)
			})

			Convey("should return true if no authentication step exists after skipping first step", func() {
				So(declarative.IsLastAuthentication(cfg, 1), ShouldBeTrue)
			})

			Convey("should return true if no authentication step exists after skipping first two steps", func() {
				So(declarative.IsLastAuthentication(cfg, 2), ShouldBeTrue)
			})
		})

		Convey("login flow nested steps", func() {
			cfgYAML := `
authentication: primary_password
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
`

			var cfg *config.AuthenticationFlowLoginFlowOneOf
			jsonBytes, err := yaml.YAMLToJSON([]byte(cfgYAML))
			So(err, ShouldBeNil)
			err = json.Unmarshal(jsonBytes, &cfg)
			So(err, ShouldBeNil)

			Convey("should return false if authentication step exists from skip 0", func() {
				So(declarative.IsLastAuthentication(cfg, 0), ShouldBeFalse)
			})

			Convey("should return true if no authentication step exists after skipping first step", func() {
				So(declarative.IsLastAuthentication(cfg, 1), ShouldBeTrue)
			})
		})

		Convey("signup nested steps", func() {
			cfgYAML := `
identification: email
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
- type: view_recovery_code
`

			var cfg *config.AuthenticationFlowSignupFlowOneOf
			jsonBytes, err := yaml.YAMLToJSON([]byte(cfgYAML))
			So(err, ShouldBeNil)
			err = json.Unmarshal(jsonBytes, &cfg)
			So(err, ShouldBeNil)

			Convey("should return false if authentication step exists from skip 0", func() {
				So(declarative.IsLastAuthentication(cfg, 0), ShouldBeFalse)
			})

			Convey("should return false if no authentication step exists after skipping first step", func() {
				So(declarative.IsLastAuthentication(cfg, 1), ShouldBeFalse)
			})

			Convey("should return false if no authentication step exists after skipping two steps", func() {
				So(declarative.IsLastAuthentication(cfg, 2), ShouldBeFalse)
			})

			Convey("should return true if no authentication step exists after skipping three steps", func() {
				So(declarative.IsLastAuthentication(cfg, 3), ShouldBeTrue)
			})
		})
	})
}
