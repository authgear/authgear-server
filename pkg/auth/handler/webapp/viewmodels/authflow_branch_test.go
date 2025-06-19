package viewmodels

import (
	"testing"

	"github.com/authgear/authgear-server/pkg/api/model"
	. "github.com/smartystreets/goconvey/convey"
)

func TestAuthflowBranchViewModel(t *testing.T) {
	Convey("reorderBranches", t, func() {
		input := []AuthflowBranch{
			{
				Authentication: model.AuthenticationFlowAuthenticationPrimaryOOBOTPSMS,
				Index:          0,
			},
			{
				Authentication: model.AuthenticationFlowAuthenticationPrimaryPasskey,
				Index:          1,
			},
			{
				Authentication: model.AuthenticationFlowAuthenticationPrimaryOOBOTPEmail,
				Index:          2,
			},
		}
		reordered := reorderBranches(input)

		expected := []AuthflowBranch{
			{
				Authentication: model.AuthenticationFlowAuthenticationPrimaryOOBOTPSMS,
				Index:          0,
			},
			{
				Authentication: model.AuthenticationFlowAuthenticationPrimaryOOBOTPEmail,
				Index:          2,
			},
			{
				Authentication: model.AuthenticationFlowAuthenticationPrimaryPasskey,
				Index:          1,
			},
		}

		So(reordered, ShouldResemble, expected)
	})
}
