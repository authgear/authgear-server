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

var TemplateWebForgotPasswordHTML = template.RegisterHTML(
	"web/forgot_password.html",
	Components...,
)

var ForgotPasswordSchema = validation.NewSimpleSchema(`
	{
		"type": "object",
		"properties": {
			"q_login_id_input_type": { "type": "string", "enum": ["email", "phone", "text"] },
			"q_login_id": { "type": "string" }
		},
		"required": ["q_login_id_input_type", "q_login_id"]
	}
`)

func ConfigureForgotPasswordRoute(route httproute.Route) httproute.Route {
	return route.
		WithMethods("OPTIONS", "POST", "GET").
		WithPathPattern("/flows/forgot_password")
}

type ForgotPasswordViewModel struct {
	LoginIDInputType string
	LoginID          string
}

func NewForgotPasswordViewModel(r *http.Request) ForgotPasswordViewModel {
	loginIDInputType := r.Form.Get("q_login_id_input_type")
	loginID := r.Form.Get("q_login_id")
	return ForgotPasswordViewModel{
		LoginIDInputType: loginIDInputType,
		LoginID:          loginID,
	}
}

type ForgotPasswordHandler struct {
	ControllerFactory       ControllerFactory
	BaseViewModel           *viewmodels.BaseViewModeler
	AuthenticationViewModel *viewmodels.AuthenticationViewModeler
	FormPrefiller           *FormPrefiller
	Renderer                Renderer
}

func (h *ForgotPasswordHandler) GetData(r *http.Request, rw http.ResponseWriter, graph *interaction.Graph) (map[string]interface{}, error) {
	data := make(map[string]interface{})
	baseViewModel := h.BaseViewModel.ViewModel(r, rw)
	authenticationViewModel := h.AuthenticationViewModel.NewWithGraph(graph, r.Form)
	viewmodels.Embed(data, NewForgotPasswordViewModel(r))
	viewmodels.Embed(data, baseViewModel)
	viewmodels.Embed(data, authenticationViewModel)
	return data, nil
}

func (h *ForgotPasswordHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ctrl, err := h.ControllerFactory.New(r, w)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	defer ctrl.ServeWithDBTx()

	opts := webapp.SessionOptions{
		KeepAfterFinish: true,
	}
	intent := intents.NewIntentForgotPassword()

	h.FormPrefiller.Prefill(r.Form)

	ctrl.Get(func() error {
		graph, err := ctrl.EntryPointGet(opts, intent)
		if err != nil {
			return err
		}

		data, err := h.GetData(r, w, graph)
		if err != nil {
			return err
		}

		h.Renderer.RenderHTML(w, r, TemplateWebForgotPasswordHTML, data)
		return nil
	})

	ctrl.PostAction("", func() error {
		result, err := ctrl.EntryPointPost(opts, intent, func() (input interface{}, err error) {
			err = ForgotPasswordSchema.Validator().ValidateValue(FormToJSON(r.Form))
			if err != nil {
				return
			}

			loginID := r.Form.Get("q_login_id")

			input = &InputUseLoginID{
				LoginID: loginID,
			}
			return
		})
		if err != nil {
			return err
		}

		result.WriteResponse(w, r)
		return nil
	})
}
