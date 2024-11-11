package webapp

import (
	"context"
	"net/http"

	"time"

	"github.com/authgear/authgear-server/pkg/auth/handler/webapp/viewmodels"
	"github.com/authgear/authgear-server/pkg/auth/webapp"
	"github.com/authgear/authgear-server/pkg/lib/authenticationflow/declarative"
	"github.com/authgear/authgear-server/pkg/util/clock"
	"github.com/authgear/authgear-server/pkg/util/httproute"
	"github.com/authgear/authgear-server/pkg/util/template"
	"github.com/authgear/authgear-server/pkg/util/validation"
)

var TemplateWebAuthflowWhatsappOTPHTML = template.RegisterHTML(
	"web/authflow_whatsapp_otp.html",
	Components...,
)

var AuthflowWhatsappOTPSchema = validation.NewSimpleSchema(`
	{
		"type": "object",
		"properties": {
			"x_code": {
				"type": "string",
				"format": "x_oob_otp_code"
			}
		},
		"required": ["x_code"]
	}
`)

func ConfigureAuthflowWhatsappOTPRoute(route httproute.Route) httproute.Route {
	return route.
		WithMethods("OPTIONS", "POST", "GET").
		WithPathPattern(webapp.AuthflowRouteWhatsappOTP)
}

type AuthflowWhatsappOTPViewModel struct {
	MaskedClaimValue               string
	CodeLength                     int
	FailedAttemptRateLimitExceeded bool
	ResendCooldown                 int
}

func NewAuthflowWhatsappOTPViewModel(s *webapp.Session, screen *webapp.AuthflowScreenWithFlowResponse, now time.Time) AuthflowWhatsappOTPViewModel {
	var maskedClaimValue string
	var codeLength int
	var failedAttemptRateLimitExceeded bool
	var resendCooldown int

	switch data := screen.StateTokenFlowResponse.Action.Data.(type) {
	case declarative.VerifyOOBOTPData:
		maskedClaimValue = data.MaskedClaimValue
		codeLength = data.CodeLength
		failedAttemptRateLimitExceeded = data.FailedAttemptRateLimitExceeded
		resendCooldown = int(data.CanResendAt.Sub(now).Seconds())
	default:
		panic("authflowv2: unexpected action data")
	}
	if resendCooldown < 0 {
		resendCooldown = 0
	}

	return AuthflowWhatsappOTPViewModel{
		MaskedClaimValue:               maskedClaimValue,
		CodeLength:                     codeLength,
		FailedAttemptRateLimitExceeded: failedAttemptRateLimitExceeded,
		ResendCooldown:                 resendCooldown,
	}
}

type AuthflowWhatsappOTPHandler struct {
	Controller    *AuthflowController
	BaseViewModel *viewmodels.BaseViewModeler
	Renderer      Renderer
	FlashMessage  FlashMessage
	Clock         clock.Clock
}

func (h *AuthflowWhatsappOTPHandler) GetData(w http.ResponseWriter, r *http.Request, s *webapp.Session, screen *webapp.AuthflowScreenWithFlowResponse) (map[string]interface{}, error) {
	now := h.Clock.NowUTC()
	data := make(map[string]interface{})

	baseViewModel := h.BaseViewModel.ViewModelForAuthFlow(r, w)
	viewmodels.Embed(data, baseViewModel)

	screenViewModel := NewAuthflowWhatsappOTPViewModel(s, screen, now)
	viewmodels.Embed(data, screenViewModel)

	branchViewModel := viewmodels.NewAuthflowBranchViewModel(screen)
	viewmodels.Embed(data, branchViewModel)

	return data, nil
}

func (h *AuthflowWhatsappOTPHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	var handlers AuthflowControllerHandlers
	handlers.Get(func(ctx context.Context, s *webapp.Session, screen *webapp.AuthflowScreenWithFlowResponse) error {
		data, err := h.GetData(w, r, s, screen)
		if err != nil {
			return err
		}

		h.Renderer.RenderHTML(w, r, TemplateWebAuthflowWhatsappOTPHTML, data)
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

		h.FlashMessage.Flash(w, string(webapp.FlashMessageTypeResendCodeSuccess))
		result.WriteResponse(w, r)
		return nil
	})
	handlers.PostAction("submit", func(ctx context.Context, s *webapp.Session, screen *webapp.AuthflowScreenWithFlowResponse) error {
		err := AuthflowWhatsappOTPSchema.Validator().ValidateValue(FormToJSON(r.Form))
		if err != nil {
			return err
		}

		code := r.Form.Get("x_code")
		requestDeviceToken := r.Form.Get("x_device_token") == "true"

		input := map[string]interface{}{
			"code":                 code,
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
