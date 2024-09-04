package authflowv2

import (
	"net/http"

	handlerwebapp "github.com/authgear/authgear-server/pkg/auth/handler/webapp"
	"github.com/authgear/authgear-server/pkg/auth/handler/webapp/viewmodels"
	"github.com/authgear/authgear-server/pkg/lib/session"
	"github.com/authgear/authgear-server/pkg/util/httproute"
	"github.com/authgear/authgear-server/pkg/util/template"
)

var TemplateWebSettingsMFACreatePasswordHTML = template.RegisterHTML(
	"web/authflowv2/settings_mfa_create_password.html",
	handlerwebapp.SettingsComponents...,
)

func ConfigureAuthflowV2SettingsMFACreatePassword(route httproute.Route) httproute.Route {
	return route.
		WithMethods("OPTIONS", "POST", "GET").
		WithPathPattern(AuthflowV2RouteSettingsMFACreatePassword)
}

type AuthflowV2SettingsMFACreatePasswordHandler struct {
	ControllerFactory handlerwebapp.ControllerFactory
	BaseViewModel     *viewmodels.BaseViewModeler
	SettingsViewModel *viewmodels.SettingsViewModeler
	Renderer          handlerwebapp.Renderer
}

func (h *AuthflowV2SettingsMFACreatePasswordHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ctrl, err := h.ControllerFactory.New(r, w)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	defer ctrl.Serve()

	ctrl.Get(func() error {
		userID := session.GetUserID(r.Context())

		data := map[string]interface{}{}

		baseViewModel := h.BaseViewModel.ViewModel(r, w)
		viewmodels.Embed(data, baseViewModel)

		viewModelPtr, err := h.SettingsViewModel.ViewModel(*userID)
		if err != nil {
			return err
		}
		viewmodels.Embed(data, *viewModelPtr)

		h.Renderer.RenderHTML(w, r, TemplateWebSettingsMFACreatePasswordHTML, data)

		return nil
	})
}
