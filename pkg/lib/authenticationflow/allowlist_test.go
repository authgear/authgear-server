package authenticationflow

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"

	"github.com/authgear/authgear-server/pkg/lib/config"
)

func TestFlowAllowlist(t *testing.T) {
	Convey("Given a FlowAllowlist", t, func() {
		Convey("When initialized with groups", func() {
			allowlist := &config.AuthenticationFlowAllowlist{
				Groups: []*config.AuthenticationFlowAllowlistGroup{
					{Name: "group1"},
					{Name: "group2"},
				},
			}
			allGroups := []*config.UIAuthenticationFlowGroup{
				{
					Name: "group1",
					Flows: []*config.UIAuthenticationFlowGroupFlow{
						{Type: config.AuthenticationFlowTypeSignup, Name: "signup1"},
						{Type: config.AuthenticationFlowTypePromote, Name: "promote1"},
						{Type: config.AuthenticationFlowTypeLogin, Name: "login1"},
						{Type: config.AuthenticationFlowTypeSignupLogin, Name: "signuplogin1"},
						{Type: config.AuthenticationFlowTypeReauth, Name: "reauth1"},
						{Type: config.AuthenticationFlowTypeAccountRecovery, Name: "recovery1"},
					},
				},
				{
					Name: "group2",
					Flows: []*config.UIAuthenticationFlowGroupFlow{
						{Type: config.AuthenticationFlowTypeSignup, Name: "signup2"},
						{Type: config.AuthenticationFlowTypePromote, Name: "promote2"},
						{Type: config.AuthenticationFlowTypeLogin, Name: "login2"},
						{Type: config.AuthenticationFlowTypeSignupLogin, Name: "signuplogin2"},
						{Type: config.AuthenticationFlowTypeReauth, Name: "reauth2"},
						{Type: config.AuthenticationFlowTypeAccountRecovery, Name: "recovery2"},
					},
				},
			}

			result := NewFlowAllowlist(allowlist, allGroups)

			Convey("Then the result should contain all flows from the group allowlist", func() {
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

				So(result.CanCreateFlow(FlowReference{Type: FlowTypeSignup, Name: "default"}), ShouldBeFalse)
				So(result.CanCreateFlow(FlowReference{Type: FlowTypePromote, Name: "default"}), ShouldBeFalse)
				So(result.CanCreateFlow(FlowReference{Type: FlowTypeLogin, Name: "default"}), ShouldBeFalse)
				So(result.CanCreateFlow(FlowReference{Type: FlowTypeSignupLogin, Name: "default"}), ShouldBeFalse)
				So(result.CanCreateFlow(FlowReference{Type: FlowTypeReauth, Name: "default"}), ShouldBeFalse)
				So(result.CanCreateFlow(FlowReference{Type: FlowTypeAccountRecovery, Name: "default"}), ShouldBeFalse)
			})
		})

		Convey("When initialized with flows", func() {
			allowlist := &config.AuthenticationFlowAllowlist{
				Flows: []*config.AuthenticationFlowAllowlistFlow{
					{Type: config.AuthenticationFlowTypeLogin, Name: "flow1"},
					{Type: config.AuthenticationFlowTypeSignup, Name: "flow2"},
				},
			}
			allGroups := []*config.UIAuthenticationFlowGroup{}

			result := NewFlowAllowlist(allowlist, allGroups)

			Convey("Then the result should contain all flows in the flow allowlist", func() {
				So(result.CanCreateFlow(FlowReference{Type: FlowTypeLogin, Name: "flow1"}), ShouldBeTrue)
				So(result.CanCreateFlow(FlowReference{Type: FlowTypeSignup, Name: "flow2"}), ShouldBeTrue)

				So(result.CanCreateFlow(FlowReference{Type: FlowTypeLogin, Name: "default"}), ShouldBeFalse)
				So(result.CanCreateFlow(FlowReference{Type: FlowTypeSignup, Name: "default"}), ShouldBeFalse)
			})
		})

		Convey("When initialized with empty allowlist", func() {
			allowlist := &config.AuthenticationFlowAllowlist{}
			allGroups := []*config.UIAuthenticationFlowGroup{}

			result := NewFlowAllowlist(allowlist, allGroups)

			Convey("Then the result should allow all flows", func() {
				So(result.CanCreateFlow(FlowReference{Type: FlowTypeLogin, Name: "flow1"}), ShouldBeTrue)
				So(result.CanCreateFlow(FlowReference{Type: FlowTypeSignup, Name: "flow2"}), ShouldBeTrue)

				So(result.CanCreateFlow(FlowReference{Type: FlowTypeSignup, Name: "default"}), ShouldBeTrue)
				So(result.CanCreateFlow(FlowReference{Type: FlowTypePromote, Name: "default"}), ShouldBeTrue)
				So(result.CanCreateFlow(FlowReference{Type: FlowTypeLogin, Name: "default"}), ShouldBeTrue)
				So(result.CanCreateFlow(FlowReference{Type: FlowTypeSignupLogin, Name: "default"}), ShouldBeTrue)
				So(result.CanCreateFlow(FlowReference{Type: FlowTypeReauth, Name: "default"}), ShouldBeTrue)
				So(result.CanCreateFlow(FlowReference{Type: FlowTypeAccountRecovery, Name: "default"}), ShouldBeTrue)
			})
		})

		Convey("When initialized with defined default group", func() {
			allowlist := &config.AuthenticationFlowAllowlist{}
			allGroups := []*config.UIAuthenticationFlowGroup{
				{
					Name: "default",
					Flows: []*config.UIAuthenticationFlowGroupFlow{
						{Type: config.AuthenticationFlowTypeSignup, Name: "signup1"},
						{Type: config.AuthenticationFlowTypePromote, Name: "promote1"},
						{Type: config.AuthenticationFlowTypeLogin, Name: "login1"},
						{Type: config.AuthenticationFlowTypeSignupLogin, Name: "signuplogin1"},
						{Type: config.AuthenticationFlowTypeReauth, Name: "reauth1"},
						{Type: config.AuthenticationFlowTypeAccountRecovery, Name: "recovery1"},
					},
				},
			}

			result := NewFlowAllowlist(allowlist, allGroups)

			Convey("Then the result should contain all flows from the default group, as well as default flows", func() {
				So(result.CanCreateFlow(FlowReference{Type: FlowTypeSignup, Name: "signup1"}), ShouldBeTrue)
				So(result.CanCreateFlow(FlowReference{Type: FlowTypePromote, Name: "promote1"}), ShouldBeTrue)
				So(result.CanCreateFlow(FlowReference{Type: FlowTypeLogin, Name: "login1"}), ShouldBeTrue)
				So(result.CanCreateFlow(FlowReference{Type: FlowTypeSignupLogin, Name: "signuplogin1"}), ShouldBeTrue)
				So(result.CanCreateFlow(FlowReference{Type: FlowTypeReauth, Name: "reauth1"}), ShouldBeTrue)
				So(result.CanCreateFlow(FlowReference{Type: FlowTypeAccountRecovery, Name: "recovery1"}), ShouldBeTrue)

				So(result.CanCreateFlow(FlowReference{Type: FlowTypeSignup, Name: "default"}), ShouldBeTrue)
				So(result.CanCreateFlow(FlowReference{Type: FlowTypePromote, Name: "default"}), ShouldBeTrue)
				So(result.CanCreateFlow(FlowReference{Type: FlowTypeLogin, Name: "default"}), ShouldBeTrue)
				So(result.CanCreateFlow(FlowReference{Type: FlowTypeSignupLogin, Name: "default"}), ShouldBeTrue)
				So(result.CanCreateFlow(FlowReference{Type: FlowTypeReauth, Name: "default"}), ShouldBeTrue)
				So(result.CanCreateFlow(FlowReference{Type: FlowTypeAccountRecovery, Name: "default"}), ShouldBeTrue)
			})
		})

		Convey("When initialized with defined default group and a group allowlist", func() {
			allowlist := &config.AuthenticationFlowAllowlist{
				Groups: []*config.AuthenticationFlowAllowlistGroup{
					{Name: "default"},
				},
			}
			allGroups := []*config.UIAuthenticationFlowGroup{
				{
					Name: "default",
					Flows: []*config.UIAuthenticationFlowGroupFlow{
						{Type: config.AuthenticationFlowTypeSignup, Name: "signup1"},
						{Type: config.AuthenticationFlowTypePromote, Name: "promote1"},
						{Type: config.AuthenticationFlowTypeLogin, Name: "login1"},
						{Type: config.AuthenticationFlowTypeSignupLogin, Name: "signuplogin1"},
						{Type: config.AuthenticationFlowTypeReauth, Name: "reauth1"},
						{Type: config.AuthenticationFlowTypeAccountRecovery, Name: "recovery1"},
					},
				},
			}

			result := NewFlowAllowlist(allowlist, allGroups)

			Convey("Then the result should contain all flows from the default group, but not the default flows", func() {
				So(result.CanCreateFlow(FlowReference{Type: FlowTypeSignup, Name: "signup1"}), ShouldBeTrue)
				So(result.CanCreateFlow(FlowReference{Type: FlowTypePromote, Name: "promote1"}), ShouldBeTrue)
				So(result.CanCreateFlow(FlowReference{Type: FlowTypeLogin, Name: "login1"}), ShouldBeTrue)
				So(result.CanCreateFlow(FlowReference{Type: FlowTypeSignupLogin, Name: "signuplogin1"}), ShouldBeTrue)
				So(result.CanCreateFlow(FlowReference{Type: FlowTypeReauth, Name: "reauth1"}), ShouldBeTrue)
				So(result.CanCreateFlow(FlowReference{Type: FlowTypeAccountRecovery, Name: "recovery1"}), ShouldBeTrue)

				So(result.CanCreateFlow(FlowReference{Type: FlowTypeSignup, Name: "default"}), ShouldBeFalse)
				So(result.CanCreateFlow(FlowReference{Type: FlowTypePromote, Name: "default"}), ShouldBeFalse)
				So(result.CanCreateFlow(FlowReference{Type: FlowTypeLogin, Name: "default"}), ShouldBeFalse)
				So(result.CanCreateFlow(FlowReference{Type: FlowTypeSignupLogin, Name: "default"}), ShouldBeFalse)
				So(result.CanCreateFlow(FlowReference{Type: FlowTypeReauth, Name: "default"}), ShouldBeFalse)
				So(result.CanCreateFlow(FlowReference{Type: FlowTypeAccountRecovery, Name: "default"}), ShouldBeFalse)
			})
		})

		Convey("When initialized with no group allowlist and a flow allowlist", func() {
			allowlist := &config.AuthenticationFlowAllowlist{
				Flows: []*config.AuthenticationFlowAllowlistFlow{
					{Type: config.AuthenticationFlowTypeLogin, Name: "flow1"},
				},
			}
			allGroups := []*config.UIAuthenticationFlowGroup{
				{
					Name: "default",
					Flows: []*config.UIAuthenticationFlowGroupFlow{
						{Type: config.AuthenticationFlowTypeSignup, Name: "signup1"},
						{Type: config.AuthenticationFlowTypePromote, Name: "promote1"},
						{Type: config.AuthenticationFlowTypeLogin, Name: "login1"},
						{Type: config.AuthenticationFlowTypeSignupLogin, Name: "signuplogin1"},
						{Type: config.AuthenticationFlowTypeReauth, Name: "reauth1"},
						{Type: config.AuthenticationFlowTypeAccountRecovery, Name: "recovery1"},
					},
				},
			}

			result := NewFlowAllowlist(allowlist, allGroups)

			Convey("Then the result should contain all flows the flow allowlist", func() {
				So(result.CanCreateFlow(FlowReference{Type: FlowTypeLogin, Name: "flow1"}), ShouldBeTrue)

				So(result.CanCreateFlow(FlowReference{Type: FlowTypeSignup, Name: "signup1"}), ShouldBeFalse)
				So(result.CanCreateFlow(FlowReference{Type: FlowTypePromote, Name: "promote1"}), ShouldBeFalse)
				So(result.CanCreateFlow(FlowReference{Type: FlowTypeLogin, Name: "login1"}), ShouldBeFalse)
				So(result.CanCreateFlow(FlowReference{Type: FlowTypeSignupLogin, Name: "signuplogin1"}), ShouldBeFalse)
				So(result.CanCreateFlow(FlowReference{Type: FlowTypeReauth, Name: "reauth1"}), ShouldBeFalse)
				So(result.CanCreateFlow(FlowReference{Type: FlowTypeAccountRecovery, Name: "recovery1"}), ShouldBeFalse)

				So(result.CanCreateFlow(FlowReference{Type: FlowTypeSignup, Name: "default"}), ShouldBeFalse)
				So(result.CanCreateFlow(FlowReference{Type: FlowTypePromote, Name: "default"}), ShouldBeFalse)
				So(result.CanCreateFlow(FlowReference{Type: FlowTypeLogin, Name: "default"}), ShouldBeFalse)
				So(result.CanCreateFlow(FlowReference{Type: FlowTypeSignupLogin, Name: "default"}), ShouldBeFalse)
				So(result.CanCreateFlow(FlowReference{Type: FlowTypeReauth, Name: "default"}), ShouldBeFalse)
				So(result.CanCreateFlow(FlowReference{Type: FlowTypeAccountRecovery, Name: "default"}), ShouldBeFalse)
			})
		})
	})
}

