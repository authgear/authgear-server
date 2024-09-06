package webapp

import (
	"net/http"

	"github.com/authgear/authgear-server/pkg/auth/handler/webapp/viewmodels"
	"github.com/authgear/authgear-server/pkg/auth/webapp"
	"github.com/authgear/authgear-server/pkg/lib/interaction"
	"github.com/authgear/authgear-server/pkg/lib/interaction/intents"
	"github.com/authgear/authgear-server/pkg/util/httproute"
	"github.com/authgear/authgear-server/pkg/util/template"
	"github.com/authgear/authgear-server/pkg/util/validation"
)

var TemplateWebEnterPasswordHTML = template.RegisterHTML(
	"web/enter_password.html",
	Components...,
)

var EnterPasswordSchema = validation.NewSimpleSchema(`
	{
		"type": "object",
		"properties": {
			"x_password": { "type": "string" },
			"x_stage": { "type": "string" }
		},
		"required": ["x_password", "x_stage"]
	}
`)

func ConfigureEnterPasswordRoute(route httproute.Route) httproute.Route {
	return route.
		WithMethods("OPTIONS", "POST", "GET").
		WithPathPattern("/flows/enter_password")
}

type EnterPasswordViewVariant string

const (
	EnterPasswordViewVariantDefault EnterPasswordViewVariant = "default"
	EnterPasswordViewVariantReAuth  EnterPasswordViewVariant = "reauth"
)

type EnterPasswordViewModel struct {
	IdentityDisplayID string
	// ForgotPasswordInputType is either phone or email
	ForgotPasswordInputType string
	ForgotPasswordLoginID   string
	Variant                 EnterPasswordViewVariant
}

type EnterPasswordHandler struct {
	ControllerFactory         ControllerFactory
	BaseViewModel             *viewmodels.BaseViewModeler
	AlternativeStepsViewModel *viewmodels.AlternativeStepsViewModeler
	Renderer                  Renderer
}

func (h *EnterPasswordHandler) GetData(r *http.Request, rw http.ResponseWriter, session *webapp.Session, graph *interaction.Graph) (map[string]interface{}, error) {
	data := map[string]interface{}{}

	baseViewModel := h.BaseViewModel.ViewModel(r, rw)

	alternatives, err := h.AlternativeStepsViewModel.AuthenticationAlternatives(graph, webapp.SessionStepEnterPassword)
	if err != nil {
		return nil, err
	}

	identityDisplayID := ""
	forgotPasswordInputType := ""
	forgotPasswordLoginID := ""

	if identityInfo, ok := graph.GetUserLastIdentity(); ok {
		identityDisplayID = identityInfo.DisplayID()
		phoneFormat := validation.FormatPhone{}
		emailFormat := validation.FormatEmail{AllowName: false}

		// Instead of using the login id type, we parse the login id value for the type
		// So user cannot use this flow to check the identity type
		if err := phoneFormat.CheckFormat(identityDisplayID); err == nil {
			forgotPasswordInputType = "phone"
			forgotPasswordLoginID = identityDisplayID
		} else if err := emailFormat.CheckFormat(identityDisplayID); err == nil {
			forgotPasswordInputType = "email"
			forgotPasswordLoginID = identityDisplayID
		}
	}

	var variant EnterPasswordViewVariant
	switch graph.Intent.(type) {
	case *intents.IntentReauthenticate:
		variant = EnterPasswordViewVariantReAuth
	default:
		variant = EnterPasswordViewVariantDefault
	}

	enterPasswordViewModel := EnterPasswordViewModel{
		IdentityDisplayID:       identityDisplayID,
		ForgotPasswordInputType: forgotPasswordInputType,
		ForgotPasswordLoginID:   forgotPasswordLoginID,
		Variant:                 variant,
	}

	viewmodels.Embed(data, baseViewModel)
	viewmodels.Embed(data, enterPasswordViewModel)
	viewmodels.Embed(data, *alternatives)

	return data, nil
}

func (h *EnterPasswordHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ctrl, err := h.ControllerFactory.New(r, w)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	defer ctrl.ServeWithDBTx()

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

		h.Renderer.RenderHTML(w, r, TemplateWebEnterPasswordHTML, data)
		return nil
	})

	ctrl.PostAction("", func() error {
		result, err := ctrl.InteractionPost(func() (input interface{}, err error) {
			err = EnterPasswordSchema.Validator().ValidateValue(FormToJSON(r.Form))
			if err != nil {
				return
			}

			plainPassword := r.Form.Get("x_password")
			deviceToken := r.Form.Get("x_device_token") == "true"
			stage := r.Form.Get("x_stage")

			input = &InputAuthPassword{
				Stage:       stage,
				Password:    plainPassword,
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
