package webapp

import (
	"fmt"
	"net/http"

	"github.com/authgear/authgear-server/pkg/api/model"
	"github.com/authgear/authgear-server/pkg/auth/handler/webapp/viewmodels"
	"github.com/authgear/authgear-server/pkg/auth/webapp"
	"github.com/authgear/authgear-server/pkg/lib/authn/otp"
	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/lib/infra/mail"
	"github.com/authgear/authgear-server/pkg/lib/interaction"
	"github.com/authgear/authgear-server/pkg/lib/interaction/nodes"
	"github.com/authgear/authgear-server/pkg/util/clock"
	"github.com/authgear/authgear-server/pkg/util/httproute"
	"github.com/authgear/authgear-server/pkg/util/phone"
	"github.com/authgear/authgear-server/pkg/util/template"
	"github.com/authgear/authgear-server/pkg/util/validation"
)

var TemplateWebEnterOOBOTPHTML = template.RegisterHTML(
	"web/enter_oob_otp.html",
	Components...,
)

var EnterOOBOTPSchema = validation.NewSimpleSchema(`
	{
		"type": "object",
		"properties": {
			"x_oob_otp_code": {
				"type": "string",
				"format": "x_oob_otp_code"
			}
		},
		"required": ["x_oob_otp_code"]
	}
`)

func ConfigureEnterOOBOTPRoute(route httproute.Route) httproute.Route {
	return route.
		WithMethods("OPTIONS", "POST", "GET").
		WithPathPattern("/flows/enter_oob_otp")
}

type EnterOOBOTPViewModel struct {
	OOBOTPTarget                   string
	OOBOTPCodeSendCooldown         int
	OOBOTPCodeLength               int
	OOBOTPChannel                  string
	FailedAttemptRateLimitExceeded bool
}

type EnterOOBOTPHandler struct {
	ControllerFactory         ControllerFactory
	BaseViewModel             *viewmodels.BaseViewModeler
	AlternativeStepsViewModel *viewmodels.AlternativeStepsViewModeler
	Renderer                  Renderer
	FlashMessage              FlashMessage
	OTPCodeService            OTPCodeService
	Clock                     clock.Clock
	Config                    *config.AppConfig
}

type EnterOOBOTPNode interface {
	GetOOBOTPTarget() string
	GetOOBOTPChannel() string
	GetOOBOTPCodeLength() int
	GetOOBOTPOOBType() interaction.OOBType
}

