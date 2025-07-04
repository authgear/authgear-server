package authenticationflow

import (
	"testing"

	"github.com/iawaknahc/jsonschema/pkg/jsonpointer"
	. "github.com/smartystreets/goconvey/convey"

	"github.com/authgear/authgear-server/pkg/api/model"
	"github.com/authgear/authgear-server/pkg/lib/config"
)

var jp = jsonpointer.MustParse

var fixtureSignupFlow *config.AuthenticationFlowSignupFlow = &config.AuthenticationFlowSignupFlow{
	Name: "id",
	Steps: []*config.AuthenticationFlowSignupFlowStep{
		{
			Name: "step0",
			Type: config.AuthenticationFlowSignupFlowStepTypeIdentify,
			OneOf: []*config.AuthenticationFlowSignupFlowOneOf{
				{
					Identification: model.AuthenticationFlowIdentificationEmail,
				},
			},
		},
		{
			Type:       config.AuthenticationFlowSignupFlowStepTypeVerify,
			TargetStep: "step0",
		},
		{
			Type: config.AuthenticationFlowSignupFlowStepTypeCreateAuthenticator,
			OneOf: []*config.AuthenticationFlowSignupFlowOneOf{
				{
					Authentication: model.AuthenticationFlowAuthenticationPrimaryPassword,
					Steps: []*config.AuthenticationFlowSignupFlowStep{
						{
							Type: config.AuthenticationFlowSignupFlowStepTypeCreateAuthenticator,
							OneOf: []*config.AuthenticationFlowSignupFlowOneOf{
								{
									Authentication: model.AuthenticationFlowAuthenticationSecondaryTOTP,
								},
							},
						},
					},
				},
			},
		},
		{
			Type: config.AuthenticationFlowSignupFlowStepTypeFillInUserProfile,
			UserProfile: []*config.AuthenticationFlowSignupFlowUserProfile{
				{
					Pointer:  "/given_name",
					Required: true,
				},
			},
		},
	},
}

var fixtureLoginFlow *config.AuthenticationFlowLoginFlow = &config.AuthenticationFlowLoginFlow{
	Name: "id",
	Steps: []*config.AuthenticationFlowLoginFlowStep{
		{
			Name: "step0",
			Type: config.AuthenticationFlowLoginFlowStepTypeIdentify,
			OneOf: []*config.AuthenticationFlowLoginFlowOneOf{
				{
					Identification: model.AuthenticationFlowIdentificationEmail,
				},
			},
		},
		{
			Type: config.AuthenticationFlowLoginFlowStepTypeAuthenticate,
			OneOf: []*config.AuthenticationFlowLoginFlowOneOf{
				{
					Authentication: model.AuthenticationFlowAuthenticationPrimaryPassword,
					Steps: []*config.AuthenticationFlowLoginFlowStep{
						{
							Type: config.AuthenticationFlowLoginFlowStepTypeAuthenticate,
							OneOf: []*config.AuthenticationFlowLoginFlowOneOf{
								{
									Authentication: model.AuthenticationFlowAuthenticationSecondaryTOTP,
								},
							},
						},
					},
				},
			},
		},
	},
}

var fixtureSignupLoginFlow *config.AuthenticationFlowSignupLoginFlow = &config.AuthenticationFlowSignupLoginFlow{
	Name: "id",
	Steps: []*config.AuthenticationFlowSignupLoginFlowStep{
		{
			Type: config.AuthenticationFlowSignupLoginFlowStepTypeIdentify,
			OneOf: []*config.AuthenticationFlowSignupLoginFlowOneOf{
				{
					Identification: model.AuthenticationFlowIdentificationEmail,
				},
			},
		},
	},
}

var fixtureReauthFlow *config.AuthenticationFlowReauthFlow = &config.AuthenticationFlowReauthFlow{
	Name: "id",
	Steps: []*config.AuthenticationFlowReauthFlowStep{
		{
			Type: config.AuthenticationFlowReauthFlowStepTypeAuthenticate,
			OneOf: []*config.AuthenticationFlowReauthFlowOneOf{
				{
					Authentication: model.AuthenticationFlowAuthenticationPrimaryPassword,
					Steps: []*config.AuthenticationFlowReauthFlowStep{
						{
							Type: config.AuthenticationFlowReauthFlowStepTypeAuthenticate,
							OneOf: []*config.AuthenticationFlowReauthFlowOneOf{
								{
									Authentication: model.AuthenticationFlowAuthenticationSecondaryTOTP,
								},
							},
						},
					},
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
		})
	})
}

func TestJSONPointerForStep(t *testing.T) {
	Convey("JSONPointerForStep", t, func() {
		test := func(p jsonpointer.T, index int, expected string) {
			actual := JSONPointerForStep(p, index)
			So(actual.String(), ShouldEqual, expected)
		}

		test(nil, 1, "/steps/1")
		test(jsonpointer.MustParse("/a"), 1, "/a/steps/1")
	})
}

func TestJSONPointerForOneOf(t *testing.T) {
	Convey("JSONPointerForOneOf", t, func() {
		test := func(p jsonpointer.T, index int, expected string) {
			actual := JSONPointerForOneOf(p, index)
			So(actual.String(), ShouldEqual, expected)
		}

		test(nil, 1, "/one_of/1")
		test(jsonpointer.MustParse("/a"), 1, "/a/one_of/1")
	})
}

func TestJSONPointerToParent(t *testing.T) {
	Convey("JSONPointerToParent", t, func() {
		test := func(p jsonpointer.T, expected string) {
			actual := JSONPointerToParent(p)
			So(actual.String(), ShouldEqual, expected)
		}

		Convey("should panic for length 0", func() {
			So(func() {
				JSONPointerToParent(nil)
			}, ShouldPanic)
		})
		Convey("should panic for length 1", func() {
			So(func() {
				JSONPointerToParent(jsonpointer.MustParse("/a"))
			}, ShouldPanic)
		})

		Convey("should work for length 2", func() {
			test(jsonpointer.MustParse("/steps/0"), "")
			test(jsonpointer.MustParse("/one_of/0"), "")
		})

		Convey("should work for length 4", func() {
			test(jsonpointer.MustParse("/steps/0/one_of/1"), "/steps/0")
			test(jsonpointer.MustParse("/one_of/0/steps/1"), "/one_of/0")
		})
	})
}

func TestJSONPointerSubtract(t *testing.T) {
	Convey("JSONPointerSubtract", t, func() {
		test := func(p1 jsonpointer.T, p2 jsonpointer.T, expected string) {
			actual := JSONPointerSubtract(p1, p2)
			So(actual.String(), ShouldEqual, expected)
		}

		test(jsonpointer.MustParse("/steps/0"), jsonpointer.MustParse("/steps/0"), "")
		test(jsonpointer.MustParse("/steps/0"), jsonpointer.MustParse("/steps/0/steps/1"), "")
		test(jsonpointer.MustParse("/steps/0/one_of/1/steps/2"), jsonpointer.MustParse("/steps/0"), "/one_of/1/steps/2")
		test(jsonpointer.MustParse("/steps/0/one_of/1/steps/2"), jsonpointer.MustParse("/steps/0/one_of/3"), "/1/steps/2")
		test(jsonpointer.MustParse("/steps/0/one_of/3"), jsonpointer.MustParse("/steps/0/one_of/1/steps/2"), "/3")
	})
}
