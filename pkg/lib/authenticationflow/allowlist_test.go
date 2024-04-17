package authenticationflow

import (
	"testing"

	"github.com/authgear/authgear-server/pkg/lib/config"
	. "github.com/smartystreets/goconvey/convey"
)

func TestFlowAllowlist(t *testing.T) {
	Convey("Given a FlowAllowlist", t, func() {
		Convey("When initialized with groups", func() {
			allowlist := &config.AuthenticationFlowAllowlist{
				Groups: &[]string{"group1", "group2"},
			}
			allGroups := []config.UIAuthenticationFlowGroup{
				{
					Name:                 "group1",
					SignupFlows:          []string{"signup1"},
					PromoteFlows:         []string{"promote1"},
					LoginFlows:           []string{"login1"},
					SignupLoginFlows:     []string{"signuplogin1"},
					ReauthFlows:          []string{"reauth1"},
					AccountRecoveryFlows: []string{"recovery1"},
				},
				{
					Name:                 "group2",
					SignupFlows:          []string{"signup2"},
					PromoteFlows:         []string{"promote2"},
					LoginFlows:           []string{"login2"},
					SignupLoginFlows:     []string{"signuplogin2"},
					ReauthFlows:          []string{"reauth2"},
					AccountRecoveryFlows: []string{"recovery2"},
				},
			}

			result := NewFlowAllowlist(allowlist, allGroups)

			Convey("Then the result should contain all flows from the groups in the allowlist", func() {
				So(result.CanCreateFlow(FlowReference{Type: FlowTypeSignup, Name: "signup1"}), ShouldBeTrue)
				So(result.CanCreateFlow(FlowReference{Type: FlowTypeSignup, Name: "signup2"}), ShouldBeTrue)
				So(result.CanCreateFlow(FlowReference{Type: FlowTypePromote, Name: "promote1"}), ShouldBeTrue)
				So(result.CanCreateFlow(FlowReference{Type: FlowTypePromote, Name: "promote2"}), ShouldBeTrue)
				So(result.CanCreateFlow(FlowReference{Type: FlowTypeLogin, Name: "login1"}), ShouldBeTrue)
				So(result.CanCreateFlow(FlowReference{Type: FlowTypeLogin, Name: "login2"}), ShouldBeTrue)
				So(result.CanCreateFlow(FlowReference{Type: FlowTypeSignupLogin, Name: "signuplogin1"}), ShouldBeTrue)
				So(result.CanCreateFlow(FlowReference{Type: FlowTypeSignupLogin, Name: "signuplogin2"}), ShouldBeTrue)
				So(result.CanCreateFlow(FlowReference{Type: FlowTypeReauth, Name: "reauth1"}), ShouldBeTrue)
				So(result.CanCreateFlow(FlowReference{Type: FlowTypeReauth, Name: "reauth2"}), ShouldBeTrue)
				So(result.CanCreateFlow(FlowReference{Type: FlowTypeAccountRecovery, Name: "recovery1"}), ShouldBeTrue)
				So(result.CanCreateFlow(FlowReference{Type: FlowTypeAccountRecovery, Name: "recovery2"}), ShouldBeTrue)
			})
		})

		Convey("When initialized with flows", func() {
			allowlist := &config.AuthenticationFlowAllowlist{
				Flows: &config.AuthenticationFlowAllowlistFlows{
					{Type: config.AuthenticationFlowAllowlistFlowTypeLogin, Name: "flow1"},
					{Type: config.AuthenticationFlowAllowlistFlowTypeSignup, Name: "flow2"},
				},
			}
			allGroups := []config.UIAuthenticationFlowGroup{}

			result := NewFlowAllowlist(allowlist, allGroups)

			Convey("Then the result should contain all flows in the allowlist", func() {
				So(result.CanCreateFlow(FlowReference{Type: FlowTypeLogin, Name: "flow1"}), ShouldBeTrue)
				So(result.CanCreateFlow(FlowReference{Type: FlowTypeSignup, Name: "flow2"}), ShouldBeTrue)
			})
		})

		Convey("When initialized with empty allowlist", func() {
			allowlist := &config.AuthenticationFlowAllowlist{}
			allGroups := []config.UIAuthenticationFlowGroup{}

			result := NewFlowAllowlist(allowlist, allGroups)

			Convey("Then the result should allow all flows", func() {
				So(result.CanCreateFlow(FlowReference{Type: FlowTypeLogin, Name: "flow1"}), ShouldBeTrue)
				So(result.CanCreateFlow(FlowReference{Type: FlowTypeSignup, Name: "flow2"}), ShouldBeTrue)
			})
		})

		Convey("When initialized with default group", func() {
			allowlist := &config.AuthenticationFlowAllowlist{}
			allGroups := []config.UIAuthenticationFlowGroup{
				{
					Name:                 "default",
					SignupFlows:          []string{"signup1"},
					PromoteFlows:         []string{"promote1"},
					LoginFlows:           []string{"login1"},
					SignupLoginFlows:     []string{"signuplogin1"},
					ReauthFlows:          []string{"reauth1"},
					AccountRecoveryFlows: []string{"recovery1"},
				},
			}

			result := NewFlowAllowlist(allowlist, allGroups)

			Convey("Then the result should contain all flows from the default group", func() {
				So(result.CanCreateFlow(FlowReference{Type: FlowTypeSignup, Name: "signup1"}), ShouldBeTrue)
				So(result.CanCreateFlow(FlowReference{Type: FlowTypePromote, Name: "promote1"}), ShouldBeTrue)
				So(result.CanCreateFlow(FlowReference{Type: FlowTypeLogin, Name: "login1"}), ShouldBeTrue)
				So(result.CanCreateFlow(FlowReference{Type: FlowTypeSignupLogin, Name: "signuplogin1"}), ShouldBeTrue)
				So(result.CanCreateFlow(FlowReference{Type: FlowTypeReauth, Name: "reauth1"}), ShouldBeTrue)
				So(result.CanCreateFlow(FlowReference{Type: FlowTypeAccountRecovery, Name: "recovery1"}), ShouldBeTrue)
			})
		})

		Convey("When initialized with default group and allowlist", func() {
			allowlist := &config.AuthenticationFlowAllowlist{
				Flows: &config.AuthenticationFlowAllowlistFlows{
					{Type: config.AuthenticationFlowAllowlistFlowTypeLogin, Name: "flow1"},
				},
			}
			allGroups := []config.UIAuthenticationFlowGroup{
				{
					Name:                 "default",
					SignupFlows:          []string{"signup1"},
					PromoteFlows:         []string{"promote1"},
					LoginFlows:           []string{"login1"},
					SignupLoginFlows:     []string{"signuplogin1"},
					ReauthFlows:          []string{"reauth1"},
					AccountRecoveryFlows: []string{"recovery1"},
				},
			}

			result := NewFlowAllowlist(allowlist, allGroups)

			Convey("Then the result should contain all flows from the default group and the allowlist", func() {
				So(result.CanCreateFlow(FlowReference{Type: FlowTypeLogin, Name: "flow1"}), ShouldBeTrue)
				So(result.CanCreateFlow(FlowReference{Type: FlowTypeSignup, Name: "signup1"}), ShouldBeTrue)
				So(result.CanCreateFlow(FlowReference{Type: FlowTypePromote, Name: "promote1"}), ShouldBeTrue)
				So(result.CanCreateFlow(FlowReference{Type: FlowTypeLogin, Name: "login1"}), ShouldBeTrue)
				So(result.CanCreateFlow(FlowReference{Type: FlowTypeSignupLogin, Name: "signuplogin1"}), ShouldBeTrue)
				So(result.CanCreateFlow(FlowReference{Type: FlowTypeReauth, Name: "reauth1"}), ShouldBeTrue)
				So(result.CanCreateFlow(FlowReference{Type: FlowTypeAccountRecovery, Name: "recovery1"}), ShouldBeTrue)
			})
		})
	})
}
