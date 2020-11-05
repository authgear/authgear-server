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
	components...,
)

func ConfigureSettingsRecoveryCodeRoute(route httproute.Route) httproute.Route {
	return route.
		WithMethods("OPTIONS", "POST", "GET").
		WithPathPattern("/settings/recovery_code")
}

type SettingsRecoveryCodeViewModel struct {
	AllowListRecoveryCodes bool
	RecoveryCodes          []string
}

type SettingsRecoveryCodeHandler struct {
	ControllerFactory ControllerFactory
	BaseViewModel     *viewmodels.BaseViewModeler
	Renderer          Renderer
	Authentication    *config.AuthenticationConfig
	MFA               SettingsMFAService
	CSRFCookie        webapp.CSRFCookieDef
}

func (h *SettingsRecoveryCodeHandler) GetData(r *http.Request, rw http.ResponseWriter) (map[string]interface{}, error) {
	data := map[string]interface{}{}
	baseViewModel := h.BaseViewModel.ViewModel(r, rw)
	userID := *session.GetUserID(r.Context())

	viewModel := SettingsRecoveryCodeViewModel{}
	viewModel.AllowListRecoveryCodes = h.Authentication.RecoveryCode.ListEnabled
	if viewModel.AllowListRecoveryCodes {
		codes, err := h.MFA.ListRecoveryCodes(userID)
		if err != nil {
			return nil, err
		}

		recoveryCodes := make([]string, len(codes))
		for i, code := range codes {
			recoveryCodes[i] = code.Code
		}
		viewModel.RecoveryCodes = formatRecoveryCodes(recoveryCodes)
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

	redirectURI := httputil.HostRelative(r.URL).String()
	userID := ctrl.RequireUserID()

	ctrl.Get(func() error {
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

		h.Renderer.Render(w, r, TemplateWebDownloadRecoveryCodeTXT, data, setRecoveryCodeAttachmentHeaders)
		return nil
	})

	ctrl.PostAction("regenerate", func() error {
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
