package workflowconfig

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"

	"github.com/authgear/authgear-server/pkg/api/apierrors"
	"github.com/authgear/authgear-server/pkg/lib/authn/attrs"
	"github.com/authgear/authgear-server/pkg/lib/config"
)

func TestIntentSignupFlowStepUserProfileValidate(t *testing.T) {
	i := &IntentSignupFlowStepUserProfile{}
	f := i.validate

	Convey("IntentSignupFlowStepUserProfile.validate", t, func() {
		test := func(step *config.WorkflowSignupFlowStep, attributes []attrs.T, expectedResult []string, expectedErr error) {
			actualResult, actualErr := f(step, attributes)
			if expectedErr == nil {
				So(actualResult, ShouldResemble, expectedResult)
			} else {
				expectedAPIErr := apierrors.AsAPIError(expectedErr)
				actualAPIErr := apierrors.AsAPIError(actualErr)
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
		}, []attrs.T{
			{
				Pointer: "/given_name",
				Value:   "john",
			},
		}, []string{"/family_name"}, nil)

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
		}, []attrs.T{
			{
				Pointer: "/family_name",
				Value:   "doe",
			},
			{
				Pointer: "/middle_name",
				Value:   "mid",
			},
		}, nil, InvalidUserProfile.NewWithInfo("invalid attributes", apierrors.Details{
			"allowed":    []string{"/given_name", "/family_name"},
			"required":   []string{"/given_name"},
			"present":    []string{"/family_name", "/middle_name"},
			"absent":     []string{"/given_name"},
			"missing":    []string{"/given_name"},
			"disallowed": []string{"/middle_name"},
		}))
	})
}
