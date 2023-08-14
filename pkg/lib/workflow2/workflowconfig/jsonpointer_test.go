package workflowconfig

import (
	"testing"

	"github.com/iawaknahc/jsonschema/pkg/jsonpointer"
	. "github.com/smartystreets/goconvey/convey"

	"github.com/authgear/authgear-server/pkg/lib/config"
)

var jp = jsonpointer.MustParse

var fixtureSignupFlow *config.WorkflowSignupFlow = &config.WorkflowSignupFlow{
	ID: "id",
	Steps: []*config.WorkflowSignupFlowStep{
		{
			ID:   "step0",
			Type: config.WorkflowSignupFlowStepTypeIdentify,
			OneOf: []*config.WorkflowSignupFlowOneOf{
				{
					Identification: config.WorkflowIdentificationMethodEmail,
				},
			},
		},
		{
			Type:       config.WorkflowSignupFlowStepTypeVerify,
			TargetStep: "step0",
		},
		{
			Type: config.WorkflowSignupFlowStepTypeAuthenticate,
			OneOf: []*config.WorkflowSignupFlowOneOf{
				{
					Authentication: config.WorkflowAuthenticationMethodPrimaryPassword,
					Steps: []*config.WorkflowSignupFlowStep{
						{
							Type: config.WorkflowSignupFlowStepTypeAuthenticate,
							OneOf: []*config.WorkflowSignupFlowOneOf{
								{
									Authentication: config.WorkflowAuthenticationMethodSecondaryTOTP,
								},
							},
						},
					},
				},
				{
					Authentication: config.WorkflowAuthenticationMethodPrimaryPasskey,
				},
			},
		},
		{
			Type: config.WorkflowSignupFlowStepTypeUserProfile,
			UserProfile: []*config.WorkflowSignupFlowUserProfile{
				{
					Pointer:  "/given_name",
					Required: true,
				},
			},
		},
	},
}

var fixtureLoginFlow *config.WorkflowLoginFlow = &config.WorkflowLoginFlow{
	ID: "id",
	Steps: []*config.WorkflowLoginFlowStep{
		{
			ID:   "step0",
			Type: config.WorkflowLoginFlowStepTypeIdentify,
			OneOf: []*config.WorkflowLoginFlowOneOf{
				{
					Identification: config.WorkflowIdentificationMethodEmail,
				},
			},
		},
		{
			Type: config.WorkflowLoginFlowStepTypeAuthenticate,
			OneOf: []*config.WorkflowLoginFlowOneOf{
				{
					Authentication: config.WorkflowAuthenticationMethodPrimaryPassword,
					Steps: []*config.WorkflowLoginFlowStep{
						{
							Type: config.WorkflowLoginFlowStepTypeAuthenticate,
							OneOf: []*config.WorkflowLoginFlowOneOf{
								{
									Authentication: config.WorkflowAuthenticationMethodSecondaryTOTP,
								},
							},
						},
					},
				},
				{
					Authentication: config.WorkflowAuthenticationMethodPrimaryPasskey,
				},
			},
		},
	},
}

var fixtureSignupLoginFlow *config.WorkflowSignupLoginFlow = &config.WorkflowSignupLoginFlow{
	ID: "id",
	Steps: []*config.WorkflowSignupLoginFlowStep{
		{
			Type: config.WorkflowSignupLoginFlowStepTypeIdentify,
			OneOf: []*config.WorkflowSignupLoginFlowOneOf{
				{
					Identification: config.WorkflowIdentificationMethodEmail,
				},
			},
		},
	},
}

var fixtureReauthFlow *config.WorkflowReauthFlow = &config.WorkflowReauthFlow{
	ID: "id",
	Steps: []*config.WorkflowReauthFlowStep{
		{
			Type: config.WorkflowReauthFlowStepTypeAuthenticate,
			OneOf: []*config.WorkflowReauthFlowOneOf{
				{
					Authentication: config.WorkflowAuthenticationMethodPrimaryPassword,
					Steps: []*config.WorkflowReauthFlowStep{
						{
							Type: config.WorkflowReauthFlowStepTypeAuthenticate,
							OneOf: []*config.WorkflowReauthFlowOneOf{
								{
									Authentication: config.WorkflowAuthenticationMethodSecondaryTOTP,
								},
							},
						},
					},
				},
				{
					Authentication: config.WorkflowAuthenticationMethodPrimaryPasskey,
				},
			},
		},
	},
}

