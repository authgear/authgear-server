package webapp

import (
	"net/http"

	"github.com/authgear/authgear-server/pkg/auth/handler/webapp/viewmodels"
	"github.com/authgear/authgear-server/pkg/auth/webapp"
	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/lib/infra/db"
	"github.com/authgear/authgear-server/pkg/lib/session"
	"github.com/authgear/authgear-server/pkg/util/httproute"
	"github.com/authgear/authgear-server/pkg/util/template"
)

const (
	TemplateItemTypeAuthUISettingsRecoveryCodeHTML string = "auth_ui_settings_recovery_code.html"
)

var TemplateAuthUISettingsRecoveryCodeHTML = template.Register(template.T{
	Type:                    TemplateItemTypeAuthUISettingsRecoveryCodeHTML,
	IsHTML:                  true,
	TranslationTemplateType: TemplateItemTypeAuthUITranslationJSON,
	Defines:                 defines,
	ComponentTemplateTypes:  components,
})

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
	Database       *db.Handle
	BaseViewModel  *viewmodels.BaseViewModeler
	Renderer       Renderer
	WebApp         WebAppService
	Authentication *config.AuthenticationConfig
	MFA            SettingsMFAService
	CSRFCookie     webapp.CSRFCookieDef
}

func (h *SettingsRecoveryCodeHandler) GetData(r *http.Request, state *webapp.State) (map[string]interface{}, error) {
	data := map[string]interface{}{}
	var anyError interface{}
	if state != nil {
		anyError = state.Error
	}

	baseViewModel := h.BaseViewModel.ViewModel(r, anyError)
	userID := *session.GetUserID(r.Context())

	viewModel := SettingsRecoveryCodeViewModel{}
	viewModel.AllowListRecoveryCodes = h.Authentication.RecoveryCode.ListEnabled
	if viewModel.AllowListRecoveryCodes {
		codes, err := h.MFA.ListRecoveryCodes(userID)
		if err != nil {
			return nil, err
		}

		viewModel.RecoveryCodes = make([]string, len(codes))
		for i, code := range codes {
			viewModel.RecoveryCodes[i] = code.Code
		}
	}

	viewmodels.Embed(data, baseViewModel)
	viewmodels.Embed(data, viewModel)

	return data, nil
}

func (h *SettingsRecoveryCodeHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if r.Method == "GET" {
		err := h.Database.WithTx(func() error {
			state, err := h.WebApp.GetState(StateID(r))
			if err != nil {
				return err
			}

			data, err := h.GetData(r, state)
			if err != nil {
				return err
			}

			h.Renderer.RenderHTML(w, r, TemplateItemTypeAuthUISettingsRecoveryCodeHTML, data)
			return nil
		})
		if err != nil {
			panic(err)
		}
	}

	if r.Method == "POST" && r.Form.Get("x_action") == "download" {
		if !h.Authentication.RecoveryCode.ListEnabled {
			http.Error(w, "listing recovery code is disabled", http.StatusForbidden)
			return
		}

		err := h.Database.WithTx(func() error {
			state, err := h.WebApp.GetState(StateID(r))
			if err != nil {
				return err
			}

			data, err := h.GetData(r, state)
			if err != nil {
				return err
			}

			h.Renderer.Render(w, r, TemplateItemTypeAuthUIDownloadRecoveryCodeTXT, data, setRecoveryCodeAttachmentHeaders)
			return nil
		})
		if err != nil {
			panic(err)
		}
	}
}