func TestDeriveFlowNameForDefaultUI(t *testing.T) {
	Convey("DeriveFlowNameForDefaultUI", t, func() {
		Convey("flow group is specified; allowlist is specified; allowed", func() {
			clientAllowlist := &config.AuthenticationFlowAllowlist{
				Groups: []*config.AuthenticationFlowAllowlistGroup{
					{
						Name: "group-1",
					},
				},
			}
			definedGroups := []*config.UIAuthenticationFlowGroup{
				{
					Name: "group-1",
					Flows: []*config.UIAuthenticationFlowGroupFlow{
						{
							Type: config.AuthenticationFlowTypeLogin,
							Name: "flow-1",
						},
					},
				},
			}

			Convey("It should return the correct flow name", func() {
				flowName, err := NewFlowAllowlist(clientAllowlist, definedGroups).DeriveFlowNameForDefaultUI(FlowTypeLogin, "group-1")
				So(err, ShouldBeNil)
				So(flowName, ShouldEqual, "flow-1")
			})

			Convey("It should return error for undefined flow", func() {
				_, err := NewFlowAllowlist(clientAllowlist, definedGroups).DeriveFlowNameForDefaultUI(FlowTypeSignup, "group-1")
				So(err, ShouldBeError)
			})
		})

		Convey("flow group is specified; allowlist is specified; disallowed", func() {
			clientAllowlist := &config.AuthenticationFlowAllowlist{
				Groups: []*config.AuthenticationFlowAllowlistGroup{
					{
						Name: "group-2",
					},
				},
			}
			definedGroups := []*config.UIAuthenticationFlowGroup{
				{
					Name: "group-1",
					Flows: []*config.UIAuthenticationFlowGroupFlow{
						{
							Type: config.AuthenticationFlowTypeSignup,
							Name: "flow-1",
						},
					},
				},
			}

			Convey("It should return an error", func() {
				_, err := NewFlowAllowlist(clientAllowlist, definedGroups).DeriveFlowNameForDefaultUI(FlowTypeSignup, "group-1")
				So(err, ShouldEqual, ErrFlowNotAllowed)
			})
		})

		Convey("flow group is unspecified; allowlist is specified", func() {
			clientAllowlist := &config.AuthenticationFlowAllowlist{
				Groups: []*config.AuthenticationFlowAllowlistGroup{
					{
						Name: "group-1",
					},
				},
			}
			definedGroups := []*config.UIAuthenticationFlowGroup{
				{
					Name: "group-1",
					Flows: []*config.UIAuthenticationFlowGroupFlow{
						{
							Type: config.AuthenticationFlowTypeLogin,
							Name: "flow-1",
						},
					},
				},
			}

			Convey("It should return the most appropriate flow name and no error", func() {
				flowName, err := NewFlowAllowlist(clientAllowlist, definedGroups).DeriveFlowNameForDefaultUI(FlowTypeLogin, "")
				So(err, ShouldBeNil)
				So(flowName, ShouldEqual, "flow-1")
			})
		})

		Convey("flow group is specified as 'default'; allowlist is unspecified", func() {
			clientAllowlist := &config.AuthenticationFlowAllowlist{}
			definedGroups := []*config.UIAuthenticationFlowGroup{
				{
					Name: "group-1",
					Flows: []*config.UIAuthenticationFlowGroupFlow{
						{
							Type: config.AuthenticationFlowTypeLogin,
							Name: "flow-1",
						},
					},
				},
			}

			Convey("it should return the flow in the default group", func() {
				flowName, err := NewFlowAllowlist(clientAllowlist, definedGroups).DeriveFlowNameForDefaultUI(FlowTypeLogin, "default")
				So(err, ShouldBeNil)
				So(flowName, ShouldEqual, "default")
			})
		})

		Convey("flow group is specified; allowlist is unspecified", func() {
			clientAllowlist := &config.AuthenticationFlowAllowlist{}
			definedGroups := []*config.UIAuthenticationFlowGroup{
				{
					Name: "group-1",
					Flows: []*config.UIAuthenticationFlowGroupFlow{
						{
							Type: config.AuthenticationFlowTypeLogin,
							Name: "flow-1",
						},
					},
				},
			}

			Convey("it should return the flow in the specified group", func() {
				flowName, err := NewFlowAllowlist(clientAllowlist, definedGroups).DeriveFlowNameForDefaultUI(FlowTypeLogin, "group-1")
				So(err, ShouldBeNil)
				So(flowName, ShouldEqual, "flow-1")
			})
		})

		Convey("flow group is unspecified; allowlist is unspecified", func() {
			clientAllowlist := &config.AuthenticationFlowAllowlist{}
			definedGroups := []*config.UIAuthenticationFlowGroup{
				{
					Name: "group-1",
					Flows: []*config.UIAuthenticationFlowGroupFlow{
						{
							Type: config.AuthenticationFlowTypeLogin,
							Name: "flow-1",
						},
					},
				},
			}

			Convey("It should return the most appropriate flow name and no error", func() {
				flowName, err := NewFlowAllowlist(clientAllowlist, definedGroups).DeriveFlowNameForDefaultUI(FlowTypeLogin, "")
				So(err, ShouldBeNil)
				So(flowName, ShouldEqual, "default")
			})
		})

	})
}
