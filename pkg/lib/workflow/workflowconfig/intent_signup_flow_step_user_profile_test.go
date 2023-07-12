package workflowconfig

import (
	"testing"

	"github.com/iawaknahc/jsonschema/pkg/jsonpointer"
	. "github.com/smartystreets/goconvey/convey"

	"github.com/authgear/authgear-server/pkg/api/apierrors"
	"github.com/authgear/authgear-server/pkg/lib/config"
)

func TestIntentSignupFlowStepUserProfileValidate(t *testing.T) {
	i := &IntentSignupFlowStepUserProfile{}
	f := i.validate

	Convey("IntentSignupFlowStepUserProfile.validate", t, func() {
		test := func(step *config.WorkflowSignupFlowStep, attributes []InputFillUserProfileAttribute, expected error) {
			actual := f(step, attributes)
			if expected == nil {
				So(actual, ShouldBeNil)
			} else {
				expectedAPIErr := apierrors.AsAPIError(expected)
				actualAPIErr := apierrors.AsAPIError(actual)
				So(actualAPIErr.Info, ShouldResemble, expectedAPIErr.Info)
			}
		}

		test(&config.WorkflowSignupFlowStep{
			UserProfile: []*config.WorkflowSignupFlowUserProfile{
				{
					Pointer:  "/given_name",
					Required: true,
				},
				{
					Pointer:  "/family_name",
					Required: false,
				},
			},
		}, []InputFillUserProfileAttribute{
			{
				Pointer: jsonpointer.MustParse("/given_name"),
				Value:   "john",
			},
		}, nil)

		test(&config.WorkflowSignupFlowStep{
			UserProfile: []*config.WorkflowSignupFlowUserProfile{
				{
					Pointer:  "/given_name",
					Required: true,
				},
				{
					Pointer:  "/family_name",
					Required: false,
				},
			},
		}, []InputFillUserProfileAttribute{
			{
				Pointer: jsonpointer.MustParse("/family_name"),
				Value:   "doe",
			},
			{
				Pointer: jsonpointer.MustParse("/middle_name"),
				Value:   "mid",
			},
		}, InvalidUserProfile.NewWithInfo("invalid attributes", apierrors.Details{
			"allowed":    []string{"/given_name", "/family_name"},
			"required":   []string{"/given_name"},
			"actual":     []string{"/family_name", "/middle_name"},
			"missing":    []string{"/given_name"},
			"disallowed": []string{"/middle_name"},
		}))
	})
}
