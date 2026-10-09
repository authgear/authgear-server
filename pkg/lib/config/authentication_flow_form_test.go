package config

import (
	"context"
	"testing"

	. "github.com/smartystreets/goconvey/convey"

	"github.com/authgear/authgear-server/pkg/util/validation"
)

func TestIsFillableUserProfilePointer(t *testing.T) {
	Convey("isFillableUserProfilePointer", t, func() {
		up := &UserProfileConfig{
			StandardAttributes: &StandardAttributesConfig{
				AccessControl: []*StandardAttributesAccessControlConfig{
					{
						Pointer: "/nickname",
						AccessControl: &UserProfileAttributesAccessControl{
							EndUser:  AccessControlLevelStringReadwrite,
							Bearer:   AccessControlLevelStringReadonly,
							PortalUI: AccessControlLevelStringReadwrite,
						},
					},
					{
						Pointer: "/gender",
						AccessControl: &UserProfileAttributesAccessControl{
							EndUser:  AccessControlLevelStringReadonly,
							Bearer:   AccessControlLevelStringReadonly,
							PortalUI: AccessControlLevelStringReadwrite,
						},
					},
				},
			},
			CustomAttributes: &CustomAttributesConfig{
				Attributes: []*CustomAttributesAttributeConfig{
					{
						ID: "0", Pointer: "/writable", Type: CustomAttributeTypeString,
						AccessControl: &UserProfileAttributesAccessControl{
							EndUser:  AccessControlLevelStringReadwrite,
							Bearer:   AccessControlLevelStringReadonly,
							PortalUI: AccessControlLevelStringReadwrite,
						},
					},
					{
						ID: "1", Pointer: "/readonly", Type: CustomAttributeTypeString,
						AccessControl: &UserProfileAttributesAccessControl{
							EndUser:  AccessControlLevelStringReadonly,
							Bearer:   AccessControlLevelStringReadonly,
							PortalUI: AccessControlLevelStringReadwrite,
						},
					},
					{ID: "2", Pointer: "/hidden", Type: CustomAttributeTypeString},
				},
			},
		}
		up.StandardAttributes.SetDefaults()
		for _, c := range up.CustomAttributes.Attributes {
			c.SetDefaults()
		}

		test := func(pointer string, expected bool) {
			So(isFillableUserProfilePointer(up, pointer), ShouldEqual, expected)
		}

		// Identity-backed attributes.
		test("/email", false)
		test("/phone_number", false)
		test("/preferred_username", false)
		// Sub-attributes of address.
		test("/address/locality", false)
		// Readwrite by default.
		test("/given_name", true)
		// Hidden by default.
		test("/name", false)
		// Hidden by default, made readwrite.
		test("/nickname", true)
		// Readwrite by default, made readonly.
		test("/gender", false)
		test("/writable", true)
		test("/readonly", false)
		test("/hidden", false)
		test("/unknown", false)
	})
}

func TestValidateFormsInSteps(t *testing.T) {
	Convey("validateFormsInSteps", t, func() {
		up := &UserProfileConfig{
			StandardAttributes: &StandardAttributesConfig{},
			CustomAttributes:   &CustomAttributesConfig{},
		}
		up.StandardAttributes.SetDefaults()

		form := &AuthenticationFlowForm{
			Fields: []*AuthenticationFlowFormField{
				{UserProfile: &AuthenticationFlowFormFieldUserProfile{Pointer: "/given_name"}},
			},
		}
		fillForm := &AuthenticationFlowLoginFlowStep{Type: AuthenticationFlowLoginFlowStepTypeFillForm, Form: form}
		identify := func(steps ...*AuthenticationFlowLoginFlowStep) *AuthenticationFlowLoginFlowStep {
			return &AuthenticationFlowLoginFlowStep{
				Type: AuthenticationFlowLoginFlowStepTypeIdentify,
				OneOf: []*AuthenticationFlowLoginFlowOneOf{
					{Identification: "email", Steps: steps},
				},
			}
		}
		authenticate := func(steps ...*AuthenticationFlowLoginFlowStep) *AuthenticationFlowLoginFlowStep {
			return &AuthenticationFlowLoginFlowStep{
				Type: AuthenticationFlowLoginFlowStepTypeAuthenticate,
				OneOf: []*AuthenticationFlowLoginFlowOneOf{
					{Authentication: "primary_password", Steps: steps},
				},
			}
		}

		locations := func(steps ...*AuthenticationFlowLoginFlowStep) []string {
			flow := &AuthenticationFlowLoginFlow{Steps: steps}
			vctx := &validation.Context{}
			validateFormsInSteps(vctx.Child("steps"), up, flow.GetSteps(), false)
			err := vctx.Error("invalid")
			if err == nil {
				return nil
			}
			var out []string
			for _, e := range err.(*validation.AggregatedError).Errors {
				out = append(out, e.Location)
			}
			return out
		}

		Convey("fill_form before identify", func() {
			So(locations(fillForm, identify()), ShouldResemble, []string{"/steps/0"})
		})

		Convey("fill_form right after identify", func() {
			So(locations(identify(), fillForm, authenticate()), ShouldBeNil)
		})

		Convey("fill_form nested in an identify branch", func() {
			So(locations(identify(fillForm)), ShouldBeNil)
		})

		Convey("fill_form nested in a non-identify branch inherits the state", func() {
			So(locations(authenticate(fillForm), identify()), ShouldResemble, []string{"/steps/0/one_of/0/steps/0"})
			So(locations(identify(), authenticate(fillForm)), ShouldBeNil)
		})

		Convey("a non-fillable attribute and a duplicated pointer", func() {
			bad := &AuthenticationFlowLoginFlowStep{
				Type: AuthenticationFlowLoginFlowStepTypeFillForm,
				Form: &AuthenticationFlowForm{
					Fields: []*AuthenticationFlowFormField{
						{UserProfile: &AuthenticationFlowFormFieldUserProfile{Pointer: "/email"}},
						{UserProfile: &AuthenticationFlowFormFieldUserProfile{Pointer: "/given_name"}},
						{UserProfile: &AuthenticationFlowFormFieldUserProfile{Pointer: "/given_name"}},
					},
				},
			}
			So(locations(identify(), bad), ShouldResemble, []string{
				"/steps/1/form/fields/0/user_profile/pointer",
				"/steps/1/form/fields/2/user_profile/pointer",
			})
		})
	})
}

func TestParseAuthenticationFlowForm(t *testing.T) {
	Convey("a form is nil when it is not configured", t, func() {
		// Parse allocates every nil pointer to a struct unless it is tagged nullable.
		cfg, err := Parse(context.Background(), []byte(`
id: test
http:
  public_origin: http://test
authentication_flow:
  signup_flows:
  - name: custom
    steps:
    - type: identify
      one_of:
      - identification: username
        form:
          fields:
          - user_profile:
              pointer: /given_name
      - identification: email
    - type: view_recovery_code
`))
		So(err, ShouldBeNil)

		step := cfg.AuthenticationFlow.SignupFlows[0].Steps[0]
		So(step.Form, ShouldBeNil)
		So(step.OneOf[0].Form, ShouldNotBeNil)
		So(step.OneOf[0].Form.Fields[0].UserProfile.Pointer, ShouldEqual, "/given_name")
		So(step.OneOf[1].Form, ShouldBeNil)
		So(cfg.AuthenticationFlow.SignupFlows[0].Steps[1].Form, ShouldBeNil)
	})
}
