package webapp

import (
	"net/http"

	"github.com/authgear/authgear-server/pkg/auth/handler/webapp/viewmodels"
	"github.com/authgear/authgear-server/pkg/auth/webapp"
	"github.com/authgear/authgear-server/pkg/lib/authn/authenticationinfo"
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

type SettingsDeleteAccountSuccessUIInfoResolver interface {
	SetAuthenticationInfoInQuery(redirectURI string, e *authenticationinfo.Entry) string
}

type SettingsDeleteAccountSuccessAuthenticationInfoService interface {
	Get(entryID string) (entry *authenticationinfo.Entry, err error)
}

type SettingsDeleteAccountSuccessHandler struct {
	ControllerFactory         ControllerFactory
	BaseViewModel             *viewmodels.BaseViewModeler
	Renderer                  Renderer
	AccountDeletion           *config.AccountDeletionConfig
	Clock                     clock.Clock
	UIInfoResolver            SettingsDeleteAccountSuccessUIInfoResolver
	AuthenticationInfoService SettingsDeleteAccountSuccessAuthenticationInfoService
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
	defer ctrl.ServeWithDBTx()

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
		redirectURI := "/login"
		if webSession != nil && webSession.RedirectURI != "" {
			// delete account triggered by sdk via settings action
			// redirect to oauth callback
			redirectURI = webSession.RedirectURI
			if authInfoID, ok := webSession.Extra["authentication_info_id"].(string); ok {
				authInfo, err := h.AuthenticationInfoService.Get(authInfoID)
				if err != nil {
					return err
				}
				redirectURI = h.UIInfoResolver.SetAuthenticationInfoInQuery(redirectURI, authInfo)
			}
		}

		result := webapp.Result{
			RedirectURI: redirectURI,
		}
		result.WriteResponse(w, r)
		return nil
	})
}
