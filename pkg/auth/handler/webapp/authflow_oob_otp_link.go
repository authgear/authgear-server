package webapp

import (
	"context"
	htmltemplate "html/template"
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
	Components...,
)

func ConfigureAuthflowOOBOTPLinkRoute(route httproute.Route) httproute.Route {
	return route.
		WithMethods("OPTIONS", "POST", "GET").
		WithPathPattern(webapp.AuthflowRouteOOBOTPLink)
}

type AuthflowOOBOTPLinkViewModel struct {
	WebsocketURL     htmltemplate.URL
	StateToken       string
	StateQuery       LoginLinkOTPPageQueryState
	MaskedClaimValue string
	ResendCooldown   int
}

func NewAuthflowOOBOTPLinkViewModel(s *webapp.Session, screen *webapp.AuthflowScreenWithFlowResponse, now time.Time) AuthflowOOBOTPLinkViewModel {
	var maskedClaimValue string
	var resendCooldown int
	var canCheck bool
	var websocketURL string

	switch data := screen.StateTokenFlowResponse.Action.Data.(type) {
	case declarative.VerifyOOBOTPData:
		maskedClaimValue = data.MaskedClaimValue
		resendCooldown = int(data.CanResendAt.Sub(now).Seconds())
		canCheck = data.CanCheck
		websocketURL = data.WebsocketURL
	default:
		panic("authflowv2: unexpected action data")
	}
	if resendCooldown < 0 {
		resendCooldown = 0
	}

	stateQuery := LoginLinkOTPPageQueryStateInitial
	if canCheck {
		stateQuery = LoginLinkOTPPageQueryStateMatched
	}

	return AuthflowOOBOTPLinkViewModel{
		// nolint: gosec
		WebsocketURL:     htmltemplate.URL(websocketURL),
		StateToken:       screen.Screen.StateToken.StateToken,
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

	baseViewModel := h.BaseViewModel.ViewModelForAuthFlow(r, w)
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
	handlers.Get(func(ctx context.Context, s *webapp.Session, screen *webapp.AuthflowScreenWithFlowResponse) error {
		data, err := h.GetData(w, r, s, screen)
		if err != nil {
			return err
		}

		h.Renderer.RenderHTML(w, r, TemplateWebAuthflowOOBOTPLinkHTML, data)
		return nil
	})
	handlers.PostAction("resend", func(ctx context.Context, s *webapp.Session, screen *webapp.AuthflowScreenWithFlowResponse) error {
		input := map[string]interface{}{
			"resend": true,
		}

		result, err := h.Controller.UpdateWithInput(ctx, r, s, screen, input)
		if err != nil {
			return err
		}

		h.FlashMessage.Flash(w, string(webapp.FlashMessageTypeResendLoginLinkSuccess))
		result.WriteResponse(w, r)
		return nil
	})
	handlers.PostAction("check", func(ctx context.Context, s *webapp.Session, screen *webapp.AuthflowScreenWithFlowResponse) error {
		requestDeviceToken := r.Form.Get("x_device_token") == "true"

		input := map[string]interface{}{
			"check":                true,
			"request_device_token": requestDeviceToken,
		}

		result, err := h.Controller.AdvanceWithInput(ctx, r, s, screen, input, nil)
		if err != nil {
			return err
		}

		result.WriteResponse(w, r)
		return nil
	})
	h.Controller.HandleStep(r.Context(), w, r, &handlers)
}
