package webapp

import (
	"net/http"
	"time"

	"github.com/authgear/authgear-server/pkg/auth/handler/webapp/viewmodels"
	"github.com/authgear/authgear-server/pkg/auth/webapp"
	"github.com/authgear/authgear-server/pkg/lib/authenticationflow/declarative"
	"github.com/authgear/authgear-server/pkg/util/clock"
	"github.com/authgear/authgear-server/pkg/util/httproute"
	"github.com/authgear/authgear-server/pkg/util/template"
)

var TemplateWebAuthflowOOBOTPLinkHTML = template.RegisterHTML(
	"web/authflow_oob_otp_link.html",
	components...,
)

func ConfigureAuthflowOOBOTPLinkRoute(route httproute.Route) httproute.Route {
	return route.
		WithMethods("OPTIONS", "POST", "GET").
		WithPathPattern(webapp.AuthflowRouteOOBOTPLink)
}

type AuthflowOOBOTPLinkViewModel struct {
	StateQuery       LoginLinkOTPPageQueryState
	MaskedClaimValue string
	ResendCooldown   int
}

func NewAuthflowOOBOTPLinkViewModel(s *webapp.Session, screen *webapp.AuthflowScreenWithFlowResponse, now time.Time) AuthflowOOBOTPLinkViewModel {
	data := screen.StateTokenFlowResponse.Action.Data.(declarative.NodeVerifyClaimData)
	maskedClaimValue := data.MaskedClaimValue
	resendCooldown := int(data.CanResendAt.Sub(now).Seconds())
	if resendCooldown < 0 {
		resendCooldown = 0
	}

	stateQuery := LoginLinkOTPPageQueryStateInitial
	if data.CanCheck {
		stateQuery = LoginLinkOTPPageQueryStateMatched
	}

	return AuthflowOOBOTPLinkViewModel{
		StateQuery:       stateQuery,
		MaskedClaimValue: maskedClaimValue,
		ResendCooldown:   resendCooldown,
	}
}

type AuthflowOOBOTPLinkHandler struct {
	Controller    *AuthflowController
	BaseViewModel *viewmodels.BaseViewModeler
	Renderer      Renderer
	FlashMessage  FlashMessage
	Clock         clock.Clock
}

func (h *AuthflowOOBOTPLinkHandler) GetData(w http.ResponseWriter, r *http.Request, s *webapp.Session, screen *webapp.AuthflowScreenWithFlowResponse) (map[string]interface{}, error) {
	data := map[string]interface{}{}

	baseViewModel := h.BaseViewModel.ViewModel(r, w)
	viewmodels.Embed(data, baseViewModel)

	now := h.Clock.NowUTC()
	screenViewModel := NewAuthflowOOBOTPLinkViewModel(s, screen, now)
	viewmodels.Embed(data, screenViewModel)

	branchViewModel := viewmodels.NewAuthflowBranchViewModel(screen)
	viewmodels.Embed(data, branchViewModel)

	return data, nil
}

func (h *AuthflowOOBOTPLinkHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	var handlers AuthflowControllerHandlers
	handlers.Get(func(s *webapp.Session, screen *webapp.AuthflowScreenWithFlowResponse) error {
		data, err := h.GetData(w, r, s, screen)
		if err != nil {
			return err
		}

		h.Renderer.RenderHTML(w, r, TemplateWebAuthflowOOBOTPLinkHTML, data)
		return nil
	})
	handlers.PostAction("resend", func(s *webapp.Session, screen *webapp.AuthflowScreenWithFlowResponse) error {
		input := map[string]interface{}{
			"resend": true,
		}

		result, err := h.Controller.UpdateWithInput(r, s, screen, input)
		if err != nil {
			return err
		}

		if !result.IsInteractionErr {
			h.FlashMessage.Flash(w, string(webapp.FlashMessageTypeResendLoginLinkSuccess))
		}
		result.WriteResponse(w, r)
		return nil
	})
	handlers.PostAction("check", func(s *webapp.Session, screen *webapp.AuthflowScreenWithFlowResponse) error {
		requestDeviceToken := r.Form.Get("x_device_token") == "true"

		input := map[string]interface{}{
			"check":                true,
			"request_device_token": requestDeviceToken,
		}

		result, err := h.Controller.AdvanceWithInput(r, s, screen, input)
		if err != nil {
			return err
		}

		result.WriteResponse(w, r)
		return nil
	})
	h.Controller.HandleStep(w, r, &handlers)
}
