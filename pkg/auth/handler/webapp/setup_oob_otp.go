package webapp

import (
	"fmt"
	"net/http"
	"net/url"

	"github.com/authgear/authgear-server/pkg/auth/handler/webapp/viewmodels"
	"github.com/authgear/authgear-server/pkg/auth/webapp"
	"github.com/authgear/authgear-server/pkg/lib/authn"
	"github.com/authgear/authgear-server/pkg/lib/interaction"
	"github.com/authgear/authgear-server/pkg/util/httproute"
	"github.com/authgear/authgear-server/pkg/util/template"
	"github.com/authgear/authgear-server/pkg/util/validation"
)

var TemplateWebSetupOOBOTPHTML = template.RegisterHTML(
	"web/setup_oob_otp.html",
	components...,
)

var SetupOOBOTPEmailSchema = validation.NewSimpleSchema(`
	{
		"type": "object",
		"properties": {
			"x_email": { "type": "string" }
		},
		"required": ["x_email"]
	}
`)

var SetupOOBOTPSMSSchema = validation.NewSimpleSchema(`
	{
		"type": "object",
		"properties": {
			"x_calling_code": { "type": "string" },
			"x_national_number": { "type": "string" }
		},
		"required": ["x_calling_code", "x_national_number"]
	}
`)

func ConfigureSetupOOBOTPRoute(route httproute.Route) httproute.Route {
	return route.
		WithMethods("OPTIONS", "POST", "GET").
		WithPathPattern("/setup_oob_otp_:channel")
}

type SetupOOBOTPNode interface {
	IsOOBAuthenticatorTypeAllowed(oobAuthenticatorType authn.AuthenticatorType) (bool, error)
}

type SetupOOBOTPViewModel struct {
	// OOBAuthenticatorType is either AuthenticatorTypeOOBSMS or AuthenticatorTypeOOBEmail.
	OOBAuthenticatorType authn.AuthenticatorType
	AlternativeSteps     []viewmodels.AlternativeStep
}

func NewSetupOOBOTPViewModel(session *webapp.Session, graph *interaction.Graph, oobAuthenticatorType authn.AuthenticatorType) (*SetupOOBOTPViewModel, error) {
	var node SetupOOBOTPNode
	if !graph.FindLastNode(&node) {
		panic("setup_oob_otp: expected graph has node implementing SetupOOBOTPNode")
	}

	allowedOOBAuthenticatorType, err := node.IsOOBAuthenticatorTypeAllowed(oobAuthenticatorType)
	if err != nil {
		panic(fmt.Errorf("setup_oob_otp: unexpected error: %w", err))
	}

	if !allowedOOBAuthenticatorType {
		panic(fmt.Errorf("webapp: unexpected oob authenticator type: %s", oobAuthenticatorType))
	}

	var stepKind webapp.SessionStepKind
	switch oobAuthenticatorType {
	case authn.AuthenticatorTypeOOBSMS:
		stepKind = webapp.SessionStepSetupOOBOTPSMS
	case authn.AuthenticatorTypeOOBEmail:
		stepKind = webapp.SessionStepSetupOOBOTPEmail
	}

	alternatives := &viewmodels.AlternativeStepsViewModel{}
	err = alternatives.AddCreateAuthenticatorAlternatives(graph, stepKind)
	if err != nil {
		return nil, err
	}

	return &SetupOOBOTPViewModel{
		OOBAuthenticatorType: oobAuthenticatorType,
		AlternativeSteps:     alternatives.AlternativeSteps,
	}, nil
}

type SetupOOBOTPHandler struct {
	ControllerFactory ControllerFactory
	BaseViewModel     *viewmodels.BaseViewModeler
	Renderer          Renderer
}

func (h *SetupOOBOTPHandler) GetData(r *http.Request, rw http.ResponseWriter, session *webapp.Session, graph *interaction.Graph, oobAuthenticatorType authn.AuthenticatorType) (map[string]interface{}, error) {
	data := map[string]interface{}{}
	baseViewModel := h.BaseViewModel.ViewModel(r, rw)
	viewModel, err := NewSetupOOBOTPViewModel(session, graph, oobAuthenticatorType)
	if err != nil {
		return nil, err
	}

	viewmodels.Embed(data, baseViewModel)
	viewmodels.Embed(data, *viewModel)
	return data, nil
}

func (h *SetupOOBOTPHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ctrl, err := h.ControllerFactory.New(r, w)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	oc := httproute.GetParam(r, "channel")
	oobAuthenticatorType, err := authn.GetOOBAuthenticatorType(authn.AuthenticatorOOBChannel(oc))
	if err != nil {
		http.Error(w, "404 page not found", http.StatusNotFound)
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

		data, err := h.GetData(r, w, session, graph, oobAuthenticatorType)
		if err != nil {
			return err
		}

		h.Renderer.RenderHTML(w, r, TemplateWebSetupOOBOTPHTML, data)
		return nil
	})

	ctrl.PostAction("", func() error {
		result, err := ctrl.InteractionPost(func() (input interaction.Input, err error) {
			err = GetValidationSchema(oobAuthenticatorType).Validator().ValidateValue(FormToJSON(r.Form))
			if err != nil {
				return
			}

			target, inputType, err := FormToOOBTarget(oobAuthenticatorType, r.Form)

			input = &InputSetupOOB{
				InputType: inputType,
				Target:    target,
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

func GetValidationSchema(oobAuthenticatorType authn.AuthenticatorType) *validation.SimpleSchema {
	switch oobAuthenticatorType {
	case authn.AuthenticatorTypeOOBEmail:
		return SetupOOBOTPEmailSchema
	case authn.AuthenticatorTypeOOBSMS:
		return SetupOOBOTPSMSSchema
	}

	return nil
}

func FormToOOBTarget(oobAuthenticatorType authn.AuthenticatorType, form url.Values) (target string, inputType string, err error) {
	if oobAuthenticatorType == authn.AuthenticatorTypeOOBSMS {
		nationalNumber := form.Get("x_national_number")
		countryCallingCode := form.Get("x_calling_code")

		inputType = "phone"
		target = fmt.Sprintf("+%s%s", countryCallingCode, nationalNumber)
		return
	}

	target = form.Get("x_email")
	inputType = "email"
	return
}
