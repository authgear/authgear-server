package webapp

import (
	"net/http"

	"github.com/authgear/authgear-server/pkg/auth/handler/webapp/viewmodels"
	"github.com/authgear/authgear-server/pkg/auth/webapp"
	"github.com/authgear/authgear-server/pkg/lib/interaction"
	"github.com/authgear/authgear-server/pkg/lib/interaction/intents"
	"github.com/authgear/authgear-server/pkg/util/httproute"
	"github.com/authgear/authgear-server/pkg/util/template"
)

var TemplateWebConfirmWeb3AccountHTML = template.RegisterHTML(
	"web/confirm_web3_account.html",
	components...,
)

func ConfigureConfirmWeb3AccountRoute(route httproute.Route) httproute.Route {
	return route.
		WithMethods("OPTIONS", "POST", "GET").
		WithPathPattern("/confirm_web3_account")
}

type ConfirmWeb3AccountViewModel struct {
	Provider string
}

type ConfirmWeb3AccountHandler struct {
	ControllerFactory         ControllerFactory
	BaseViewModel             *viewmodels.BaseViewModeler
	AlternativeStepsViewModel *viewmodels.AlternativeStepsViewModeler
	Renderer                  Renderer
}

func (h *ConfirmWeb3AccountHandler) GetData(r *http.Request, rw http.ResponseWriter, graph *interaction.Graph) (map[string]interface{}, error) {
	data := map[string]interface{}{}

	baseViewModel := h.BaseViewModel.ViewModel(r, rw)

	provider := ""
	if p := r.Form.Get("provider"); p == "" {
		provider = "metamask"
	} else {
		provider = p
	}

	confirmWeb3AccountViewModel := ConfirmWeb3AccountViewModel{
		Provider: provider,
	}

	authenticationViewModel := viewmodels.NewAuthenticationViewModelWithGraph(graph)
	viewmodels.Embed(data, authenticationViewModel)
	viewmodels.Embed(data, confirmWeb3AccountViewModel)
	viewmodels.Embed(data, baseViewModel)

	return data, nil
}

func (h *ConfirmWeb3AccountHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ctrl, err := h.ControllerFactory.New(r, w)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	defer ctrl.Serve()

	opts := webapp.SessionOptions{
		RedirectURI: ctrl.RedirectURI(),
	}

	userIDHint := ""
	webhookState := ""
	suppressIDPSessionCookie := false
	if s := webapp.GetSession(r.Context()); s != nil {
		webhookState = s.WebhookState
		userIDHint = s.UserIDHint
		suppressIDPSessionCookie = s.SuppressIDPSessionCookie
	}
	intent := &intents.IntentAuthenticate{
		Kind:                     intents.IntentAuthenticateKindLogin,
		WebhookState:             webhookState,
		UserIDHint:               userIDHint,
		SuppressIDPSessionCookie: suppressIDPSessionCookie,
	}

	ctrl.Get(func() error {
		graph, err := ctrl.EntryPointGet(opts, intent)
		if err != nil {
			return err
		}

		data, err := h.GetData(r, w, graph)
		if err != nil {
			return err
		}

		h.Renderer.RenderHTML(w, r, TemplateWebConfirmWeb3AccountHTML, data)
		return nil
	})
}
