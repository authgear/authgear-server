package authflowv2

import (
	"net/http"

	handlerwebapp "github.com/authgear/authgear-server/pkg/auth/handler/webapp"
	"github.com/authgear/authgear-server/pkg/auth/handler/webapp/viewmodels"
	"github.com/authgear/authgear-server/pkg/lib/session"
	"github.com/authgear/authgear-server/pkg/util/template"
)

var TemplateWebSettingsProfileHTML = template.RegisterHTML(
	"web/authflowv2/settings_profile.html",
	handlerwebapp.SettingsComponents...,
)

type AuthflowV2SettingsProfileHandler struct {
	ControllerFactory        handlerwebapp.ControllerFactory
	BaseViewModel            *viewmodels.BaseViewModeler
	SettingsProfileViewModel *viewmodels.SettingsProfileViewModeler
	Renderer                 handlerwebapp.Renderer
}

func (h *AuthflowV2SettingsProfileHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
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
			return nil
		}

		h.Renderer.RenderHTML(w, r, TemplateWebSettingsProfileHTML, data)

		return nil
	})
}
