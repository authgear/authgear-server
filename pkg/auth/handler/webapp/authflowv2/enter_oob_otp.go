package authflowv2

import (
	"net/http"
	"time"

	"github.com/authgear/authgear-server/pkg/api/model"
	handlerwebapp "github.com/authgear/authgear-server/pkg/auth/handler/webapp"
	"github.com/authgear/authgear-server/pkg/auth/handler/webapp/viewmodels"
	"github.com/authgear/authgear-server/pkg/auth/webapp"
	authflow "github.com/authgear/authgear-server/pkg/lib/authenticationflow"
	"github.com/authgear/authgear-server/pkg/lib/authenticationflow/declarative"
	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/lib/infra/mail"
	"github.com/authgear/authgear-server/pkg/util/clock"
	"github.com/authgear/authgear-server/pkg/util/httproute"
	"github.com/authgear/authgear-server/pkg/util/secretcode"
	"github.com/authgear/authgear-server/pkg/util/template"
	"github.com/authgear/authgear-server/pkg/util/validation"
)

var TemplateWebAuthflowEnterOOBOTPHTML = template.RegisterHTML(
	"web/authflowv2/enter_oob_otp.html",
	handlerwebapp.Components...,
)

var AuthflowEnterOOBOTPSchema = validation.NewSimpleSchema(`
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

func ConfigureAuthflowV2EnterOOBOTPRoute(route httproute.Route) httproute.Route {
	return route.
		WithMethods("OPTIONS", "POST", "GET").
		WithPathPattern(AuthflowV2RouteEnterOOBOTP)
}

type AuthflowEnterOOBOTPViewModel struct {
	FlowActionType                 string
	Channel                        string
	MaskedClaimValue               string
	CodeLength                     int
	FailedAttemptRateLimitExceeded bool
	ResendCooldown                 int
}

func NewAuthflowEnterOOBOTPViewModel(s *webapp.Session, screen *webapp.AuthflowScreenWithFlowResponse, now time.Time) AuthflowEnterOOBOTPViewModel {
	flowActionType := screen.StateTokenFlowResponse.Action.Type

	var channel model.AuthenticatorOOBChannel
	var maskedClaimValue string
	var codeLength int
	var failedAttemptRateLimitExceeded bool
	var resendCooldown int

	switch data := screen.StateTokenFlowResponse.Action.Data.(type) {
	case declarative.VerifyOOBOTPData:
		channel = data.Channel
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

	return AuthflowEnterOOBOTPViewModel{
		FlowActionType:                 string(flowActionType),
		Channel:                        string(channel),
		MaskedClaimValue:               maskedClaimValue,
		CodeLength:                     codeLength,
		FailedAttemptRateLimitExceeded: failedAttemptRateLimitExceeded,
		ResendCooldown:                 resendCooldown,
	}
}

func NewInlinePreviewAuthflowEnterOOBOTPViewModel(
	authenticatorConfig *config.AuthenticatorConfig,
	identityConfig *config.IdentityConfig,
) AuthflowEnterOOBOTPViewModel {
	firstLoginIDTypeForPreview := model.LoginIDKeyTypeEmail
	if len(identityConfig.LoginID.Keys) > 0 {
		firstLoginIDKey := identityConfig.LoginID.Keys[0]
		if firstLoginIDKey.Type != model.LoginIDKeyTypeUsername {
			firstLoginIDTypeForPreview = firstLoginIDKey.Type
		}
	}

	var channelForPreview model.AuthenticatorOOBChannel
	var maskedClaimValue string
	switch firstLoginIDTypeForPreview {
	case model.LoginIDKeyTypeEmail:
		channelForPreview = model.AuthenticatorOOBChannelEmail
		maskedClaimValue = mail.MaskAddress(viewmodels.PreviewDummyEmail)
	case model.LoginIDKeyTypePhone:
		if authenticatorConfig.OOB.SMS.PhoneOTPMode.IsWhatsappEnabled() {
			channelForPreview = model.AuthenticatorOOBChannelWhatsapp
		} else {
			channelForPreview = model.AuthenticatorOOBChannelSMS
		}
		maskedClaimValue = viewmodels.PreviewDummyPhoneNumberMasked
	}

	return AuthflowEnterOOBOTPViewModel{
		FlowActionType:                 string(authflow.FlowTypeLogin),
		Channel:                        string(channelForPreview),
		MaskedClaimValue:               maskedClaimValue,
		CodeLength:                     secretcode.OOBOTPSecretCode.Length(),
		FailedAttemptRateLimitExceeded: false,
		ResendCooldown:                 0,
	}
}

type AuthflowV2EnterOOBOTPHandler struct {
	Controller                             *handlerwebapp.AuthflowController
	BaseViewModel                          *viewmodels.BaseViewModeler
	InlinePreviewAuthflowBranchViewModeler *viewmodels.InlinePreviewAuthflowBranchViewModeler
	Renderer                               handlerwebapp.Renderer
	FlashMessage                           handlerwebapp.FlashMessage
	Clock                                  clock.Clock
	AuthenticatorConfig                    *config.AuthenticatorConfig
	IdentityConfig                         *config.IdentityConfig
}

func (h *AuthflowV2EnterOOBOTPHandler) GetData(w http.ResponseWriter, r *http.Request, s *webapp.Session, screen *webapp.AuthflowScreenWithFlowResponse) (map[string]interface{}, error) {
	now := h.Clock.NowUTC()
	data := make(map[string]interface{})

	baseViewModel := h.BaseViewModel.ViewModelForAuthFlow(r, w)
	viewmodels.Embed(data, baseViewModel)

	screenViewModel := NewAuthflowEnterOOBOTPViewModel(s, screen, now)
	viewmodels.Embed(data, screenViewModel)

	branchViewModel := viewmodels.NewAuthflowBranchViewModel(screen)
	viewmodels.Embed(data, branchViewModel)

	return data, nil
}

func (h *AuthflowV2EnterOOBOTPHandler) GetInlinePreviewData(w http.ResponseWriter, r *http.Request) (map[string]interface{}, error) {
	data := make(map[string]interface{})

	baseViewModel := h.BaseViewModel.ViewModelForInlinePreviewAuthFlow(r, w)
	viewmodels.Embed(data, baseViewModel)

	screenViewModel := NewInlinePreviewAuthflowEnterOOBOTPViewModel(
		h.AuthenticatorConfig,
		h.IdentityConfig,
	)
	viewmodels.Embed(data, screenViewModel)

	branchViewModel := h.InlinePreviewAuthflowBranchViewModeler.NewAuthflowBranchViewModelForInlinePreviewEnterOOBOTP()
	viewmodels.Embed(data, branchViewModel)
	return data, nil
}

func (h *AuthflowV2EnterOOBOTPHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	var handlers handlerwebapp.AuthflowControllerHandlers
	handlers.Get(func(s *webapp.Session, screen *webapp.AuthflowScreenWithFlowResponse) error {
		data, err := h.GetData(w, r, s, screen)
		if err != nil {
			return err
		}

		h.Renderer.RenderHTML(w, r, TemplateWebAuthflowEnterOOBOTPHTML, data)
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

		h.FlashMessage.Flash(w, string(webapp.FlashMessageTypeResendCodeSuccess))
		result.WriteResponse(w, r)
		return nil
	})
	handlers.PostAction("submit", func(s *webapp.Session, screen *webapp.AuthflowScreenWithFlowResponse) error {
		err := AuthflowEnterOOBOTPSchema.Validator().ValidateValue(handlerwebapp.FormToJSON(r.Form))
		if err != nil {
			return err
		}

		code := r.Form.Get("x_code")
		requestDeviceToken := r.Form.Get("x_device_token") == "true"

		input := map[string]interface{}{
			"code":                 code,
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
		h.Renderer.RenderHTML(w, r, TemplateWebAuthflowEnterOOBOTPHTML, data)
		return nil
	})

	if webapp.IsPreviewModeInline(r) {
		h.Controller.HandleInlinePreview(w, r, &handlers)
		return
	}
	h.Controller.HandleStep(w, r, &handlers)
}
