package webapp

import (
	"net/http"

	"github.com/authgear/authgear-server/pkg/auth/handler/webapp/viewmodels"
	"github.com/authgear/authgear-server/pkg/auth/webapp"
	"github.com/authgear/authgear-server/pkg/util/httproute"
	"github.com/authgear/authgear-server/pkg/util/template"
)

var TemplateWebSelectAccountHTML = template.RegisterHTML(
	"web/select_account.html",
	components...,
)

func ConfigureSelectAccountRoute(route httproute.Route) httproute.Route {
	return route.
		WithMethods("OPTIONS", "POST", "GET").
		WithPathPattern("/select_account")
}

type SelectAccountHandler struct {
	ControllerFactory ControllerFactory
	BaseViewModel     *viewmodels.BaseViewModeler
	Renderer          Renderer
}

func (h *SelectAccountHandler) GetData(r *http.Request, rw http.ResponseWriter) (map[string]interface{}, error) {
	data := make(map[string]interface{})
	baseViewModel := h.BaseViewModel.ViewModel(r, rw)
	viewmodels.EmbedForm(data, r.Form)
	viewmodels.Embed(data, baseViewModel)
	return data, nil
}

func (h *SelectAccountHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ctrl, err := h.ControllerFactory.New(r, w)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	defer ctrl.Serve()

	redirectURI := "/settings"
	if s := webapp.GetSession(r.Context()); s != nil {
		redirectURI = s.RedirectURI
	}

	ctrl.Get(func() error {
		data, err := h.GetData(r, w)
		if err != nil {
			return err
		}
		h.Renderer.RenderHTML(w, r, TemplateWebSelectAccountHTML, data)
		return nil
	})

	ctrl.PostAction("continue", func() error {
		http.Redirect(w, r, redirectURI, http.StatusFound)
		return nil
	})

	ctrl.PostAction("login", func() error {
		http.Redirect(w, r, "/login", http.StatusFound)
		return nil
	})

}
