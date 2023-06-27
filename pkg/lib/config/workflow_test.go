package config

import (
	"bytes"
	"encoding/json"
	"testing"

	"sigs.k8s.io/yaml"

	. "github.com/smartystreets/goconvey/convey"
)

func TestWorkflowSignupFlow(t *testing.T) {
	Convey("WorkflowSignupFlow", t, func() {
		test := func(inputYAML string) {
			inputJSON, err := yaml.YAMLToJSON([]byte(inputYAML))
			So(err, ShouldBeNil)

			err = Schema.PartValidator("WorkflowConfig").Validate(bytes.NewReader(inputJSON))
			So(err, ShouldBeNil)

			var cfg WorkflowConfig
			err = json.Unmarshal([]byte(inputJSON), &cfg)
			So(err, ShouldBeNil)

			var input interface{}
			err = json.Unmarshal([]byte(inputJSON), &input)
			So(err, ShouldBeNil)

			encodedCfg, err := json.Marshal(cfg)
			So(err, ShouldBeNil)

			encodedInput, err := json.Marshal(input)
			So(err, ShouldBeNil)

			err = Schema.PartValidator("WorkflowConfig").Validate(bytes.NewReader(encodedCfg))
			So(err, ShouldBeNil)

			So(string(encodedInput), ShouldEqualJSON, string(encodedCfg))
		}

		test(`
signup_flows:
- id: signup_flow
  steps:
  - type: identify
    id: my_step
    one_of:
    - identification: email
  - type: authenticate
    one_of:
    - authentication: primary_password
  - type: verify
    target_step: my_step
  - type: user_profile
    user_profile:
    - pointer: /given_name
      required: true
`)
	})
}

func TestWorkflowLoginFlow(t *testing.T) {
	Convey("WorkflowLoginFlow", t, func() {
		test := func(inputYAML string) {
			inputJSON, err := yaml.YAMLToJSON([]byte(inputYAML))
			So(err, ShouldBeNil)

			err = Schema.PartValidator("WorkflowConfig").Validate(bytes.NewReader(inputJSON))
			So(err, ShouldBeNil)

			var cfg WorkflowConfig
			err = json.Unmarshal([]byte(inputJSON), &cfg)
			So(err, ShouldBeNil)

			var input interface{}
			err = json.Unmarshal([]byte(inputJSON), &input)
			So(err, ShouldBeNil)

			encodedCfg, err := json.Marshal(cfg)
			So(err, ShouldBeNil)

			encodedInput, err := json.Marshal(input)
			So(err, ShouldBeNil)

			err = Schema.PartValidator("WorkflowConfig").Validate(bytes.NewReader(encodedCfg))
			So(err, ShouldBeNil)

			So(string(encodedInput), ShouldEqualJSON, string(encodedCfg))
		}

		test(`
login_flows:
- id: login_flow
  steps:
  - type: identify
    id: my_step
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
