package webapp

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"

	"github.com/authgear/authgear-server/pkg/api/model"
	authflow "github.com/authgear/authgear-server/pkg/lib/authenticationflow"
	"github.com/authgear/authgear-server/pkg/lib/authenticationflow/declarative"
	"github.com/authgear/authgear-server/pkg/lib/config"
)

func TestAuthflowScreenWithFlowResponseDefaultBranchIndex(t *testing.T) {
	Convey("AuthflowScreenWithFlowResponse.DefaultBranchIndex", t, func() {
		makeScreen := func(flowType authflow.FlowType, stepType config.AuthenticationFlowStepType, data authflow.Data) *AuthflowScreenWithFlowResponse {
			return &AuthflowScreenWithFlowResponse{
				StateTokenFlowResponse: &authflow.FlowResponse{
					Type: flowType,
					Action: &authflow.FlowAction{
						Type: authflow.FlowActionTypeFromStepType(stepType),
						Data: data,
					},
				},
			}
		}
		makeData := func(authentications ...model.AuthenticationFlowAuthentication) declarative.StepAuthenticateData {
			options := []declarative.AuthenticateOptionForOutput{}
			for _, a := range authentications {
				options = append(options, declarative.AuthenticateOptionForOutput{Authentication: a})
			}
			return declarative.StepAuthenticateData{Options: options}
		}

		Convey("prefer passkey in login authenticate step", func() {
			s := makeScreen(authflow.FlowTypeLogin, config.AuthenticationFlowStepTypeAuthenticate, makeData(
				model.AuthenticationFlowAuthenticationPrimaryPassword,
				model.AuthenticationFlowAuthenticationPrimaryOOBOTPEmail,
				model.AuthenticationFlowAuthenticationPrimaryPasskey,
			))
			So(s.DefaultBranchIndex(), ShouldEqual, 2)
		})

		Convey("take the first branch if the user has no passkey", func() {
			s := makeScreen(authflow.FlowTypeLogin, config.AuthenticationFlowStepTypeAuthenticate, makeData(
				model.AuthenticationFlowAuthenticationPrimaryOOBOTPEmail,
				model.AuthenticationFlowAuthenticationPrimaryPassword,
			))
			So(s.DefaultBranchIndex(), ShouldEqual, 0)
		})

		Convey("take the first branch in secondary authentication", func() {
			s := makeScreen(authflow.FlowTypeLogin, config.AuthenticationFlowStepTypeAuthenticate, makeData(
				model.AuthenticationFlowAuthenticationSecondaryTOTP,
				model.AuthenticationFlowAuthenticationSecondaryOOBOTPSMS,
			))
			So(s.DefaultBranchIndex(), ShouldEqual, 0)
		})

		Convey("prefer passkey in reauth authenticate step", func() {
			s := makeScreen(authflow.FlowTypeReauth, config.AuthenticationFlowStepTypeAuthenticate, makeData(
				model.AuthenticationFlowAuthenticationPrimaryPassword,
				model.AuthenticationFlowAuthenticationPrimaryPasskey,
			))
			So(s.DefaultBranchIndex(), ShouldEqual, 1)
		})

		Convey("take the first branch in other flow types", func() {
			s := makeScreen(authflow.FlowTypeSignup, config.AuthenticationFlowStepTypeAuthenticate, makeData(
				model.AuthenticationFlowAuthenticationPrimaryPassword,
				model.AuthenticationFlowAuthenticationPrimaryPasskey,
			))
			So(s.DefaultBranchIndex(), ShouldEqual, 0)
		})

		Convey("take the first branch in create authenticator", func() {
			s := makeScreen(authflow.FlowTypeLogin, config.AuthenticationFlowStepTypeAuthenticate, declarative.CreateAuthenticatorData{
				Options: []declarative.CreateAuthenticatorOptionForOutput{
					{Authentication: model.AuthenticationFlowAuthenticationPrimaryPassword},
				},
			})
			So(s.DefaultBranchIndex(), ShouldEqual, 0)
		})
	})
}
