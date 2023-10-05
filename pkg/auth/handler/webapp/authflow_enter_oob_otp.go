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

var TemplateWebAuthflowEnterOOBOTPHTML = template.RegisterHTML(
	"web/authflow_enter_oob_otp.html",
	components...,
)

func ConfigureAuthflowEnterOOBOTPRoute(route httproute.Route) httproute.Route {
	return route.
		WithMethods("OPTIONS", "POST", "GET").
		WithPathPattern(webapp.AuthflowRouteEnterOOBOTP)
}

type AuthflowEnterOOBOTPViewModel struct {
	FlowActionType                 string
	Channel                        string
	MaskedClaimValue               string
	CodeLength                     int
	FailedAttemptRateLimitExceeded bool
	ResendCooldown                 int
	DeviceTokenEnabled             bool
}

func NewAuthflowEnterOOBOTPViewModel(s *webapp.Session, screen *webapp.AuthflowScreenWithFlowResponse, now time.Time) AuthflowEnterOOBOTPViewModel {
	flowActionType := screen.StateTokenFlowResponse.Action.Type

	data := screen.StateTokenFlowResponse.Action.Data.(declarative.NodeVerifyClaimData)
	channel := data.Channel
	maskedClaimValue := data.MaskedClaimValue
	codeLength := data.CodeLength
	failedAttemptRateLimitExceeded := data.FailedAttemptRateLimitExceeded
	resendCooldown := int(data.CanResendAt.Sub(now).Seconds())
	if resendCooldown < 0 {
		resendCooldown = 0
	}

	deviceTokenEnabled := false
	if screen.BranchStateTokenFlowResponse != nil {
		if data, ok := screen.BranchStateTokenFlowResponse.Action.Data.(declarative.IntentLoginFlowStepAuthenticateData); ok {
			deviceTokenEnabled = data.DeviceTokenEnabled
		}
	}

	return AuthflowEnterOOBOTPViewModel{
		FlowActionType:                 string(flowActionType),
		Channel:                        string(channel),
		MaskedClaimValue:               maskedClaimValue,
		CodeLength:                     codeLength,
		FailedAttemptRateLimitExceeded: failedAttemptRateLimitExceeded,
		ResendCooldown:                 resendCooldown,
		DeviceTokenEnabled:             deviceTokenEnabled,
	}
}

type AuthflowEnterOOBOTPHandler struct {
	Controller    *AuthflowController
	BaseViewModel *viewmodels.BaseViewModeler
	Renderer      Renderer
	Clock         clock.Clock
}

func (h *AuthflowEnterOOBOTPHandler) GetData(w http.ResponseWriter, r *http.Request, s *webapp.Session, screen *webapp.AuthflowScreenWithFlowResponse) (map[string]interface{}, error) {
	now := h.Clock.NowUTC()
	data := make(map[string]interface{})

	baseViewModel := h.BaseViewModel.ViewModel(r, w)
	viewmodels.Embed(data, baseViewModel)

	screenViewModel := NewAuthflowEnterOOBOTPViewModel(s, screen, now)
	viewmodels.Embed(data, screenViewModel)

	return data, nil
}

func (h *AuthflowEnterOOBOTPHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	var handlers AuthflowControllerHandlers
	handlers.Get(func(s *webapp.Session, screen *webapp.AuthflowScreenWithFlowResponse) error {
		data, err := h.GetData(w, r, s, screen)
		if err != nil {
			return err
		}

		h.Renderer.RenderHTML(w, r, TemplateWebAuthflowEnterOOBOTPHTML, data)
		return nil
	})
	handlers.PostAction("resend", func(s *webapp.Session, screen *webapp.AuthflowScreenWithFlowResponse) error {
		return nil
	})
	handlers.PostAction("submit", func(s *webapp.Session, screen *webapp.AuthflowScreenWithFlowResponse) error {
		return nil
	})
	h.Controller.HandleStep(w, r, &handlers)
}
