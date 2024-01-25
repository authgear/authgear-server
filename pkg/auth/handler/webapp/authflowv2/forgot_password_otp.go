package authflowv2

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"time"

	handlerwebapp "github.com/authgear/authgear-server/pkg/auth/handler/webapp"
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

var TemplateWebAuthflowForgotPasswordOTPHTML = template.RegisterHTML(
	"web/authflowv2/forgot_password_otp.html",
	handlerwebapp.Components...,
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

func ConfigureAuthflowV2ForgotPasswordOTPRoute(route httproute.Route) httproute.Route {
	return route.
		WithMethods("OPTIONS", "POST", "GET").
		WithPathPattern(AuthflowV2RouteForgotPasswordOTP)
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
	c *handlerwebapp.AuthflowController,
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

	prevScreen, err := c.GetScreen(s, screen.Screen.PreviousXStep)
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

type AuthflowV2ForgotPasswordOTPHandler struct {
	Controller    *handlerwebapp.AuthflowController
	BaseViewModel *viewmodels.BaseViewModeler
	Renderer      handlerwebapp.Renderer
	FlashMessage  handlerwebapp.FlashMessage
	Clock         clock.Clock
}

func (h *AuthflowV2ForgotPasswordOTPHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	var handlers handlerwebapp.AuthflowControllerHandlers
	handlers.Get(func(s *webapp.Session, screen *webapp.AuthflowScreenWithFlowResponse) error {
		now := h.Clock.NowUTC()
		data := make(map[string]interface{})

		baseViewModel := h.BaseViewModel.ViewModelForAuthFlow(r, w)
		viewmodels.Embed(data, baseViewModel)

		screenViewModel, err := NewAuthflowForgotPasswordOTPViewModel(h.Controller, s, screen, now)
		if err != nil {
			return err
		}
		viewmodels.Embed(data, screenViewModel)

		h.Renderer.RenderHTML(w, r, TemplateWebAuthflowForgotPasswordOTPHTML, data)
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
	handlers.PostAction("select_channel", func(s *webapp.Session, screen *webapp.AuthflowScreenWithFlowResponse) error {
		xIndex, err := strconv.Atoi(r.Form.Get("x_index"))
		if err != nil {
			return err
		}

		input := map[string]interface{}{
			"index": xIndex,
		}

		// prevScreen should be select_destination
		prevScreen, err := h.Controller.GetScreen(s, screen.Screen.PreviousXStep)
		if err != nil {
			return err
		}
		if string(prevScreen.StateTokenFlowResponse.Action.Type) != string(config.AuthenticationFlowStepTypeSelectDestination) {
			return fmt.Errorf("authflow webapp: unexpected previous step")
		}

		result, err := h.Controller.AdvanceWithInput(r, s, prevScreen, input, nil)
		if err != nil {
			return err
		}

		result.WriteResponse(w, r)
		return nil
	})
	handlers.PostAction("submit", func(s *webapp.Session, screen *webapp.AuthflowScreenWithFlowResponse) error {
		err := AuthflowForgotPasswordOTPSchema.Validator().ValidateValue(handlerwebapp.FormToJSON(r.Form))
		if err != nil {
			return err
		}

		code := r.Form.Get("x_code")

		input := map[string]interface{}{
			"account_recovery_code": code,
		}

		result, err := h.Controller.AdvanceWithInput(r, s, screen, input, nil)
		if err != nil {
			return err
		}

		result.WriteResponse(w, r)
		return nil
	})
	h.Controller.HandleStep(w, r, &handlers)
}
