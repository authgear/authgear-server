package authflowv2

import (
	"net/http"

	handlerwebapp "github.com/authgear/authgear-server/pkg/auth/handler/webapp"
	"github.com/authgear/authgear-server/pkg/auth/handler/webapp/viewmodels"
	"github.com/authgear/authgear-server/pkg/auth/webapp"
	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/util/clock"
	"github.com/authgear/authgear-server/pkg/util/template"
)

var TemplateWebSettingsV2DeleteAccountSuccessHTML = template.RegisterHTML(
	"web/authflowv2/settings_delete_account_success.html",
	handlerwebapp.SettingsComponents...,
)

type AuthflowV2SettingsDeleteAccountSuccessHandler struct {
	ControllerFactory         handlerwebapp.ControllerFactory
	BaseViewModel             *viewmodels.BaseViewModeler
	Renderer                  handlerwebapp.Renderer
	AccountDeletion           *config.AccountDeletionConfig
	Clock                     clock.Clock
	UIInfoResolver            handlerwebapp.SettingsDeleteAccountSuccessUIInfoResolver
	AuthenticationInfoService handlerwebapp.SettingsDeleteAccountSuccessAuthenticationInfoService
}

func (h *AuthflowV2SettingsDeleteAccountSuccessHandler) GetData(r *http.Request, rw http.ResponseWriter) (map[string]interface{}, error) {
	data := map[string]interface{}{}

	// BaseViewModel
	baseViewModel := h.BaseViewModel.ViewModel(r, rw)
	viewmodels.Embed(data, baseViewModel)

	// DeleteAccountViewModel
	now := h.Clock.NowUTC()
	deletionTime := now.Add(h.AccountDeletion.GracePeriod.Duration())
	deleteAccountViewModel := AuthflowV2SettingsDeleteAccountViewModel{
		ExpectedAccountDeletionTime: deletionTime,
	}
	viewmodels.Embed(data, deleteAccountViewModel)

	return data, nil
}

func (h *AuthflowV2SettingsDeleteAccountSuccessHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ctrl, err := h.ControllerFactory.New(r, w)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	defer ctrl.ServeWithoutDBTx()

	webSession := webapp.GetSession(r.Context())

	ctrl.Get(func() error {
		data, err := h.GetData(r, w)
		if err != nil {
			return nil
		}
		h.Renderer.RenderHTML(w, r, TemplateWebSettingsV2DeleteAccountSuccessHTML, data)
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
