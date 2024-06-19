package authflowv2

import (
	htmltemplate "html/template"
	"net/http"
	"time"

	handlerwebapp "github.com/authgear/authgear-server/pkg/auth/handler/webapp"
	"github.com/authgear/authgear-server/pkg/auth/handler/webapp/viewmodels"
	"github.com/authgear/authgear-server/pkg/auth/webapp"
	"github.com/authgear/authgear-server/pkg/lib/authenticationflow/declarative"
	"github.com/authgear/authgear-server/pkg/lib/infra/mail"
	"github.com/authgear/authgear-server/pkg/util/clock"
	"github.com/authgear/authgear-server/pkg/util/httproute"
	"github.com/authgear/authgear-server/pkg/util/template"
)

var TemplateWebAuthflowOOBOTPLinkHTML = template.RegisterHTML(
	"web/authflowv2/oob_otp_link.html",
	handlerwebapp.Components...,
)

func ConfigureV2AuthflowOOBOTPLinkRoute(route httproute.Route) httproute.Route {
	return route.
		WithMethods("OPTIONS", "POST", "GET").
		WithPathPattern(AuthflowV2RouteOOBOTPLink)
}

type AuthflowOOBOTPLinkViewModel struct {
	WebsocketURL     htmltemplate.URL
	StateToken       string
	StateQuery       handlerwebapp.LoginLinkOTPPageQueryState
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

	stateQuery := handlerwebapp.LoginLinkOTPPageQueryStateInitial
	if canCheck {
		stateQuery = handlerwebapp.LoginLinkOTPPageQueryStateMatched
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

func NewInlinePreviewAuthflowOOBOTPLinkViewModel() AuthflowOOBOTPLinkViewModel {
	maskedClaimValue := mail.MaskAddress(viewmodels.PreviewDummyEmail)
	return AuthflowOOBOTPLinkViewModel{
		WebsocketURL:     "",
		StateToken:       "",
		StateQuery:       handlerwebapp.LoginLinkOTPPageQueryStateInitial,
		MaskedClaimValue: maskedClaimValue,
		ResendCooldown:   0,
	}
}

type AuthflowV2OOBOTPLinkHandler struct {
	Controller    *handlerwebapp.AuthflowController
	BaseViewModel *viewmodels.BaseViewModeler
	Renderer      handlerwebapp.Renderer
	Clock         clock.Clock
}

func (h *AuthflowV2OOBOTPLinkHandler) GetData(w http.ResponseWriter, r *http.Request, s *webapp.Session, screen *webapp.AuthflowScreenWithFlowResponse) (map[string]interface{}, error) {
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

func (h *AuthflowV2OOBOTPLinkHandler) GetInlinePreviewData(w http.ResponseWriter, r *http.Request) (map[string]interface{}, error) {
	data := map[string]interface{}{}

	baseViewModel := h.BaseViewModel.ViewModelForInlinePreviewAuthFlow(r, w)
	viewmodels.Embed(data, baseViewModel)

	screenViewModel := NewInlinePreviewAuthflowOOBOTPLinkViewModel()
	viewmodels.Embed(data, screenViewModel)

	branchViewModel := viewmodels.NewInlinePreviewAuthflowBranchViewModel()
	viewmodels.Embed(data, branchViewModel)

	return data, nil
}

func (h *AuthflowV2OOBOTPLinkHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	var handlers handlerwebapp.AuthflowControllerHandlers
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

		result.WriteResponse(w, r)
		return nil
	})
	handlers.PostAction("check", func(s *webapp.Session, screen *webapp.AuthflowScreenWithFlowResponse) error {
		requestDeviceToken := r.Form.Get("x_device_token") == "true"

		input := map[string]interface{}{
			"check":                true,
			"request_device_token": requestDeviceToken,
		}

		result, err := h.Controller.AdvanceWithInput(r, s, screen, input, nil)
		if err != nil {
			return err
		}

		result.WriteResponse(w, r)
		return nil
	})
	handlers.InlinePreview(func(w http.ResponseWriter, r *http.Request) error {
		data, err := h.GetInlinePreviewData(w, r)
		if err != nil {
			return err
		}
		h.Renderer.RenderHTML(w, r, TemplateWebAuthflowOOBOTPLinkHTML, data)
		return nil
	})

	if webapp.IsPreviewModeInline(r) {
		h.Controller.HandleInlinePreview(w, r, &handlers)
		return
	}
	h.Controller.HandleStep(w, r, &handlers)
}