func TestGetCurrentObject(t *testing.T) {
	Convey("GetCurrentObject", t, func() {
		Convey("SignupFlow", func() {
			test := func(pointer jsonpointer.T, expected any) {
				entries, err := Traverse(fixtureSignupFlow, pointer)
				So(err, ShouldBeNil)

				actual, err := GetCurrentObject(entries)
				So(err, ShouldBeNil)

				So(actual, ShouldResemble, expected)
			}

			test(jp(""), fixtureSignupFlow)
			test(jp("/steps/0"), fixtureSignupFlow.Steps[0])
			test(jp("/steps/0/one_of/0"), fixtureSignupFlow.Steps[0].OneOf[0])
			test(jp("/steps/1"), fixtureSignupFlow.Steps[1])
			test(jp("/steps/2"), fixtureSignupFlow.Steps[2])
			test(jp("/steps/2/one_of/0"), fixtureSignupFlow.Steps[2].OneOf[0])
			test(jp("/steps/2/one_of/0/steps/0"), fixtureSignupFlow.Steps[2].OneOf[0].Steps[0])
			test(jp("/steps/2/one_of/0/steps/0/one_of/0"), fixtureSignupFlow.Steps[2].OneOf[0].Steps[0].OneOf[0])
			test(jp("/steps/2/one_of/1"), fixtureSignupFlow.Steps[2].OneOf[1])
			test(jp("/steps/3"), fixtureSignupFlow.Steps[3])
		})

		Convey("LoginFlow", func() {
			test := func(pointer jsonpointer.T, expected any) {
				entries, err := Traverse(fixtureLoginFlow, pointer)
				So(err, ShouldBeNil)

				actual, err := GetCurrentObject(entries)
				So(err, ShouldBeNil)

				So(actual, ShouldResemble, expected)
			}

			test(jp(""), fixtureLoginFlow)
			test(jp("/steps/0"), fixtureLoginFlow.Steps[0])
			test(jp("/steps/0/one_of/0"), fixtureLoginFlow.Steps[0].OneOf[0])
			test(jp("/steps/1"), fixtureLoginFlow.Steps[1])
			test(jp("/steps/1/one_of/0"), fixtureLoginFlow.Steps[1].OneOf[0])
			test(jp("/steps/1/one_of/0/steps/0"), fixtureLoginFlow.Steps[1].OneOf[0].Steps[0])
			test(jp("/steps/1/one_of/0/steps/0/one_of/0"), fixtureLoginFlow.Steps[1].OneOf[0].Steps[0].OneOf[0])
			test(jp("/steps/1/one_of/1"), fixtureLoginFlow.Steps[1].OneOf[1])
		})

		Convey("SignupLoginFlow", func() {
			test := func(pointer jsonpointer.T, expected any) {
				entries, err := Traverse(fixtureSignupLoginFlow, pointer)
				So(err, ShouldBeNil)

				actual, err := GetCurrentObject(entries)
				So(err, ShouldBeNil)

				So(actual, ShouldResemble, expected)
			}

			test(jp(""), fixtureSignupLoginFlow)
			test(jp("/steps/0"), fixtureSignupLoginFlow.Steps[0])
			test(jp("/steps/0/one_of/0"), fixtureSignupLoginFlow.Steps[0].OneOf[0])
		})

		Convey("ReauthFlow", func() {
			test := func(pointer jsonpointer.T, expected any) {
				entries, err := Traverse(fixtureReauthFlow, pointer)
				So(err, ShouldBeNil)

				actual, err := GetCurrentObject(entries)
				So(err, ShouldBeNil)

				So(actual, ShouldResemble, expected)
			}

			test(jp(""), fixtureReauthFlow)
			test(jp("/steps/0"), fixtureReauthFlow.Steps[0])
			test(jp("/steps/0/one_of/0"), fixtureReauthFlow.Steps[0].OneOf[0])
			test(jp("/steps/0/one_of/0/steps/0"), fixtureReauthFlow.Steps[0].OneOf[0].Steps[0])
			test(jp("/steps/0/one_of/0/steps/0/one_of/0"), fixtureReauthFlow.Steps[0].OneOf[0].Steps[0].OneOf[0])
			test(jp("/steps/0/one_of/1"), fixtureReauthFlow.Steps[0].OneOf[1])
		})
	})
}
