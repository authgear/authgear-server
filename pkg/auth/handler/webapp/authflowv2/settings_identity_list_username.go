package authflowv2

import (
	"net/http"

	handlerwebapp "github.com/authgear/authgear-server/pkg/auth/handler/webapp"
	"github.com/authgear/authgear-server/pkg/auth/handler/webapp/viewmodels"
	"github.com/authgear/authgear-server/pkg/util/httproute"
	"github.com/authgear/authgear-server/pkg/util/template"
)

var TemplateSettingsIdentityListUsernameTemplate = template.RegisterHTML(
	"web/authflowv2/settings_identity_list_username.html",
	handlerwebapp.SettingsComponents...,
)

func ConfigureAuthflowV2SettingsIdentityListUsername(route httproute.Route) httproute.Route {
	return route.
		WithMethods("OPTIONS", "POST", "GET").
		WithPathPattern(AuthflowV2RouteSettingsIdentityListUsername)
}

type AuthflowV2SettingsIdentityListUsernameHandler struct {
	ControllerFactory        handlerwebapp.ControllerFactory
	BaseViewModel            *viewmodels.BaseViewModeler
	SettingsProfileViewModel *viewmodels.SettingsProfileViewModeler
	Renderer                 handlerwebapp.Renderer
}

func (h *AuthflowV2SettingsIdentityListUsernameHandler) GetData(w http.ResponseWriter, r *http.Request) (map[string]interface{}, error) {
	data := map[string]interface{}{}
	baseViewModel := h.BaseViewModel.ViewModel(r, w)
	viewmodels.Embed(data, baseViewModel)
	return data, nil
}

func (h *AuthflowV2SettingsIdentityListUsernameHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ctrl, err := h.ControllerFactory.New(r, w)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	defer ctrl.ServeWithDBTx()

	ctrl.Get(func() error {
		data, err := h.GetData(w, r)
		if err != nil {
			return err
		}
		h.Renderer.RenderHTML(w, r, TemplateSettingsIdentityListUsernameTemplate, data)
		return nil
	})
}
