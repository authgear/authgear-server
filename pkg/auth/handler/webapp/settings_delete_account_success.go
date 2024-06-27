package webapp

import (
	"net/http"

	"github.com/authgear/authgear-server/pkg/auth/handler/webapp/viewmodels"
	"github.com/authgear/authgear-server/pkg/auth/webapp"
	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/util/clock"
	"github.com/authgear/authgear-server/pkg/util/httproute"
	"github.com/authgear/authgear-server/pkg/util/template"
)

var TemplateWebSettingsDeleteAccountSuccessHTML = template.RegisterHTML(
	"web/settings_delete_account_success.html",
	Components...,
)

func ConfigureSettingsDeleteAccountSuccessRoute(route httproute.Route) httproute.Route {
	return route.
		WithMethods("OPTIONS", "POST", "GET").
		WithPathPattern("/settings/delete_account/success")
}

type SettingsDeleteAccountSuccessHandler struct {
	ControllerFactory ControllerFactory
	BaseViewModel     *viewmodels.BaseViewModeler
	Renderer          Renderer
	AccountDeletion   *config.AccountDeletionConfig
	Clock             clock.Clock
}

func (h *SettingsDeleteAccountSuccessHandler) GetData(r *http.Request, rw http.ResponseWriter) (map[string]interface{}, error) {
	data := map[string]interface{}{}
	baseViewModel := h.BaseViewModel.ViewModel(r, rw)

	now := h.Clock.NowUTC()
	deletionTime := now.Add(h.AccountDeletion.GracePeriod.Duration())
	viewModel := SettingsDeleteAccountViewModel{
		ExpectedAccountDeletionTime: deletionTime,
	}

	viewmodels.Embed(data, baseViewModel)
	viewmodels.Embed(data, viewModel)

	return data, nil
}

func (h *SettingsDeleteAccountSuccessHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ctrl, err := h.ControllerFactory.New(r, w)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	defer ctrl.Serve()

	webSession := webapp.GetSession(r.Context())

	ctrl.Get(func() error {
		data, err := h.GetData(r, w)
		if err != nil {
			return err
		}

		h.Renderer.RenderHTML(w, r, TemplateWebSettingsDeleteAccountSuccessHTML, data)
		return nil
	})

	ctrl.PostAction("", func() error {
		if webSession != nil && webSession.OAuthSessionID != "" {
			// delete account triggered by sdk via settings action
			// redirect to oauth callback
			http.Redirect(w, r, ctrl.RedirectURI(), http.StatusFound)
			return nil
		}
		http.Redirect(w, r, "/", http.StatusFound)
		return nil
	})
}
