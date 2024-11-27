package authflowv2

import (
	"context"
	"net/http"

	handlerwebapp "github.com/authgear/authgear-server/pkg/auth/handler/webapp"
	"github.com/authgear/authgear-server/pkg/auth/handler/webapp/viewmodels"
	"github.com/authgear/authgear-server/pkg/auth/webapp"
	"github.com/authgear/authgear-server/pkg/lib/authn/mfa"
	"github.com/authgear/authgear-server/pkg/lib/infra/db/appdb"
	"github.com/authgear/authgear-server/pkg/lib/session"
	"github.com/authgear/authgear-server/pkg/util/template"
)

var TemplateWebSettingsMFAHTML = template.RegisterHTML(
	"web/authflowv2/settings_mfa.html",
	handlerwebapp.SettingsComponents...,
)

type AuthflowV2SettingsMFAHandler struct {
	Database          *appdb.Handle
	ControllerFactory handlerwebapp.ControllerFactory
	BaseViewModel     *viewmodels.BaseViewModeler
	SettingsViewModel *viewmodels.SettingsViewModeler
	Renderer          handlerwebapp.Renderer
	MFA               *mfa.Service
}

func (h *AuthflowV2SettingsMFAHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ctrl, err := h.ControllerFactory.New(r, w)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	defer ctrl.ServeWithoutDBTx(r.Context())

	ctrl.Get(func(ctx context.Context) error {
		userID := session.GetUserID(ctx)

		data := map[string]interface{}{}

		err := h.Database.WithTx(ctx, func(ctx context.Context) error {
			baseViewModel := h.BaseViewModel.ViewModel(r, w)
			viewmodels.Embed(data, baseViewModel)

			viewModelPtr, err := h.SettingsViewModel.ViewModel(ctx, *userID)
			if err != nil {
				return err
			}
			viewmodels.Embed(data, *viewModelPtr)
			return nil
		})
		if err != nil {
			return err
		}

		h.Renderer.RenderHTML(w, r, TemplateWebSettingsMFAHTML, data)

		return nil
	})

	ctrl.PostAction("revoke_device", func(ctx context.Context) error {
		userID := session.GetUserID(ctx)
		err := h.MFA.InvalidateAllDeviceTokens(ctx, *userID)
		if err != nil {
			return err
		}

		result := webapp.Result{}
		result.WriteResponse(w, r)
		return nil
	})
}
