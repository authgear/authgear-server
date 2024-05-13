package viewmodels

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"

	"github.com/authgear/authgear-server/pkg/lib/config"
)

func TestAuthflowBranchViewModel(t *testing.T) {
	Convey("reorderBranches", t, func() {
		input := []AuthflowBranch{
			{
				Authentication: config.AuthenticationFlowAuthenticationPrimaryOOBOTPSMS,
				Index:          0,
			},
			{
				Authentication: config.AuthenticationFlowAuthenticationPrimaryPasskey,
				Index:          1,
			},
			{
				Authentication: config.AuthenticationFlowAuthenticationPrimaryOOBOTPEmail,
				Index:          2,
			},
		}
		reordered := reorderBranches(input)

		expected := []AuthflowBranch{
			{
				Authentication: config.AuthenticationFlowAuthenticationPrimaryOOBOTPSMS,
				Index:          0,
			},
			{
				Authentication: config.AuthenticationFlowAuthenticationPrimaryOOBOTPEmail,
				Index:          2,
			},
			{
				Authentication: config.AuthenticationFlowAuthenticationPrimaryPasskey,
				Index:          1,
			},
		}

		So(reordered, ShouldResemble, expected)
	})
}
