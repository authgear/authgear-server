package declarative

import (
	"bytes"
	"encoding/json"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
	"sigs.k8s.io/yaml"

	"github.com/authgear/authgear-server/pkg/lib/config"
)

func TestGenerateAccountRecoveryFlowConfig(t *testing.T) {
	Convey("GenerateAccountRecoveryFlowConfig", t, func() {
		test := func(cfgStr string, expected string) {

			jsonData, err := yaml.YAMLToJSON([]byte(cfgStr))
			So(err, ShouldBeNil)

			var appConfig config.AppConfig
			decoder := json.NewDecoder(bytes.NewReader(jsonData))
			err = decoder.Decode(&appConfig)
			So(err, ShouldBeNil)

			config.PopulateDefaultValues(&appConfig)

			flow := GenerateAccountRecoveryFlowConfig(&appConfig)
			flowJSON, err := json.Marshal(flow)
			So(err, ShouldBeNil)

			expectedJSON, err := yaml.YAMLToJSON([]byte(expected))
			So(err, ShouldBeNil)

			So(string(flowJSON), ShouldEqualJSON, string(expectedJSON))
		}

		test(
			`
identity:
  login_id:
    keys:
    - type: email
    - type: phone
`,
			`
name: default
steps:
- type: identify
  one_of:
  - identification: email
    on_failure: ignore
  - identification: phone
    on_failure: ignore
- type: select_destination
- type: verify_account_recovery_code
- type: reset_password
`)

		test(
			`
identity:
  login_id:
    keys:
      - type: phone
`,
			`
name: default
steps:
- type: identify
  one_of:
  - identification: phone
    on_failure: ignore
- type: select_destination
- type: verify_account_recovery_code
- type: reset_password
`)

		test(
			`
identity:
  login_id:
    keys:
      - type: email
`,
			`
name: default
steps:
- type: identify
  one_of:
  - identification: email
    on_failure: ignore
- type: select_destination
- type: verify_account_recovery_code
- type: reset_password
`)
	})
}
