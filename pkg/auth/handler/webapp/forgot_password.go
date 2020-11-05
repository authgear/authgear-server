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
	components...,
)

const ForgotPasswordRequestSchema = "ForgotPasswordRequestSchema"

var ForgotPasswordSchema = validation.NewMultipartSchema("").
	Add(ForgotPasswordRequestSchema, `
		{
			"type": "object",
			"properties": {
				"x_login_id_input_type": { "type": "string", "enum": ["email", "phone", "text"] },
				"x_calling_code": { "type": "string" },
				"x_national_number": { "type": "string" },
				"x_login_id": { "type": "string" }
			},
			"required": ["x_login_id_input_type"],
			"allOf": [
				{
					"if": {
						"properties": {
							"x_login_id_input_type": { "type": "string", "const": "phone" }
						}
					},
					"then": {
						"required": ["x_calling_code", "x_national_number"]
					}
				},
				{
					"if": {
						"properties": {
							"x_login_id_input_type": { "type": "string", "enum": ["text", "email"] }
						}
					},
					"then": {
						"required": ["x_login_id"]
					}
				}
			]
		}
	`).Instantiate()

func ConfigureForgotPasswordRoute(route httproute.Route) httproute.Route {
	return route.
		WithMethods("OPTIONS", "POST", "GET").
		WithPathPattern("/forgot_password")
}

type ForgotPasswordHandler struct {
	ControllerFactory ControllerFactory
	BaseViewModel     *viewmodels.BaseViewModeler
	FormPrefiller     *FormPrefiller
	Renderer          Renderer
}

func (h *ForgotPasswordHandler) GetData(r *http.Request, rw http.ResponseWriter, graph *interaction.Graph) (map[string]interface{}, error) {
	data := make(map[string]interface{})
	baseViewModel := h.BaseViewModel.ViewModel(r, rw)
	authenticationViewModel := viewmodels.NewAuthenticationViewModelWithGraph(graph)
	viewmodels.EmbedForm(data, r.Form)
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
			err = ForgotPasswordSchema.PartValidator(ForgotPasswordRequestSchema).ValidateValue(FormToJSON(r.Form))
			if err != nil {
				return
			}

			loginID, err := FormToLoginID(r.Form)
			if err != nil {
				return
			}

			input = &InputForgotPassword{
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
