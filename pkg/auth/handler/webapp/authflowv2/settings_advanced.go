package authflowv2

import (
	"net/http"

	handlerwebapp "github.com/authgear/authgear-server/pkg/auth/handler/webapp"
	"github.com/authgear/authgear-server/pkg/auth/handler/webapp/viewmodels"
	"github.com/authgear/authgear-server/pkg/util/httproute"
	"github.com/authgear/authgear-server/pkg/util/template"
)

var TemplateWebSettingsV2AdvancedSettingsHTML = template.RegisterHTML(
	"web/authflowv2/settings_advanced.html",
	handlerwebapp.SettingsComponents...,
)

type AuthflowV2SettingsAdvancedSettingsHandler struct {
	ControllerFactory handlerwebapp.ControllerFactory
	BaseViewModel     *viewmodels.BaseViewModeler
	Renderer          handlerwebapp.Renderer
}

func ConfigureSettingsV2AdvancedSettingsRoute(route httproute.Route) httproute.Route {
	return route.
		WithMethods("OPTIONS", "GET", "POST").
		WithPathPattern(SettingsV2RouteAdvancedSettings)
}

func (h *AuthflowV2SettingsAdvancedSettingsHandler) GetData(r *http.Request, w http.ResponseWriter) (map[string]interface{}, error) {
	data := map[string]interface{}{}

	// BaseViewModel
	baseViewModel := h.BaseViewModel.ViewModel(r, w)
	viewmodels.Embed(data, baseViewModel)

	return data, nil
}

func (h *AuthflowV2SettingsAdvancedSettingsHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ctrl, err := h.ControllerFactory.New(r, w)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	defer ctrl.ServeWithoutDBTx()

	ctrl.Get(func() error {
		data, err := h.GetData(r, w)
		if err != nil {
			return nil
		}
		h.Renderer.RenderHTML(w, r, TemplateWebSettingsV2AdvancedSettingsHTML, data)
		return nil
	})
}
