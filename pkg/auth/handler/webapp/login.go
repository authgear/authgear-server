package webapp

import (
	"net/http"

	"github.com/authgear/authgear-server/pkg/auth/handler/webapp/viewmodels"
	"github.com/authgear/authgear-server/pkg/auth/webapp"
	"github.com/authgear/authgear-server/pkg/lib/interaction"
	"github.com/authgear/authgear-server/pkg/lib/interaction/intents"
	"github.com/authgear/authgear-server/pkg/util/httproute"
	"github.com/authgear/authgear-server/pkg/util/httputil"
	"github.com/authgear/authgear-server/pkg/util/template"
	"github.com/authgear/authgear-server/pkg/util/validation"
)

var TemplateWebLoginHTML = template.RegisterHTML(
	"web/login.html",
	components...,
)

var LoginWithLoginIDSchema = validation.NewSimpleSchema(`
	{
		"type": "object",
		"properties": {
			"x_login_id_input_type": { "type": "string", "enum": ["email", "phone", "text"] },
			"x_login_id": { "type": "string" }
		},
		"required": ["x_login_id_input_type", "x_login_id"]
	}
`)

func ConfigureLoginRoute(route httproute.Route) httproute.Route {
	return route.
		WithMethods("OPTIONS", "POST", "GET").
		WithPathPattern("/login")
}

type LoginViewModel struct {
	AllowLoginOnly bool
}

type LoginHandler struct {
	ControllerFactory ControllerFactory
	BaseViewModel     *viewmodels.BaseViewModeler
	FormPrefiller     *FormPrefiller
	Renderer          Renderer
}

func (h *LoginHandler) GetData(r *http.Request, rw http.ResponseWriter, graph *interaction.Graph, allowLoginOnly bool) (map[string]interface{}, error) {
	data := make(map[string]interface{})
	baseViewModel := h.BaseViewModel.ViewModel(r, rw)
	viewModel := LoginViewModel{
		AllowLoginOnly: allowLoginOnly,
	}
	viewmodels.EmbedForm(data, r.Form)
	viewmodels.Embed(data, baseViewModel)
	authenticationViewModel := viewmodels.NewAuthenticationViewModelWithGraph(graph)
	viewmodels.Embed(data, authenticationViewModel)
	viewmodels.Embed(data, viewModel)
	return data, nil
}

func (h *LoginHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ctrl, err := h.ControllerFactory.New(r, w)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	defer ctrl.Serve()

	h.FormPrefiller.Prefill(r.Form)

	opts := webapp.SessionOptions{
		RedirectURI: ctrl.RedirectURI(),
	}

	userIDHint := ""
	webhookState := ""
	suppressIDPSessionCookie := false
	prompt := []string{}
	if s := webapp.GetSession(r.Context()); s != nil {
		webhookState = s.WebhookState
		prompt = s.Prompt
		userIDHint = s.UserIDHint
		suppressIDPSessionCookie = s.SuppressIDPSessionCookie
	}
	intent := &intents.IntentAuthenticate{
		Kind:                     intents.IntentAuthenticateKindLogin,
		WebhookState:             webhookState,
		UserIDHint:               userIDHint,
		SuppressIDPSessionCookie: suppressIDPSessionCookie,
	}

	allowLoginOnly := intent.UserIDHint != ""

	ctrl.Get(func() error {
		graph, err := ctrl.EntryPointGet(opts, intent)
		if err != nil {
			return err
		}

		data, err := h.GetData(r, w, graph, allowLoginOnly)
		if err != nil {
			return err
		}

		h.Renderer.RenderHTML(w, r, TemplateWebLoginHTML, data)
		return nil
	})

	ctrl.PostAction("oauth", func() error {
		providerAlias := r.Form.Get("x_provider_alias")
		result, err := ctrl.EntryPointPost(opts, intent, func() (input interface{}, err error) {
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
		result, err := ctrl.EntryPointPost(opts, intent, func() (input interface{}, err error) {
			err = LoginWithLoginIDSchema.Validator().ValidateValue(FormToJSON(r.Form))
			if err != nil {
				return
			}

			loginID := r.Form.Get("x_login_id")

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
