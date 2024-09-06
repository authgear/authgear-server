package webapp

import (
	"net/http"

	"github.com/authgear/authgear-server/pkg/api/model"
	"github.com/authgear/authgear-server/pkg/auth/handler/webapp/viewmodels"
	"github.com/authgear/authgear-server/pkg/auth/webapp"
	"github.com/authgear/authgear-server/pkg/lib/authn/identity"
	"github.com/authgear/authgear-server/pkg/lib/authn/otp"
	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/lib/infra/mail"
	"github.com/authgear/authgear-server/pkg/lib/interaction"
	"github.com/authgear/authgear-server/pkg/util/clock"
	"github.com/authgear/authgear-server/pkg/util/httproute"
	"github.com/authgear/authgear-server/pkg/util/phone"
	"github.com/authgear/authgear-server/pkg/util/template"
	"github.com/authgear/authgear-server/pkg/util/validation"
)

var TemplateWebVerifyIdentityHTML = template.RegisterHTML(
	"web/verify_identity.html",
	Components...,
)

var VerifyIdentitySchema = validation.NewSimpleSchema(`
	{
		"type": "object",
		"properties": {
			"x_verification_code": {
				"type": "string",
				"format": "x_verification_code"
			}
		},
		"required": ["x_verification_code"]
	}
`)

func ConfigureVerifyIdentityRoute(route httproute.Route) httproute.Route {
	return route.
		WithMethods("OPTIONS", "POST", "GET").
		WithPathPattern("/flows/verify_identity")
}

const (
	VerifyIdentityActionSubmit = "submit"
)

type VerifyIdentityViewModel struct {
	VerificationCode               string
	VerificationCodeSendCooldown   int
	VerificationCodeLength         int
	VerificationCodeChannel        string
	FailedAttemptRateLimitExceeded bool
	IdentityDisplayID              string
	Action                         string
}

type FlashMessage interface {
	Flash(rw http.ResponseWriter, messageType string)
}

type VerifyIdentityHandler struct {
	ControllerFactory ControllerFactory
	BaseViewModel     *viewmodels.BaseViewModeler
	Renderer          Renderer
	FlashMessage      FlashMessage
	OTPCodeService    OTPCodeService
	Clock             clock.Clock
	Config            *config.AppConfig
}

type VerifyIdentityNode interface {
	GetVerificationIdentity() *identity.Info
	GetVerificationCodeChannel() string
	GetVerificationCodeTarget() string
	GetVerificationCodeLength() int
	GetRequestedByUser() bool
}

func (h *VerifyIdentityHandler) GetData(r *http.Request, rw http.ResponseWriter, action string, maybeSession *webapp.Session, graph *interaction.Graph) (map[string]interface{}, error) {
	data := map[string]interface{}{}

	baseViewModel := h.BaseViewModel.ViewModel(r, rw)

	code := r.Form.Get("code")

	if code == "" && maybeSession != nil {
		step := maybeSession.CurrentStep()
		if c, ok := step.FormData["x_verification_code"].(string); ok {
			code = c
		}
	}

	viewModel := VerifyIdentityViewModel{
		VerificationCode: code,
		Action:           action,
	}

	var n VerifyIdentityNode
	if graph.FindLastNode(&n) {
		rawIdentityDisplayID := n.GetVerificationIdentity().DisplayID()
		viewModel.VerificationCodeLength = n.GetVerificationCodeLength()
		viewModel.VerificationCodeChannel = n.GetVerificationCodeChannel()
		target := n.GetVerificationCodeTarget()
		channel := model.AuthenticatorOOBChannel(viewModel.VerificationCodeChannel)
		switch channel {
		case model.AuthenticatorOOBChannelSMS:
			viewModel.IdentityDisplayID = phone.Mask(rawIdentityDisplayID)
		case model.AuthenticatorOOBChannelEmail:
			viewModel.IdentityDisplayID = mail.MaskAddress(rawIdentityDisplayID)
		default:
			panic("webapp: unknown verification channel")
		}

		state, err := h.OTPCodeService.InspectState(otp.KindVerification(h.Config, channel), target)
		if err != nil {
			return nil, err
		}
		viewModel.FailedAttemptRateLimitExceeded = state.TooManyAttempts
		cooldown := int(state.CanResendAt.Sub(h.Clock.NowUTC()).Seconds())
		if cooldown < 0 {
			viewModel.VerificationCodeSendCooldown = 0
		} else {
			viewModel.VerificationCodeSendCooldown = cooldown
		}
	}

	phoneOTPAlternatives := viewmodels.PhoneOTPAlternativeStepsViewModel{}
	if err := phoneOTPAlternatives.AddAlternatives(graph, webapp.SessionStepVerifyIdentityViaOOBOTP); err != nil {
		return nil, err
	}

	viewmodels.Embed(data, baseViewModel)
	viewmodels.Embed(data, viewModel)
	viewmodels.Embed(data, phoneOTPAlternatives)

	return data, nil
}

func (h *VerifyIdentityHandler) GetErrorData(r *http.Request, rw http.ResponseWriter, err error) (map[string]interface{}, error) {
	data := map[string]interface{}{}

	baseViewModel := h.BaseViewModel.ViewModel(r, rw)
	baseViewModel.SetError(err)
	viewModel := VerifyIdentityViewModel{}

	viewmodels.Embed(data, baseViewModel)
	viewmodels.Embed(data, viewModel)

	return data, nil
}

func (h *VerifyIdentityHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ctrl, err := h.ControllerFactory.New(r, w)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	defer ctrl.ServeWithDBTx()

	inputFn := func() (input interface{}, err error) {
		err = VerifyIdentitySchema.Validator().ValidateValue(FormToJSON(r.Form))
		if err != nil {
			return
		}

		code := r.Form.Get("x_verification_code")

		input = &InputVerificationCode{
			Code: code,
		}
		return
	}

	ctrl.Get(func() error {
		// This page should be opened by the original user agent.
		session, err := ctrl.InteractionSession()
		if err != nil {
			return err
		}

		graph, err := ctrl.InteractionGet()
		if err != nil {
			return err
		}

		data, err := h.GetData(r, w, VerifyIdentityActionSubmit, session, graph)
		if err != nil {
			return err
		}

		h.Renderer.RenderHTML(w, r, TemplateWebVerifyIdentityHTML, data)
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

	ctrl.PostAction(VerifyIdentityActionSubmit, func() error {
		result, err := ctrl.InteractionPost(inputFn)
		if err != nil {
			return err
		}
		result.WriteResponse(w, r)
		return nil
	})

	handleAlternativeSteps(ctrl)
}
