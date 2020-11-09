package webapp

import (
	"errors"
	"net/http"

	"github.com/authgear/authgear-server/pkg/auth/handler/webapp/viewmodels"
	"github.com/authgear/authgear-server/pkg/auth/webapp"
	"github.com/authgear/authgear-server/pkg/lib/authn"
	"github.com/authgear/authgear-server/pkg/lib/authn/identity"
	"github.com/authgear/authgear-server/pkg/lib/infra/mail"
	"github.com/authgear/authgear-server/pkg/lib/interaction"
	"github.com/authgear/authgear-server/pkg/lib/interaction/intents"
	"github.com/authgear/authgear-server/pkg/util/httproute"
	"github.com/authgear/authgear-server/pkg/util/phone"
	"github.com/authgear/authgear-server/pkg/util/template"
	"github.com/authgear/authgear-server/pkg/util/validation"
)

var TemplateWebVerifyIdentityHTML = template.RegisterHTML(
	"web/verify_identity.html",
	components...,
)

const VerifyIdentityRequestSchema = "VerifyIdentityRequestSchema"

var VerifyIdentitySchema = validation.NewMultipartSchema("").
	Add(VerifyIdentityRequestSchema, `
		{
			"type": "object",
			"properties": {
				"x_password": { "type": "string" }
			},
			"required": ["x_password"]
		}
	`).Instantiate()

func ConfigureVerifyIdentityRoute(route httproute.Route) httproute.Route {
	return route.
		WithMethods("OPTIONS", "POST", "GET").
		WithPathPattern("/verify_identity")
}

type VerifyIdentityViewModel struct {
	VerificationCode             string
	VerificationCodeSendCooldown int
	VerificationCodeLength       int
	VerificationCodeChannel      string
	IdentityDisplayID            string
}

type VerifyIdentityHandler struct {
	ControllerFactory ControllerFactory
	BaseViewModel     *viewmodels.BaseViewModeler
	Renderer          Renderer
}

type VerifyIdentityNode interface {
	GetVerificationIdentity() *identity.Info
	GetVerificationCodeChannel() string
	GetVerificationCodeSendCooldown() int
	GetVerificationCodeLength() int
}

func (h *VerifyIdentityHandler) GetData(r *http.Request, rw http.ResponseWriter, graph *interaction.Graph) (map[string]interface{}, error) {
	data := map[string]interface{}{}

	baseViewModel := h.BaseViewModel.ViewModel(r, rw)
	viewModel := VerifyIdentityViewModel{
		VerificationCode: r.Form.Get("code"),
	}
	var n VerifyIdentityNode
	if graph.FindLastNode(&n) {
		rawIdentityDisplayID := n.GetVerificationIdentity().DisplayID()
		viewModel.VerificationCodeSendCooldown = n.GetVerificationCodeSendCooldown()
		viewModel.VerificationCodeLength = n.GetVerificationCodeLength()
		viewModel.VerificationCodeChannel = n.GetVerificationCodeChannel()
		switch authn.AuthenticatorOOBChannel(viewModel.VerificationCodeChannel) {
		case authn.AuthenticatorOOBChannelSMS:
			viewModel.IdentityDisplayID = phone.Mask(rawIdentityDisplayID)
		case authn.AuthenticatorOOBChannelEmail:
			viewModel.IdentityDisplayID = mail.MaskAddress(rawIdentityDisplayID)
		default:
			viewModel.IdentityDisplayID = rawIdentityDisplayID
		}
	}

	viewmodels.Embed(data, baseViewModel)
	viewmodels.Embed(data, viewModel)

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
	defer ctrl.Serve()

	opts := webapp.SessionOptions{
		RedirectURI:     "/verify_identity/success",
		KeepAfterFinish: true,
	}
	intent := intents.NewIntentVerifyIdentityResume(r.Form.Get("id"))

	ctrl.Get(func() error {
		session, err := ctrl.InteractionSession()
		if errors.Is(err, webapp.ErrSessionNotFound) ||
			errors.Is(err, webapp.ErrInvalidSession) {
			session = nil
		}

		if session != nil {
			graph, err := ctrl.InteractionGet()
			if err != nil {
				return err
			}

			data, err := h.GetData(r, w, graph)
			if err != nil {
				return err
			}

			h.Renderer.RenderHTML(w, r, TemplateWebVerifyIdentityHTML, data)
		} else {
			graph, err := ctrl.EntryPointGet(opts, intent)
			if err != nil {
				return err
			}

			data, err := h.GetData(r, w, graph)
			if err != nil {
				return err
			}

			h.Renderer.RenderHTML(w, r, TemplateWebVerifyIdentityHTML, data)
		}
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

		result.WriteResponse(w, r)
		return nil
	})

	ctrl.PostAction("submit", func() error {
		inputFn := func() (input interface{}, err error) {
			err = VerifyIdentitySchema.PartValidator(VerifyIdentityRequestSchema).ValidateValue(FormToJSON(r.Form))
			if err != nil {
				return
			}

			code := r.Form.Get("x_password")

			input = &InputVerificationCode{
				Code: code,
			}
			return
		}

		session, err := ctrl.InteractionSession()
		if errors.Is(err, webapp.ErrSessionNotFound) {
			session = nil
		}

		var result *webapp.Result
		if session != nil {
			result, err = ctrl.InteractionPost(inputFn)
		} else {
			result, err = ctrl.EntryPointPost(opts, intent, inputFn)
		}
		if err != nil {
			return err
		}

		result.WriteResponse(w, r)
		return nil
	})
}
