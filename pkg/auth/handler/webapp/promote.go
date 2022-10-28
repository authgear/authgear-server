package webapp

import (
	"net/http"

	"github.com/authgear/authgear-server/pkg/auth/handler/webapp/viewmodels"
	"github.com/authgear/authgear-server/pkg/auth/webapp"
	"github.com/authgear/authgear-server/pkg/lib/interaction"
	"github.com/authgear/authgear-server/pkg/util/httproute"
	"github.com/authgear/authgear-server/pkg/util/httputil"
	"github.com/authgear/authgear-server/pkg/util/template"
	"github.com/authgear/authgear-server/pkg/util/validation"
)

var TemplateWebPromoteHTML = template.RegisterHTML(
	"web/promote.html",
	components...,
)

var PromoteWithLoginIDSchema = validation.NewSimpleSchema(`
	{
		"type": "object",
		"properties": {
			"q_login_id_key": { "type": "string" },
			"q_login_id_type": { "type": "string" },
			"q_login_id_input_type": { "type": "string", "enum": ["email", "phone", "text"] },
			"q_login_id": { "type": "string" }
		},
		"required": ["q_login_id_key", "q_login_id_type", "q_login_id_input_type", "q_login_id"]
	}
`)

func ConfigurePromoteRoute(route httproute.Route) httproute.Route {
	return route.
		WithMethods("OPTIONS", "POST", "GET").
		WithPathPattern("/flows/promote_user")
}

type PromoteViewModel struct {
	LoginIDKey string
}

func NewPromoteViewModel(r *http.Request) PromoteViewModel {
	loginIDKey := r.Form.Get("q_login_id_key")
	return PromoteViewModel{
		LoginIDKey: loginIDKey,
	}
}

type PromoteHandler struct {
	ControllerFactory       ControllerFactory
	BaseViewModel           *viewmodels.BaseViewModeler
	AuthenticationViewModel *viewmodels.AuthenticationViewModeler
	FormPrefiller           *FormPrefiller
	Renderer                Renderer
}

func (h *PromoteHandler) GetData(r *http.Request, rw http.ResponseWriter, graph *interaction.Graph) (map[string]interface{}, error) {
	data := make(map[string]interface{})
	baseViewModel := h.BaseViewModel.ViewModel(r, rw)
	viewmodels.Embed(data, baseViewModel)
	authenticationViewModel := h.AuthenticationViewModel.NewWithGraph(graph, r.Form)
	viewmodels.Embed(data, authenticationViewModel)
	viewmodels.Embed(data, NewPromoteViewModel(r))
	return data, nil
}

func (h *PromoteHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ctrl, err := h.ControllerFactory.New(r, w)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	defer ctrl.Serve()

	h.FormPrefiller.Prefill(r.Form)

	prompt := []string{}
	if s := webapp.GetSession(r.Context()); s != nil {
		prompt = s.Prompt
	}

	ctrl.Get(func() error {
		graph, err := ctrl.InteractionGet()
		if err != nil {
			return err
		}

		data, err := h.GetData(r, w, graph)
		if err != nil {
			return err
		}

		h.Renderer.RenderHTML(w, r, TemplateWebPromoteHTML, data)
		return nil
	})

	ctrl.PostAction("oauth", func() error {
		providerAlias := r.Form.Get("x_provider_alias")
		result, err := ctrl.InteractionPost(func() (input interface{}, err error) {
			input = &InputUseOAuth{
				ProviderAlias:    providerAlias,
				ErrorRedirectURI: httputil.HostRelative(r.URL).String(),
				Prompt:           prompt,
			}
			return
		})
		if err != nil {
			return err
		}

		result.WriteResponse(w, r)
		return nil
	})

	ctrl.PostAction("login_id", func() error {
		result, err := ctrl.InteractionPost(func() (input interface{}, err error) {
			err = PromoteWithLoginIDSchema.Validator().ValidateValue(FormToJSON(r.Form))
			if err != nil {
				return
			}

			loginIDValue := r.Form.Get("q_login_id")
			loginIDKey := r.Form.Get("q_login_id_key")
			loginIDType := r.Form.Get("q_login_id_type")

			input = &InputNewLoginID{
				LoginIDType:  loginIDType,
				LoginIDKey:   loginIDKey,
				LoginIDValue: loginIDValue,
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
