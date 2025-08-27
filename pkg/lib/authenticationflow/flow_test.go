package authenticationflow

import (
	"encoding/json"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
	"sigs.k8s.io/yaml"

	"github.com/authgear/authgear-server/pkg/lib/config"
)

func TestFindStepByType(t *testing.T) {
	Convey("TestFindStepByType", t, func() {
		test := func(flowYAML string, stepType config.AuthenticationFlowStepType, expectedPointer string) {
			var flowRoot config.AuthenticationFlowAccountRecoveryFlow
			jsonData, err := yaml.YAMLToJSON([]byte(flowYAML))
			So(err, ShouldBeNil)
			err = json.Unmarshal(jsonData, &flowRoot)
			So(err, ShouldBeNil)

			ptr, found := FindStepByType(&flowRoot, stepType)
			if expectedPointer == "" {
				So(found, ShouldBeFalse)
			} else {
				So(found, ShouldBeTrue)
				So(ptr.String(), ShouldEqual, expectedPointer)
			}
		}

		flowYAML := `name: default
steps:
- type: identify
  one_of:
  - identification: email
    on_failure: ignore
    steps:
    - type: select_destination
      allowed_channels:
      - channel: email
        otp_form: code
  - identification: phone
    on_failure: ignore
    steps:
    - type: select_destination
      allowed_channels:
      - channel: sms
        otp_form: link
- type: verify_account_recovery_code
- type: reset_password
`

		Convey("should find identify step", func() {
			test(flowYAML, config.AuthenticationFlowStepTypeIdentify, "/steps/0")
		})

		Convey("should find select_destination step", func() {
			test(flowYAML, config.AuthenticationFlowStepTypeSelectDestination, "/steps/0/one_of/0/steps/0")
		})

		Convey("should find verify_account_recovery_code step", func() {
			test(flowYAML, config.AuthenticationFlowStepTypeVerifyAccountRecoveryCode, "/steps/1")
		})

		Convey("should find reset_password step", func() {
			test(flowYAML, config.AuthenticationFlowStepTypeResetPassword, "/steps/2")
		})

		Convey("should not find non-existent step", func() {
			test(flowYAML, "non_existent_step", "")
		})
	})
}
