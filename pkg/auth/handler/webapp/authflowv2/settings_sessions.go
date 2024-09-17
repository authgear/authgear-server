package authflowv2

import (
	"net/http"

	handlerwebapp "github.com/authgear/authgear-server/pkg/auth/handler/webapp"
	"github.com/authgear/authgear-server/pkg/auth/handler/webapp/viewmodels"
	"github.com/authgear/authgear-server/pkg/lib/session"
	"github.com/authgear/authgear-server/pkg/util/template"
)

var TemplateWebSettingsV2SessionsHTML = template.RegisterHTML(
	"web/authflowv2/settings_sessions.html",
	handlerwebapp.SettingsComponents...,
)

type AuthflowV2SettingsSessionsHandler struct {
	ControllerFactory handlerwebapp.ControllerFactory
	BaseViewModel     *viewmodels.BaseViewModeler
	Renderer          handlerwebapp.Renderer
}

func (h *AuthflowV2SettingsSessionsHandler) GetData(r *http.Request, rw http.ResponseWriter, s session.ResolvedSession) (map[string]interface{}, error) {
	data := map[string]interface{}{}

	// BaseViewModel
	baseViewModel := h.BaseViewModel.ViewModel(r, rw)
	viewmodels.Embed(data, baseViewModel)

	return data, nil
}

func (h *AuthflowV2SettingsSessionsHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ctrl, err := h.ControllerFactory.New(r, w)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	defer ctrl.ServeWithDBTx()

	currentSession := session.GetSession(r.Context())

	ctrl.Get(func() error {
		data, err := h.GetData(r, w, currentSession)
		if err != nil {
			return err
		}

		h.Renderer.RenderHTML(w, r, TemplateWebSettingsV2SessionsHTML, data)

		return nil
	})
}
