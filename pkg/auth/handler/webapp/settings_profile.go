package webapp

import (
	"net/http"

	"github.com/authgear/authgear-server/pkg/auth/handler/webapp/viewmodels"
	"github.com/authgear/authgear-server/pkg/lib/session"
	"github.com/authgear/authgear-server/pkg/util/httproute"
	"github.com/authgear/authgear-server/pkg/util/template"
)

var TemplateWebSettingsProfileHTML = template.RegisterHTML(
	"web/settings_profile.html",
	Components...,
)

func ConfigureSettingsProfileRoute(route httproute.Route) httproute.Route {
	return route.
		WithMethods("GET").
		WithPathPattern("/settings/profile")
}

type SettingsProfileHandler struct {
	ControllerFactory        ControllerFactory
	BaseViewModel            *viewmodels.BaseViewModeler
	SettingsProfileViewModel *viewmodels.SettingsProfileViewModeler
	Renderer                 Renderer
}

func (h *SettingsProfileHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ctrl, err := h.ControllerFactory.New(r, w)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	defer ctrl.ServeWithDBTx()

	ctrl.Get(func() error {
		userID := session.GetUserID(r.Context())

		data := map[string]interface{}{}

		baseViewModel := h.BaseViewModel.ViewModel(r, w)
		viewmodels.Embed(data, baseViewModel)

		viewModelPtr, err := h.SettingsProfileViewModel.ViewModel(*userID)
		if err != nil {
			return err
		}
		viewmodels.Embed(data, *viewModelPtr)

		if viewModelPtr.IsStandardAttributesAllHidden {
			http.Redirect(w, r, "/settings", http.StatusFound)
		} else {
			h.Renderer.RenderHTML(w, r, TemplateWebSettingsProfileHTML, data)
		}

		return nil
	})
}
