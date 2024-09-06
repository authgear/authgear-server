package webapp

import (
	"fmt"
	"net/http"
	"net/url"

	"github.com/authgear/authgear-server/pkg/api/model"
	"github.com/authgear/authgear-server/pkg/auth/handler/webapp/viewmodels"
	"github.com/authgear/authgear-server/pkg/auth/webapp"
	"github.com/authgear/authgear-server/pkg/lib/interaction"
	"github.com/authgear/authgear-server/pkg/util/httproute"
	"github.com/authgear/authgear-server/pkg/util/template"
	"github.com/authgear/authgear-server/pkg/util/validation"
)

var TemplateWebSetupOOBOTPHTML = template.RegisterHTML(
	"web/setup_oob_otp.html",
	Components...,
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
			"x_e164": { "type": "string" }
		},
		"required": ["x_e164"]
	}
`)

func ConfigureSetupOOBOTPRoute(route httproute.Route) httproute.Route {
	return route.
		WithMethods("OPTIONS", "POST", "GET").
		WithPathPattern("/flows/setup_oob_otp_:channel")
}

type SetupOOBOTPNode interface {
	IsOOBAuthenticatorTypeAllowed(oobAuthenticatorType model.AuthenticatorType) (bool, error)
}

type SetupOOBOTPViewModel struct {
	// OOBAuthenticatorType is either AuthenticatorTypeOOBSMS or AuthenticatorTypeOOBEmail.
	OOBAuthenticatorType model.AuthenticatorType
	AlternativeSteps     []viewmodels.AlternativeStep
}

func NewSetupOOBOTPViewModel(session *webapp.Session, graph *interaction.Graph, oobAuthenticatorType model.AuthenticatorType, alternatives *viewmodels.AlternativeStepsViewModel) (*SetupOOBOTPViewModel, error) {
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

	return &SetupOOBOTPViewModel{
		OOBAuthenticatorType: oobAuthenticatorType,
		AlternativeSteps:     alternatives.AlternativeSteps,
	}, nil
}

type SetupOOBOTPHandler struct {
	ControllerFactory         ControllerFactory
	BaseViewModel             *viewmodels.BaseViewModeler
	AlternativeStepsViewModel *viewmodels.AlternativeStepsViewModeler
	Renderer                  Renderer
}

func (h *SetupOOBOTPHandler) GetData(r *http.Request, rw http.ResponseWriter, session *webapp.Session, graph *interaction.Graph, oobAuthenticatorType model.AuthenticatorType) (map[string]interface{}, error) {
	var stepKind webapp.SessionStepKind
	switch oobAuthenticatorType {
	case model.AuthenticatorTypeOOBSMS:
		stepKind = webapp.SessionStepSetupOOBOTPSMS
	case model.AuthenticatorTypeOOBEmail:
		stepKind = webapp.SessionStepSetupOOBOTPEmail
	}

	data := map[string]interface{}{}
	baseViewModel := h.BaseViewModel.ViewModel(r, rw)
	alternatives, err := h.AlternativeStepsViewModel.CreateAuthenticatorAlternatives(graph, stepKind)
	if err != nil {
		return nil, err
	}

	viewModel, err := NewSetupOOBOTPViewModel(session, graph, oobAuthenticatorType, alternatives)
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
	oobAuthenticatorType, err := model.GetOOBAuthenticatorType(model.AuthenticatorOOBChannel(oc))
	if err != nil {
		http.Error(w, "404 page not found", http.StatusNotFound)
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

		data, err := h.GetData(r, w, session, graph, oobAuthenticatorType)
		if err != nil {
			return err
		}

		h.Renderer.RenderHTML(w, r, TemplateWebSetupOOBOTPHTML, data)
		return nil
	})

	ctrl.PostAction("", func() error {
		result, err := ctrl.InteractionPost(func() (input interface{}, err error) {
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

func GetValidationSchema(oobAuthenticatorType model.AuthenticatorType) *validation.SimpleSchema {
	switch oobAuthenticatorType {
	case model.AuthenticatorTypeOOBEmail:
		return SetupOOBOTPEmailSchema
	case model.AuthenticatorTypeOOBSMS:
		return SetupOOBOTPSMSSchema
	}

	return nil
}

func FormToOOBTarget(oobAuthenticatorType model.AuthenticatorType, form url.Values) (target string, inputType string, err error) {
	if oobAuthenticatorType == model.AuthenticatorTypeOOBSMS {
		target = form.Get("x_e164")
		inputType = "phone"
		return
	}

	target = form.Get("x_email")
	inputType = "email"
	return
}
