package declarative

import (
	"encoding/json"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
	"sigs.k8s.io/yaml"

	"github.com/authgear/authgear-server/pkg/lib/config"
)

func TestGenerateAccountRecoveryFlowConfig(t *testing.T) {
	Convey("GenerateAccountRecoveryFlowConfig", t, func() {
		test := func(expected string) {
			var appConfig config.AppConfig = config.AppConfig{}

			config.PopulateDefaultValues(&appConfig)

			flow := GenerateAccountRecoveryFlowConfig(&appConfig)
			flowJSON, err := json.Marshal(flow)
			So(err, ShouldBeNil)

			expectedJSON, err := yaml.YAMLToJSON([]byte(expected))
			So(err, ShouldBeNil)

			So(string(flowJSON), ShouldEqualJSON, string(expectedJSON))
		}

		test(`
name: default
steps:
- type: identify
  one_of:
  - identification: email
    on_failure: ignore
  - identification: phone
    on_failure: ignore
- type: select_destination
`)
	})
}
