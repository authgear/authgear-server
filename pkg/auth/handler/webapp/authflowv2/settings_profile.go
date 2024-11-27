package authflowv2

import (
	"context"
	"net/http"

	handlerwebapp "github.com/authgear/authgear-server/pkg/auth/handler/webapp"
	"github.com/authgear/authgear-server/pkg/auth/handler/webapp/viewmodels"
	"github.com/authgear/authgear-server/pkg/lib/infra/db/appdb"
	"github.com/authgear/authgear-server/pkg/lib/session"
	"github.com/authgear/authgear-server/pkg/util/template"
)

var TemplateWebSettingsProfileHTML = template.RegisterHTML(
	"web/authflowv2/settings_profile.html",
	handlerwebapp.SettingsComponents...,
)

type AuthflowV2SettingsProfileHandler struct {
	Database                 *appdb.Handle
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
	defer ctrl.ServeWithoutDBTx(r.Context())

	ctrl.Get(func(ctx context.Context) error {
		data := map[string]interface{}{}
		var viewModelPtr *viewmodels.SettingsProfileViewModel

		err := h.Database.WithTx(ctx, func(ctx context.Context) error {
			userID := session.GetUserID(ctx)

			baseViewModel := h.BaseViewModel.ViewModel(r, w)
			viewmodels.Embed(data, baseViewModel)

			viewModelPtr, err = h.SettingsProfileViewModel.ViewModel(ctx, *userID)
			if err != nil {
				return err
			}
			viewmodels.Embed(data, *viewModelPtr)

			return nil
		})
		if err != nil {
			return err
		}

		if viewModelPtr.IsStandardAttributesAllHidden {
			http.Redirect(w, r, "/settings", http.StatusFound)
			return nil
		}

		h.Renderer.RenderHTML(w, r, TemplateWebSettingsProfileHTML, data)
		return nil
	})
}
