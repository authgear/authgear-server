package authflowv2

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"

	"github.com/authgear/authgear-server/pkg/api/model"
	"github.com/authgear/authgear-server/pkg/auth/webapp"
	authflow "github.com/authgear/authgear-server/pkg/lib/authenticationflow"
	"github.com/authgear/authgear-server/pkg/lib/authenticationflow/declarative"
)

func TestSelectAccountOptionFromScreen(t *testing.T) {
	Convey("selectAccountOptionFromScreen", t, func() {
		makeScreen := func(data authflow.Data) *webapp.AuthflowScreenWithFlowResponse {
			return &webapp.AuthflowScreenWithFlowResponse{
				StateTokenFlowResponse: &authflow.FlowResponse{
					Action: &authflow.FlowAction{
						Data: data,
					},
				},
			}
		}

		Convey("identify data with a select_account option not at index 0: returns it with its real index", func() {
			screen := makeScreen(declarative.IdentificationData{
				Options: []declarative.IdentificationOption{
					{Identification: model.AuthenticationFlowIdentificationEmail},
					{
						Identification: model.AuthenticationFlowIdentificationSelectAccount,
						DisplayName:    "user@example.com",
					},
				},
			})

			option, index, ok := selectAccountOptionFromScreen(screen)
			So(ok, ShouldBeTrue)
			So(index, ShouldEqual, 1)
			So(option.Identification, ShouldEqual, model.AuthenticationFlowIdentificationSelectAccount)
			So(option.DisplayName, ShouldEqual, "user@example.com")
		})

		Convey("identify data without a select_account option: not found", func() {
			screen := makeScreen(declarative.IdentificationData{
				Options: []declarative.IdentificationOption{
					{Identification: model.AuthenticationFlowIdentificationEmail},
				},
			})

			_, _, ok := selectAccountOptionFromScreen(screen)
			So(ok, ShouldBeFalse)
		})

		Convey("some other action data type entirely: not found, no panic", func() {
			screen := makeScreen(declarative.NewPasswordData{})

			_, _, ok := selectAccountOptionFromScreen(screen)
			So(ok, ShouldBeFalse)
		})
	})
}
