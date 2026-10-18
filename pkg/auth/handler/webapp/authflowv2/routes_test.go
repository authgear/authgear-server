package authflowv2

import (
	"context"
	"net/http"
	"testing"

	. "github.com/smartystreets/goconvey/convey"

	"github.com/authgear/authgear-server/pkg/auth/webapp"
	authflow "github.com/authgear/authgear-server/pkg/lib/authenticationflow"
)

func TestNavigateFillForm(t *testing.T) {
	Convey("fill_form advances to the fill_form page", t, func() {
		n := &AuthflowV2Navigator{}
		r, _ := http.NewRequest("GET", "/", nil)
		makeScreen := func(flowType authflow.FlowType) *webapp.AuthflowScreenWithFlowResponse {
			return &webapp.AuthflowScreenWithFlowResponse{
				Screen: &webapp.AuthflowScreen{StateToken: &webapp.AuthflowStateToken{XStep: "step"}},
				StateTokenFlowResponse: &authflow.FlowResponse{
					Type:   flowType,
					Action: &authflow.FlowAction{Type: "fill_form"},
				},
			}
		}

		for _, flowType := range []authflow.FlowType{authflow.FlowTypeSignup, authflow.FlowTypePromote, authflow.FlowTypeLogin} {
			s := makeScreen(flowType)
			result := &webapp.Result{}
			n.Navigate(context.Background(), s, r, "", result)
			So(result.RedirectURI, ShouldEqual, AuthflowV2RouteFillForm+"?x_step=step")
			So(result.NavigationAction, ShouldEqual, webapp.NavigationActionAdvance)
		}
	})
}
