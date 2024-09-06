package webapp

import (
	"net/http"
	"net/url"

	"github.com/authgear/authgear-server/pkg/auth/handler/webapp/viewmodels"
	"github.com/authgear/authgear-server/pkg/auth/webapp"
	"github.com/authgear/authgear-server/pkg/lib/interaction"
	"github.com/authgear/authgear-server/pkg/util/httproute"
	"github.com/authgear/authgear-server/pkg/util/template"
	"github.com/authgear/authgear-server/pkg/util/validation"
)

var TemplateWebSetupLoginLinkOTPHTML = template.RegisterHTML(
	"web/setup_login_link_otp.html",
	Components...,
)

var SetupLoginLinkOTPEmailSchema = validation.NewSimpleSchema(`
	{
		"type": "object",
		"properties": {
			"x_email": { "type": "string" }
		},
		"required": ["x_email"]
	}
`)

func ConfigureSetupLoginLinkOTPRoute(route httproute.Route) httproute.Route {
	return route.
		WithMethods("OPTIONS", "POST", "GET").
		WithPathPattern("/flows/setup_login_link_otp")
}

type SetupLoginLinkOTPNode interface {
}

type SetupLoginLinkOTPHandler struct {
	ControllerFactory         ControllerFactory
	BaseViewModel             *viewmodels.BaseViewModeler
	AlternativeStepsViewModel *viewmodels.AlternativeStepsViewModeler
	Renderer                  Renderer
}

func (h *SetupLoginLinkOTPHandler) GetData(r *http.Request, rw http.ResponseWriter, session *webapp.Session, graph *interaction.Graph) (map[string]interface{}, error) {
	data := map[string]interface{}{}
	baseViewModel := h.BaseViewModel.ViewModel(r, rw)
	alternatives, err := h.AlternativeStepsViewModel.CreateAuthenticatorAlternatives(graph, webapp.SessionStepSetupLoginLinkOTP)
	if err != nil {
		return nil, err
	}

	viewmodels.Embed(data, baseViewModel)
	viewmodels.Embed(data, alternatives)
	return data, nil
}

func (h *SetupLoginLinkOTPHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
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

		h.Renderer.RenderHTML(w, r, TemplateWebSetupLoginLinkOTPHTML, data)
		return nil
	})

	ctrl.PostAction("", func() error {
		result, err := ctrl.InteractionPost(func() (input interface{}, err error) {
			err = SetupLoginLinkOTPEmailSchema.Validator().ValidateValue(FormToJSON(r.Form))
			if err != nil {
				return
			}

			target, err := FormToLoginLinkTarget(r.Form)
			if err != nil {
				return
			}

			input = &InputSetupLoginLinkOTP{
				Target: target,
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

func FormToLoginLinkTarget(form url.Values) (target string, err error) {
	target = form.Get("x_email")
	return
}
