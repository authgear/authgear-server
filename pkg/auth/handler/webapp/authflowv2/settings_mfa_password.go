package authflowv2

import (
	"net/http"

	handlerwebapp "github.com/authgear/authgear-server/pkg/auth/handler/webapp"
	"github.com/authgear/authgear-server/pkg/auth/handler/webapp/viewmodels"
	"github.com/authgear/authgear-server/pkg/auth/webapp"
	"github.com/authgear/authgear-server/pkg/lib/accountmanagement"
	"github.com/authgear/authgear-server/pkg/lib/infra/db/appdb"
	"github.com/authgear/authgear-server/pkg/lib/session"
	"github.com/authgear/authgear-server/pkg/util/httproute"
	"github.com/authgear/authgear-server/pkg/util/template"
)

var TemplateWebSettingsMFAPasswordHTML = template.RegisterHTML(
	"web/authflowv2/settings_mfa_password.html",
	handlerwebapp.SettingsComponents...,
)

type AuthflowV2SettingsMFAPasswordHandler struct {
	Database          *appdb.Handle
	ControllerFactory handlerwebapp.ControllerFactory
	BaseViewModel     *viewmodels.BaseViewModeler
	SettingsViewModel *viewmodels.SettingsViewModeler
	Renderer          handlerwebapp.Renderer
	AccountManagement *accountmanagement.Service
}

func ConfigureAuthflowV2SettingsMFAPassword(route httproute.Route) httproute.Route {
	return route.
		WithMethods("OPTIONS", "POST", "GET").
		WithPathPattern(AuthflowV2RouteSettingsMFAPassword)
}

func (h *AuthflowV2SettingsMFAPasswordHandler) GetData(r *http.Request, w http.ResponseWriter) (map[string]interface{}, error) {
	userID := session.GetUserID(r.Context())
	data := map[string]interface{}{}

	// BaseViewModel
	baseViewModel := h.BaseViewModel.ViewModel(r, w)
	viewmodels.Embed(data, baseViewModel)

	// SettingsViewModel
	settingsViewModel, err := h.SettingsViewModel.ViewModel(*userID)
	if err != nil {
		return nil, err
	}
	viewmodels.Embed(data, *settingsViewModel)

	return data, nil

}

func (h *AuthflowV2SettingsMFAPasswordHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ctrl, err := h.ControllerFactory.New(r, w)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	defer ctrl.ServeWithoutDBTx()

	ctrl.Get(func() error {
		var data map[string]interface{}
		err := h.Database.WithTx(func() error {
			data, err = h.GetData(r, w)
			if err != nil {
				return err
			}
			return nil
		})
		if err != nil {
			return err
		}

		h.Renderer.RenderHTML(w, r, TemplateWebSettingsMFAPasswordHTML, data)
		return nil
	})

	ctrl.PostAction("remove", func() error {
		s := session.GetSession(r.Context())

		input := &accountmanagement.DeleteSecondaryPasswordInput{}
		_, err = h.AccountManagement.DeleteSecondaryPassword(s, input)
		if err != nil {
			return err
		}

		result := webapp.Result{RedirectURI: AuthflowV2RouteSettingsMFA}
		result.WriteResponse(w, r)

		return nil
	})
}
