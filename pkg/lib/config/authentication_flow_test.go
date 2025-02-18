package config

import (
	"bytes"
	"context"
	"encoding/json"
	"testing"

	"sigs.k8s.io/yaml"

	. "github.com/smartystreets/goconvey/convey"
)

func TestAuthenticationFlowSignupFlow(t *testing.T) {
	ctx := context.Background()
	Convey("AuthenticationFlowSignupFlow", t, func() {
		test := func(inputYAML string) {
			inputJSON, err := yaml.YAMLToJSON([]byte(inputYAML))
			So(err, ShouldBeNil)

			err = Schema.PartValidator("AuthenticationFlowConfig").Validate(ctx, bytes.NewReader(inputJSON))
			So(err, ShouldBeNil)

			var cfg AuthenticationFlowConfig
			err = json.Unmarshal([]byte(inputJSON), &cfg)
			So(err, ShouldBeNil)

			var input interface{}
			err = json.Unmarshal([]byte(inputJSON), &input)
			So(err, ShouldBeNil)

			encodedCfg, err := json.Marshal(cfg)
			So(err, ShouldBeNil)

			encodedInput, err := json.Marshal(input)
			So(err, ShouldBeNil)

			err = Schema.PartValidator("AuthenticationFlowConfig").Validate(ctx, bytes.NewReader(encodedCfg))
			So(err, ShouldBeNil)

			So(string(encodedInput), ShouldEqualJSON, string(encodedCfg))
		}

		test(`
signup_flows:
- name: signup_flow
  steps:
  - type: identify
    name: my_step
    one_of:
    - identification: email
  - type: create_authenticator
    one_of:
    - authentication: primary_password
  - type: verify
    target_step: my_step
  - type: fill_in_user_profile
    user_profile:
    - pointer: /given_name
      required: true
`)
	})
}

func TestAuthenticationFlowLoginFlow(t *testing.T) {
	ctx := context.Background()
	Convey("AuthenticationFlowLoginFlow", t, func() {
		test := func(inputYAML string) {
			inputJSON, err := yaml.YAMLToJSON([]byte(inputYAML))
			So(err, ShouldBeNil)

			err = Schema.PartValidator("AuthenticationFlowConfig").Validate(ctx, bytes.NewReader(inputJSON))
			So(err, ShouldBeNil)

			var cfg AuthenticationFlowConfig
			err = json.Unmarshal([]byte(inputJSON), &cfg)
			So(err, ShouldBeNil)

			var input interface{}
			err = json.Unmarshal([]byte(inputJSON), &input)
			So(err, ShouldBeNil)

			encodedCfg, err := json.Marshal(cfg)
			So(err, ShouldBeNil)

			encodedInput, err := json.Marshal(input)
			So(err, ShouldBeNil)

			err = Schema.PartValidator("AuthenticationFlowConfig").Validate(ctx, bytes.NewReader(encodedCfg))
			So(err, ShouldBeNil)

			So(string(encodedInput), ShouldEqualJSON, string(encodedCfg))
		}

		test(`
login_flows:
- name: login_flow
  steps:
  - type: identify
    name: my_step
    one_of:
    - identification: email
  - type: authenticate
    one_of:
    - authentication: primary_password
  - type: authenticate
    optional: true
    one_of:
    - authentication: secondary_totp
`)
	})
}

func TestAuthenticationFlowSignupLoginFlow(t *testing.T) {
	ctx := context.Background()
	Convey("AuthenticationFlowSignupLoginFlow", t, func() {
		test := func(inputYAML string) {
			inputJSON, err := yaml.YAMLToJSON([]byte(inputYAML))
			So(err, ShouldBeNil)

			err = Schema.PartValidator("AuthenticationFlowConfig").Validate(ctx, bytes.NewReader(inputJSON))
			So(err, ShouldBeNil)

			var cfg AuthenticationFlowConfig
			err = json.Unmarshal([]byte(inputJSON), &cfg)
			So(err, ShouldBeNil)

			var input interface{}
			err = json.Unmarshal([]byte(inputJSON), &input)
			So(err, ShouldBeNil)

			encodedCfg, err := json.Marshal(cfg)
			So(err, ShouldBeNil)

			encodedInput, err := json.Marshal(input)
			So(err, ShouldBeNil)

			err = Schema.PartValidator("AuthenticationFlowConfig").Validate(ctx, bytes.NewReader(encodedCfg))
			So(err, ShouldBeNil)

			So(string(encodedInput), ShouldEqualJSON, string(encodedCfg))
		}

		test(`
signup_login_flows:
- name: signup_login_flow
  steps:
  - type: identify
    one_of:
    - identification: email
      signup_flow: a
      login_flow: b
`)
	})
}

