package webapp

import (
	"errors"
	"net/http"

	"github.com/authgear/authgear-server/pkg/auth/handler/webapp/viewmodels"
	"github.com/authgear/authgear-server/pkg/auth/webapp"
	"github.com/authgear/authgear-server/pkg/lib/authn"
	"github.com/authgear/authgear-server/pkg/lib/authn/identity"
	"github.com/authgear/authgear-server/pkg/lib/feature/verification"
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
				"x_code": { "type": "string" }
			},
			"required": ["x_code"]
		}
	`).Instantiate()

func ConfigureVerifyIdentityRoute(route httproute.Route) httproute.Route {
	return route.
		WithMethods("OPTIONS", "POST", "GET").
		WithPathPattern("/verify_identity")
}

const (
	VerifyIdentityActionSubmit            = "submit"
	VerifyIdentityActionResume            = "resume"
	VerifyIdentityActionUpdateSessionStep = "update_session_step"
)

type VerifyIdentityViewModel struct {
	VerificationCode             string
	VerificationCodeSendCooldown int
	VerificationCodeLength       int
	VerificationCodeChannel      string
	IdentityDisplayID            string
	Action                       string
}

type VerifyIdentityVerificationService interface {
	GetCode(id string) (*verification.Code, error)
}

type VerifyIdentityHandler struct {
	ControllerFactory ControllerFactory
	BaseViewModel     *viewmodels.BaseViewModeler
	Renderer          Renderer
	Verifications     VerifyIdentityVerificationService
}

type VerifyIdentityNode interface {
	GetVerificationIdentity() *identity.Info
	GetVerificationCodeChannel() string
	GetVerificationCodeSendCooldown() int
	GetVerificationCodeLength() int
	GetRequestedByUser() bool
}

func (h *VerifyIdentityHandler) GetData(r *http.Request, rw http.ResponseWriter, action string, maybeSession *webapp.Session, graph *interaction.Graph) (map[string]interface{}, error) {
	data := map[string]interface{}{}

	baseViewModel := h.BaseViewModel.ViewModel(r, rw)

	code := r.Form.Get("code")

	if code == "" && maybeSession != nil {
		step := maybeSession.CurrentStep()
		if c, ok := step.FormData["x_code"].(string); ok {
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

	verificationCodeID := r.Form.Get("id")
	intent := intents.NewIntentVerifyIdentityResume(verificationCodeID)

	inputFn := func() (input interface{}, err error) {
		err = VerifyIdentitySchema.PartValidator(VerifyIdentityRequestSchema).ValidateValue(FormToJSON(r.Form))
		if err != nil {
			return
		}

		code := r.Form.Get("x_code")

		input = &InputVerificationCode{
			Code: code,
		}
		return
	}

	ctrl.Get(func() error {
		if verificationCodeID != "" {
			// The verification code ID is non-empty.
			// So this page was opened by verification link.
			// We have two outcomes.
			// If the verification is requested by the user, we want to start IntentVerifyIdentityResume.
			// Otherwise, we want to update the current SessionStep FormData.

			code, err := h.Verifications.GetCode(verificationCodeID)
			if err != nil {
				return err
			}

			var graph *interaction.Graph
			session, err := ctrl.GetSession(code.WebSessionID)
			if errors.Is(err, webapp.ErrSessionNotFound) || errors.Is(err, webapp.ErrInvalidSession) {
				session = nil
			} else if err != nil {
				return err
			}

			var action string
			if session == nil {
				// The web session is invalid.
				// Here we assume the verification is requested by the user.
				// This is because if verification is NOT requested by the user, the interaction can never continue.
				action = VerifyIdentityActionResume
			} else {
				graph, err = ctrl.InteractionGetWithSession(session)
				if err != nil {
					return err
				}
				var n VerifyIdentityNode
				if graph.FindLastNode(&n) {
					requestedByUser := n.GetRequestedByUser()
					if requestedByUser {
						action = VerifyIdentityActionResume
					} else {
						action = VerifyIdentityActionUpdateSessionStep
					}
				}
			}

			switch action {
			case VerifyIdentityActionResume:
				graph, err := ctrl.EntryPointGet(opts, intent)
				if err != nil {
					return err
				}

				data, err := h.GetData(r, w, action, session, graph)
				if err != nil {
					return err
				}

				h.Renderer.RenderHTML(w, r, TemplateWebVerifyIdentityHTML, data)
			case VerifyIdentityActionUpdateSessionStep:
				data, err := h.GetData(r, w, action, session, graph)
				if err != nil {
					return err
				}

				h.Renderer.RenderHTML(w, r, TemplateWebVerifyIdentityHTML, data)
			}
		} else {
			// The verification code ID is empty.
			// So this page should be opened by the original user agent.
			// We assume this user agent has web session cookie.
			session, err := ctrl.InteractionSession()
			if errors.Is(err, webapp.ErrSessionNotFound) || errors.Is(err, webapp.ErrInvalidSession) {
				session = nil
			} else if err != nil {
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

	ctrl.PostAction(VerifyIdentityActionSubmit, func() error {
		result, err := ctrl.InteractionPost(inputFn)
		if err != nil {
			return err
		}
		result.WriteResponse(w, r)
		return nil
	})

	ctrl.PostAction(VerifyIdentityActionResume, func() error {
		result, err := ctrl.EntryPointPost(opts, intent, inputFn)
		if err != nil {
			return err
		}
		result.WriteResponse(w, r)
		return nil
	})

	ctrl.PostAction(VerifyIdentityActionUpdateSessionStep, func() error {
		code, err := h.Verifications.GetCode(verificationCodeID)
		if err != nil {
			return err
		}

		session, err := ctrl.GetSession(code.WebSessionID)
		if err != nil {
			return err
		}

		step := session.CurrentStep()
		step.FormData["x_code"] = r.Form.Get("x_code")
		session.Steps[len(session.Steps)-1] = step

		err = ctrl.UpdateSession(session)
		if err != nil {
			return err
		}

		result := &webapp.Result{
			RedirectURI: "/return",
		}
		result.WriteResponse(w, r)
		return nil
	})
}