func (h *EnterOOBOTPHandler) GetData(r *http.Request, rw http.ResponseWriter, session *webapp.Session, graph *interaction.Graph) (map[string]interface{}, error) {
	data := map[string]interface{}{}

	baseViewModel := h.BaseViewModel.ViewModel(r, rw)
	viewModel := EnterOOBOTPViewModel{}
	var n EnterOOBOTPNode
	if graph.FindLastNode(&n) {
		viewModel.OOBOTPCodeLength = n.GetOOBOTPCodeLength()
		viewModel.OOBOTPChannel = n.GetOOBOTPChannel()
		target := n.GetOOBOTPTarget()
		channel := model.AuthenticatorOOBChannel(viewModel.OOBOTPChannel)
		switch channel {
		case model.AuthenticatorOOBChannelEmail:
			viewModel.OOBOTPTarget = mail.MaskAddress(target)
		case model.AuthenticatorOOBChannelSMS:
			viewModel.OOBOTPTarget = phone.Mask(target)
		}

		state, err := h.OTPCodeService.InspectState(otp.KindOOBOTPCode(h.Config, channel), target)
		if err != nil {
			return nil, err
		}

		viewModel.FailedAttemptRateLimitExceeded = state.TooManyAttempts
		cooldown := int(state.CanResendAt.Sub(h.Clock.NowUTC()).Seconds())
		if cooldown < 0 {
			viewModel.OOBOTPCodeSendCooldown = 0
		} else {
			viewModel.OOBOTPCodeSendCooldown = cooldown
		}
	}

	currentNode := graph.CurrentNode()
	var alternatives *viewmodels.AlternativeStepsViewModel
	switch currentNode.(type) {
	case *nodes.NodeAuthenticationOOBTrigger:
		switch model.AuthenticatorOOBChannel(viewModel.OOBOTPChannel) {
		case model.AuthenticatorOOBChannelEmail:
			var err error
			alternatives, err = h.AlternativeStepsViewModel.AuthenticationAlternatives(graph, webapp.SessionStepEnterOOBOTPAuthnEmail)
			if err != nil {
				return nil, err
			}
		case model.AuthenticatorOOBChannelSMS:
			var err error
			alternatives, err = h.AlternativeStepsViewModel.AuthenticationAlternatives(graph, webapp.SessionStepEnterOOBOTPAuthnSMS)
			if err != nil {
				return nil, err
			}
		}
	case *nodes.NodeCreateAuthenticatorOOBSetup:
		switch model.AuthenticatorOOBChannel(viewModel.OOBOTPChannel) {
		case model.AuthenticatorOOBChannelEmail:
			var err error
			alternatives, err = h.AlternativeStepsViewModel.CreateAuthenticatorAlternatives(graph, webapp.SessionStepEnterOOBOTPSetupEmail)
			if err != nil {
				return nil, err
			}
		case model.AuthenticatorOOBChannelSMS:
			var err error
			alternatives, err = h.AlternativeStepsViewModel.CreateAuthenticatorAlternatives(graph, webapp.SessionStepEnterOOBOTPSetupSMS)
			if err != nil {
				return nil, err
			}
		}
	default:
		panic(fmt.Errorf("enter_oob_otp: unexpected node: %T", currentNode))
	}

	phoneOTPAlternatives := viewmodels.PhoneOTPAlternativeStepsViewModel{}
	if err := phoneOTPAlternatives.AddAlternatives(graph, session.CurrentStep().Kind); err != nil {
		return nil, err
	}

	viewmodels.Embed(data, baseViewModel)
	viewmodels.Embed(data, viewModel)
	viewmodels.Embed(data, alternatives)
	viewmodels.Embed(data, phoneOTPAlternatives)

	return data, nil
}

func (h *EnterOOBOTPHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ctrl, err := h.ControllerFactory.New(r, w)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	defer ctrl.Serve()

	ctrl.Get(func() error {
		session, err := ctrl.InteractionSession()
		if err != nil {
			return err
		}

		graph, err := ctrl.InteractionGet()
		if err != nil {
			return err
		}

		data, err := h.GetData(r, w, session, graph)
		if err != nil {
			return err
		}

		h.Renderer.RenderHTML(w, r, TemplateWebEnterOOBOTPHTML, data)
		return nil
	})

	ctrl.PostAction("resend", func() error {
		result, err := ctrl.InteractionPost(func() (input interface{}, err error) {
			input = &InputResendCode{}
			return
		})
		if err != nil {
			return err
		}

		if !result.IsInteractionErr {
			h.FlashMessage.Flash(w, string(webapp.FlashMessageTypeResendCodeSuccess))
		}
		result.WriteResponse(w, r)
		return nil
	})

	ctrl.PostAction("submit", func() error {
		result, err := ctrl.InteractionPost(func() (input interface{}, err error) {
			err = EnterOOBOTPSchema.Validator().ValidateValue(FormToJSON(r.Form))
			if err != nil {
				return
			}

			code := r.Form.Get("x_oob_otp_code")
			deviceToken := r.Form.Get("x_device_token") == "true"

			input = &InputAuthOOB{
				Code:        code,
				DeviceToken: deviceToken,
			}
			return
		})
		if err != nil {
			return err
		}

		result.WriteResponse(w, r)
		return nil
	})

	handleAlternativeSteps(ctrl)
}