func TestAuthenticationFlowReauthFlow(t *testing.T) {
	ctx := context.Background()
	Convey("AuthenticationFlowReauthFlow", t, func() {
		test := func(inputYAML string) {
			inputJSON, err := yaml.YAMLToJSON([]byte(inputYAML))
			So(err, ShouldBeNil)

			err = Schema.PartValidator("AuthenticationFlowConfig").Validate(ctx, bytes.NewReader(inputJSON))
			So(err, ShouldBeNil)

			var cfg AuthenticationFlowConfig
			err = json.Unmarshal([]byte(inputJSON), &cfg)
			So(err, ShouldBeNil)

			var input interface{}
			err = json.Unmarshal([]byte(inputJSON), &input)
			So(err, ShouldBeNil)

			encodedCfg, err := json.Marshal(cfg)
			So(err, ShouldBeNil)

			encodedInput, err := json.Marshal(input)
			So(err, ShouldBeNil)

			err = Schema.PartValidator("AuthenticationFlowConfig").Validate(ctx, bytes.NewReader(encodedCfg))
			So(err, ShouldBeNil)

			So(string(encodedInput), ShouldEqualJSON, string(encodedCfg))
		}

		test(`
reauth_flows:
- name: reauth_flow
  steps:
  - type: identify
    one_of:
    - identification: id_token
  - type: authenticate
    one_of:
    - authentication: primary_password
  - type: authenticate
    one_of:
    - authentication: secondary_totp
`)
	})
}

func TestAuthenticationFlowAccountRecoveryFlow(t *testing.T) {
	ctx := context.Background()
	Convey("AuthenticationFlowAccountRecoveryFlow", t, func() {
		test := func(inputYAML string) {
			inputJSON, err := yaml.YAMLToJSON([]byte(inputYAML))
			So(err, ShouldBeNil)

			err = Schema.PartValidator("AuthenticationFlowConfig").Validate(ctx, bytes.NewReader(inputJSON))
			So(err, ShouldBeNil)

			var cfg AuthenticationFlowConfig
			err = json.Unmarshal([]byte(inputJSON), &cfg)
			So(err, ShouldBeNil)

			var input interface{}
			err = json.Unmarshal([]byte(inputJSON), &input)
			So(err, ShouldBeNil)

			encodedCfg, err := json.Marshal(cfg)
			So(err, ShouldBeNil)

			encodedInput, err := json.Marshal(input)
			So(err, ShouldBeNil)

			err = Schema.PartValidator("AuthenticationFlowConfig").Validate(ctx, bytes.NewReader(encodedCfg))
			So(err, ShouldBeNil)

			So(string(encodedInput), ShouldEqualJSON, string(encodedCfg))
		}

		test(`
account_recovery_flows:
- name: default
  steps:
    - type: identify
      one_of:
      - identification: email
        on_failure: ignore
        steps:
        - type: select_destination
          enumerate_destinations: true
          allowed_channels:
            - channel: email
              otp_form: link
            - channel: sms
              otp_form: code
            - channel: sms
              otp_form: link
      - identification: phone
        on_failure: ignore
        steps:
        - type: select_destination
          enumerate_destinations: true
          allowed_channels:
            - channel: sms
              otp_form: code
            - channel: sms
              otp_form: link
            - channel: whatsapp
              otp_form: code
    - type: verify_account_recovery_code
    - type: reset_password
`)
	})
}
