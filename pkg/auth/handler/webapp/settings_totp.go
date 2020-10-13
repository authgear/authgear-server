package webapp

import (
	"net/http"

	"github.com/authgear/authgear-server/pkg/auth/handler/webapp/viewmodels"
	"github.com/authgear/authgear-server/pkg/auth/webapp"
	"github.com/authgear/authgear-server/pkg/lib/authn"
	"github.com/authgear/authgear-server/pkg/lib/authn/authenticator"
	"github.com/authgear/authgear-server/pkg/lib/infra/db"
	"github.com/authgear/authgear-server/pkg/lib/interaction"
	"github.com/authgear/authgear-server/pkg/lib/interaction/intents"
	"github.com/authgear/authgear-server/pkg/lib/session"
	"github.com/authgear/authgear-server/pkg/util/httproute"
	"github.com/authgear/authgear-server/pkg/util/httputil"
	"github.com/authgear/authgear-server/pkg/util/template"
)

const (
	TemplateItemTypeAuthUISettingsTOTPHTML string = "auth_ui_settings_totp.html"
)

var TemplateAuthUISettingsTOTPHTML = template.Register(template.T{
	Type:                    TemplateItemTypeAuthUISettingsTOTPHTML,
	IsHTML:                  true,
	TranslationTemplateType: TemplateItemTypeAuthUITranslationJSON,
	Defines:                 defines,
	ComponentTemplateTypes:  components,
})

func ConfigureSettingsTOTPRoute(route httproute.Route) httproute.Route {
	return route.
		WithMethods("OPTIONS", "POST", "GET").
		WithPathPattern("/settings/totp")
}

type SettingsTOTPViewModel struct {
	Authenticators []*authenticator.Info
}

type SettingsTOTPHandler struct {
	Database       *db.Handle
	BaseViewModel  *viewmodels.BaseViewModeler
	Renderer       Renderer
	WebApp         WebAppService
	Authenticators SettingsAuthenticatorService
	CSRFCookie     webapp.CSRFCookieDef
}

func (h *SettingsTOTPHandler) GetData(r *http.Request, state *webapp.State) (map[string]interface{}, error) {
	data := map[string]interface{}{}
	var anyError interface{}
	if state != nil {
		anyError = state.Error
	}

	baseViewModel := h.BaseViewModel.ViewModel(r, anyError)
	userID := session.GetUserID(r.Context())
	viewModel := SettingsTOTPViewModel{}
	authenticators, err := h.Authenticators.List(*userID,
		authenticator.KeepKind(authenticator.KindSecondary),
		authenticator.KeepType(authn.AuthenticatorTypeTOTP),
	)
	if err != nil {
		return nil, err
	}
	viewModel.Authenticators = authenticators

	viewmodels.Embed(data, baseViewModel)
	viewmodels.Embed(data, viewModel)

	return data, nil
}

func (h *SettingsTOTPHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	redirectURI := httputil.HostRelative(r.URL).String()
	authenticatorID := r.Form.Get("x_authenticator_id")
	userID := session.GetUserID(r.Context())

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

			h.Renderer.RenderHTML(w, r, TemplateItemTypeAuthUISettingsTOTPHTML, data)
			return nil
		})
		if err != nil {
			panic(err)
		}
	}

	if r.Method == "POST" && r.Form.Get("x_action") == "remove" {
		err := h.Database.WithTx(func() error {
			intent := &webapp.Intent{
				RedirectURI: redirectURI,
				Intent:      intents.NewIntentRemoveAuthenticator(*userID),
			}
			result, err := h.WebApp.PostIntent(intent, func() (input interface{}, err error) {
				input = &SettingsTOTPRemove{
					AuthenticatorID: authenticatorID,
				}
				return
			})
			if err != nil {
				return err
			}
			result.WriteResponse(w, r)
			return nil
		})
		if err != nil {
			panic(err)
		}
	}

	if r.Method == "POST" && r.Form.Get("x_action") == "add" {
		err := h.Database.WithTx(func() error {
			intent := &webapp.Intent{
				RedirectURI: redirectURI,
				Intent: intents.NewIntentAddAuthenticator(
					*userID,
					interaction.AuthenticationStageSecondary,
					authn.AuthenticatorTypeTOTP,
				),
			}
			result, err := h.WebApp.PostIntent(intent, func() (input interface{}, err error) {
				input = &SettingsTOTPAdd{}
				return
			})
			if err != nil {
				return err
			}
			result.WriteResponse(w, r)
			return nil
		})
		if err != nil {
			panic(err)
		}
	}
}

type SettingsTOTPRemove struct {
	AuthenticatorID string
}

func (i *SettingsTOTPRemove) GetAuthenticatorType() authn.AuthenticatorType {
	return authn.AuthenticatorTypeTOTP
}

func (i *SettingsTOTPRemove) GetAuthenticatorID() string {
	return i.AuthenticatorID
}

type SettingsTOTPAdd struct {
}

func (i *SettingsTOTPAdd) RequestedByUser() bool {
	return true
}
