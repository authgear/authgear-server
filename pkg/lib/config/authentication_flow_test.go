package config

import (
	"bytes"
	"encoding/json"
	"testing"

	"sigs.k8s.io/yaml"

	. "github.com/smartystreets/goconvey/convey"
)

func TestAuthenticationFlowSignupFlow(t *testing.T) {
	Convey("AuthenticationFlowSignupFlow", t, func() {
		test := func(inputYAML string) {
			inputJSON, err := yaml.YAMLToJSON([]byte(inputYAML))
			So(err, ShouldBeNil)

			err = Schema.PartValidator("AuthenticationFlowConfig").Validate(bytes.NewReader(inputJSON))
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

			err = Schema.PartValidator("AuthenticationFlowConfig").Validate(bytes.NewReader(encodedCfg))
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
	Convey("AuthenticationFlowLoginFlow", t, func() {
		test := func(inputYAML string) {
			inputJSON, err := yaml.YAMLToJSON([]byte(inputYAML))
			So(err, ShouldBeNil)

			err = Schema.PartValidator("AuthenticationFlowConfig").Validate(bytes.NewReader(inputJSON))
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

			err = Schema.PartValidator("AuthenticationFlowConfig").Validate(bytes.NewReader(encodedCfg))
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
	Convey("AuthenticationFlowSignupLoginFlow", t, func() {
		test := func(inputYAML string) {
			inputJSON, err := yaml.YAMLToJSON([]byte(inputYAML))
			So(err, ShouldBeNil)

			err = Schema.PartValidator("AuthenticationFlowConfig").Validate(bytes.NewReader(inputJSON))
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

			err = Schema.PartValidator("AuthenticationFlowConfig").Validate(bytes.NewReader(encodedCfg))
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
	Convey("AuthenticationFlowReauthFlow", t, func() {
		test := func(inputYAML string) {
			inputJSON, err := yaml.YAMLToJSON([]byte(inputYAML))
			So(err, ShouldBeNil)

			err = Schema.PartValidator("AuthenticationFlowConfig").Validate(bytes.NewReader(inputJSON))
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

			err = Schema.PartValidator("AuthenticationFlowConfig").Validate(bytes.NewReader(encodedCfg))
			So(err, ShouldBeNil)

			So(string(encodedInput), ShouldEqualJSON, string(encodedCfg))
		}

		test(`
reauth_flows:
- name: reauth_flow
  steps:
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

func TestAuthenticationFlowRequestAccountRecoveryFlow(t *testing.T) {
	Convey("AuthenticationFlowRequestAccountRecoveryFlow", t, func() {
		test := func(inputYAML string) {
			inputJSON, err := yaml.YAMLToJSON([]byte(inputYAML))
			So(err, ShouldBeNil)

			err = Schema.PartValidator("AuthenticationFlowConfig").Validate(bytes.NewReader(inputJSON))
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

			err = Schema.PartValidator("AuthenticationFlowConfig").Validate(bytes.NewReader(encodedCfg))
			So(err, ShouldBeNil)

			So(string(encodedInput), ShouldEqualJSON, string(encodedCfg))
		}

		test(`
request_account_recovery_flows:
- name: default
  steps:
    - type: identify
      one_of:
      - identification: email
        on_failure: ignore
        enumerate_destinations: true
      - identification: phone
        on_failure: ignore
        enumerate_destinations: true
    - type: select_destination
`)
	})
}
