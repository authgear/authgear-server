package webapp

import (
	"net/http"

	"github.com/authgear/authgear-server/pkg/auth/handler/webapp/viewmodels"
	"github.com/authgear/authgear-server/pkg/auth/webapp"
	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/lib/interaction/intents"
	"github.com/authgear/authgear-server/pkg/lib/session"
	"github.com/authgear/authgear-server/pkg/util/httproute"
	"github.com/authgear/authgear-server/pkg/util/httputil"
	"github.com/authgear/authgear-server/pkg/util/template"
)

var TemplateWebSettingsRecoveryCodeHTML = template.RegisterHTML(
	"web/settings_recovery_code.html",
	Components...,
)

func ConfigureSettingsRecoveryCodeRoute(route httproute.Route) httproute.Route {
	return route.
		WithMethods("OPTIONS", "POST", "GET").
		WithPathPattern("/settings/mfa/recovery_code")
}

type SettingsRecoveryCodeViewModel struct {
	RecoveryCodes []string
}

type SettingsRecoveryCodeHandler struct {
	ControllerFactory ControllerFactory
	BaseViewModel     *viewmodels.BaseViewModeler
	Renderer          Renderer
	Authentication    *config.AuthenticationConfig
	MFA               SettingsMFAService
}

func (h *SettingsRecoveryCodeHandler) GetData(r *http.Request, rw http.ResponseWriter) (map[string]interface{}, error) {
	data := map[string]interface{}{}
	baseViewModel := h.BaseViewModel.ViewModel(r, rw)
	userID := *session.GetUserID(r.Context())

	viewModel := SettingsRecoveryCodeViewModel{}
	if h.Authentication.RecoveryCode.ListEnabled {
		codes, err := h.MFA.ListRecoveryCodes(userID)
		if err != nil {
			return nil, err
		}

		recoveryCodes := make([]string, len(codes))
		for i, code := range codes {
			recoveryCodes[i] = code.Code
		}
		viewModel.RecoveryCodes = FormatRecoveryCodes(recoveryCodes)
	}

	viewmodels.Embed(data, baseViewModel)
	viewmodels.Embed(data, viewModel)

	return data, nil
}

func (h *SettingsRecoveryCodeHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ctrl, err := h.ControllerFactory.New(r, w)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	defer ctrl.ServeWithDBTx()

	redirectURI := httputil.HostRelative(r.URL).String()
	userID := ctrl.RequireUserID()

	listEnabled := !*h.Authentication.RecoveryCode.Disabled && h.Authentication.RecoveryCode.ListEnabled

	ctrl.Get(func() error {
		if !listEnabled {
			http.Redirect(w, r, "/settings/mfa", http.StatusFound)
			return nil
		}

		data, err := h.GetData(r, w)
		if err != nil {
			return err
		}

		h.Renderer.RenderHTML(w, r, TemplateWebSettingsRecoveryCodeHTML, data)
		return nil
	})

	ctrl.PostAction("download", func() error {
		if !h.Authentication.RecoveryCode.ListEnabled {
			http.Error(w, "listing recovery code is disabled", http.StatusForbidden)
			return nil
		}

		data, err := h.GetData(r, w)
		if err != nil {
			return err
		}

		SetRecoveryCodeAttachmentHeaders(w)
		h.Renderer.Render(w, r, TemplateWebDownloadRecoveryCodeTXT, data)
		return nil
	})

	ctrl.PostAction("regenerate", func() error {
		if !h.Authentication.RecoveryCode.ListEnabled {
			http.Error(w, "regenerate recovery code is disabled", http.StatusForbidden)
			return nil
		}

		opts := webapp.SessionOptions{
			RedirectURI: redirectURI,
		}
		intent := intents.NewIntentRegenerateRecoveryCode(userID)

		result, err := ctrl.EntryPointPost(opts, intent, func() (input interface{}, err error) {
			return nil, nil
		})
		if err != nil {
			return err
		}
		result.WriteResponse(w, r)
		return nil
	})
}
