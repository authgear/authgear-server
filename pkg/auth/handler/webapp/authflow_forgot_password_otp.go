package webapp

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	"strconv"
	"time"

	"github.com/authgear/authgear-server/pkg/auth/handler/webapp/viewmodels"
	"github.com/authgear/authgear-server/pkg/auth/webapp"
	authflow "github.com/authgear/authgear-server/pkg/lib/authenticationflow"
	"github.com/authgear/authgear-server/pkg/lib/authenticationflow/declarative"
	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/util/clock"
	"github.com/authgear/authgear-server/pkg/util/httproute"
	"github.com/authgear/authgear-server/pkg/util/template"
	"github.com/authgear/authgear-server/pkg/util/validation"
)

var TemplateWebAuthflowForgotPasswordGenericOTPHTML = template.RegisterHTML(
	"web/authflow_forgot_password_generic_otp.html",
	Components...,
)

var TemplateWebAuthflowForgotPasswordWhatsappOTPHTML = template.RegisterHTML(
	"web/authflow_forgot_password_whatsapp_otp.html",
	Components...,
)

var AuthflowForgotPasswordOTPSchema = validation.NewSimpleSchema(`
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

func ConfigureAuthflowForgotPasswordOTPRoute(route httproute.Route) httproute.Route {
	return route.
		WithMethods("OPTIONS", "POST", "GET").
		WithPathPattern(webapp.AuthflowRouteForgotPasswordOTP)
}

type AuthflowForgotPasswordOTPAlternativeChannel struct {
	Index   int
	Channel declarative.AccountRecoveryChannel
}

type AuthflowForgotPasswordOTPViewModel struct {
	Channel                        declarative.AccountRecoveryChannel
	MaskedClaimValue               string
	CodeLength                     int
	FailedAttemptRateLimitExceeded bool
	ResendCooldown                 int
	AlternativeChannels            []AuthflowForgotPasswordOTPAlternativeChannel
}

func NewAuthflowForgotPasswordOTPViewModel(
	ctx context.Context,
	c *AuthflowController,
	s *webapp.Session,
	screen *webapp.AuthflowScreenWithFlowResponse,
	now time.Time) (*AuthflowForgotPasswordOTPViewModel, error) {
	data := screen.StateTokenFlowResponse.Action.Data.(declarative.IntentAccountRecoveryFlowStepVerifyAccountRecoveryCodeData)
	channel := data.Channel
	maskedClaimValue := data.MaskedDisplayName
	codeLength := data.CodeLength
	failedAttemptRateLimitExceeded := data.FailedAttemptRateLimitExceeded
	resendCooldown := int(data.CanResendAt.Sub(now).Seconds())
	if resendCooldown < 0 {
		resendCooldown = 0
	}

	vm := &AuthflowForgotPasswordOTPViewModel{
		Channel:                        channel,
		MaskedClaimValue:               maskedClaimValue,
		CodeLength:                     codeLength,
		FailedAttemptRateLimitExceeded: failedAttemptRateLimitExceeded,
		ResendCooldown:                 resendCooldown,
	}

	prevScreen, err := c.GetScreen(ctx, s, screen.Screen.PreviousXStep)
	if err != nil && !errors.Is(err, authflow.ErrFlowNotFound) {
		return nil, err
	}
	if errors.Is(err, authflow.ErrFlowNotFound) {
		return vm, nil
	}
	d := prevScreen.StateTokenFlowResponse.Action.
		Data.(declarative.IntentAccountRecoveryFlowStepSelectDestinationData)

	var alternativeChannels []AuthflowForgotPasswordOTPAlternativeChannel
	for idx, o := range d.Options {
		if o.Channel == channel {
			continue
		}
		if o.MaskedDisplayName != maskedClaimValue {
			continue
		}
		if o.OTPForm != declarative.AccountRecoveryOTPFormCode {
			continue
		}
		alternativeChannels = append(alternativeChannels, AuthflowForgotPasswordOTPAlternativeChannel{
			Channel: o.Channel,
			Index:   idx,
		})
	}
	vm.AlternativeChannels = alternativeChannels

	return vm, nil
}

type AuthflowForgotPasswordOTPHandler struct {
	Controller    *AuthflowController
	BaseViewModel *viewmodels.BaseViewModeler
	Renderer      Renderer
	FlashMessage  FlashMessage
	Clock         clock.Clock
}

func (h *AuthflowForgotPasswordOTPHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	var handlers AuthflowControllerHandlers
	handlers.Get(func(ctx context.Context, s *webapp.Session, screen *webapp.AuthflowScreenWithFlowResponse) error {
		now := h.Clock.NowUTC()
		data := make(map[string]interface{})

		baseViewModel := h.BaseViewModel.ViewModelForAuthFlow(r, w)
		viewmodels.Embed(data, baseViewModel)

		screenViewModel, err := NewAuthflowForgotPasswordOTPViewModel(ctx, h.Controller, s, screen, now)
		if err != nil {
			return err
		}
		viewmodels.Embed(data, screenViewModel)

		if screenViewModel.Channel == declarative.AccountRecoveryChannelWhatsapp {
			h.Renderer.RenderHTML(w, r, TemplateWebAuthflowForgotPasswordWhatsappOTPHTML, data)
		} else {

			h.Renderer.RenderHTML(w, r, TemplateWebAuthflowForgotPasswordGenericOTPHTML, data)
		}
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
	handlers.PostAction("select_channel", func(ctx context.Context, s *webapp.Session, screen *webapp.AuthflowScreenWithFlowResponse) error {
		xIndex, err := strconv.Atoi(r.Form.Get("x_index"))
		if err != nil {
			return err
		}

		input := map[string]interface{}{
			"index": xIndex,
		}

		// prevScreen should be select_destination
		prevScreen, err := h.Controller.GetScreen(ctx, s, screen.Screen.PreviousXStep)
		if err != nil {
			return err
		}
		if string(prevScreen.StateTokenFlowResponse.Action.Type) != string(config.AuthenticationFlowStepTypeSelectDestination) {
			return fmt.Errorf("authflow webapp: unexpected previous step")
		}

		result, err := h.Controller.AdvanceWithInput(ctx, r, s, prevScreen, input, nil)
		if err != nil {
			return err
		}

		result.WriteResponse(w, r)
		return nil
	})
	handlers.PostAction("submit", func(ctx context.Context, s *webapp.Session, screen *webapp.AuthflowScreenWithFlowResponse) error {
		err := AuthflowForgotPasswordOTPSchema.Validator().ValidateValue(FormToJSON(r.Form))
		if err != nil {
			return err
		}

		code := r.Form.Get("x_code")

		input := map[string]interface{}{
			"account_recovery_code": code,
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
