package webapp

import (
	"net/http"

	"github.com/authgear/authgear-server/pkg/api/model"
	"github.com/authgear/authgear-server/pkg/auth/handler/webapp/viewmodels"
	"github.com/authgear/authgear-server/pkg/auth/webapp"
	"github.com/authgear/authgear-server/pkg/lib/authn/otp"
	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/lib/interaction"
	"github.com/authgear/authgear-server/pkg/util/clock"
	"github.com/authgear/authgear-server/pkg/util/httproute"
	"github.com/authgear/authgear-server/pkg/util/phone"
	"github.com/authgear/authgear-server/pkg/util/template"
)

var TemplateWebWhatsappHTML = template.RegisterHTML(
	"web/whatsapp_otp.html",
	Components...,
)

func ConfigureWhatsappOTPRoute(route httproute.Route) httproute.Route {
	return route.
		WithMethods("OPTIONS", "POST", "GET").
		WithPathPattern("/flows/whatsapp_otp")
}

type WhatsappOTPNode interface {
	GetWhatsappOTPLength() int
	GetPhone() string
	GetOTPKindFactory() otp.DeprecatedKindFactory
}

type WhatsappOTPAuthnNode interface {
	GetAuthenticatorIndex() int
}

type WhatsappOTPViewModel struct {
	WhatsappOTPTarget              string
	WhatsappOTPCodeLength          int
	WhatsappOTPCodeSendCooldown    int
	FailedAttemptRateLimitExceeded bool
}

type WhatsappOTPHandler struct {
	Config                    *config.AppConfig
	Clock                     clock.Clock
	ControllerFactory         ControllerFactory
	BaseViewModel             *viewmodels.BaseViewModeler
	AlternativeStepsViewModel *viewmodels.AlternativeStepsViewModeler
	Renderer                  Renderer
	OTPCodeService            OTPCodeService
	FlashMessage              FlashMessage
}

func (h *WhatsappOTPHandler) GetData(r *http.Request, rw http.ResponseWriter, session *webapp.Session, graph *interaction.Graph) (map[string]interface{}, error) {
	data := map[string]interface{}{}
	baseViewModel := h.BaseViewModel.ViewModel(r, rw)
	whatsappViewModel := WhatsappOTPViewModel{}
	var n WhatsappOTPNode
	if graph.FindLastNode(&n) {
		target := n.GetPhone()
		channel := model.AuthenticatorOOBChannelWhatsapp
		otpkind := n.GetOTPKindFactory()(h.Config, channel)
		whatsappViewModel.WhatsappOTPTarget = phone.Mask(target)
		whatsappViewModel.WhatsappOTPCodeLength = n.GetWhatsappOTPLength()
		state, err := h.OTPCodeService.InspectState(otpkind, target)
		if err != nil {
			return nil, err
		}

		whatsappViewModel.FailedAttemptRateLimitExceeded = state.TooManyAttempts
		cooldown := int(state.CanResendAt.Sub(h.Clock.NowUTC()).Seconds())
		if cooldown < 0 {
			whatsappViewModel.WhatsappOTPCodeSendCooldown = 0
		} else {
			whatsappViewModel.WhatsappOTPCodeSendCooldown = cooldown
		}
	}

	currentStepKind := session.CurrentStep().Kind
	phoneOTPAlternatives := viewmodels.PhoneOTPAlternativeStepsViewModel{}
	if err := phoneOTPAlternatives.AddAlternatives(graph, currentStepKind); err != nil {
		return nil, err
	}
	// alternatives
	var alternatives *viewmodels.AlternativeStepsViewModel
	var node1 CreateAuthenticatorBeginNode
	var node2 AuthenticationBeginNode
	nodesInf := []interface{}{
		&node1,
		&node2,
	}
	node := graph.FindLastNodeFromList(nodesInf)
	switch node.(type) {
	case *CreateAuthenticatorBeginNode:
		var err error
		alternatives, err = h.AlternativeStepsViewModel.CreateAuthenticatorAlternatives(graph, currentStepKind)
		if err != nil {
			return nil, err
		}
	case *AuthenticationBeginNode:
		var err error
		alternatives, err = h.AlternativeStepsViewModel.AuthenticationAlternatives(graph, currentStepKind)
		if err != nil {
			return nil, err
		}
	default:
		// identity verification
		// alternatives are provided in PhoneOTPAlternativeStepsViewModel
		alternatives = &viewmodels.AlternativeStepsViewModel{}
	}

	viewmodels.Embed(data, baseViewModel)
	viewmodels.Embed(data, whatsappViewModel)
	viewmodels.Embed(data, phoneOTPAlternatives)
	viewmodels.Embed(data, *alternatives)
	return data, nil
}

func (h *WhatsappOTPHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
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

		h.Renderer.RenderHTML(w, r, TemplateWebWhatsappHTML, data)
		return nil
	})

	ctrl.PostAction("submit", func() error {
		deviceToken := r.Form.Get("x_device_token") == "true"
		otp := r.Form.Get("x_whatsapp_code")
		result, err := ctrl.InteractionPost(func() (input interface{}, err error) {
			input = &InputVerifyWhatsappOTP{
				DeviceToken: deviceToken,
				WhatsappOTP: otp,
			}
			return
		})
		if err != nil {
			return err
		}

		result.WriteResponse(w, r)
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

	handleAlternativeSteps(ctrl)
}
