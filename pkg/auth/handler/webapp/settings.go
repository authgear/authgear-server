package webapp

import (
	"net/http"

	"github.com/authgear/authgear-server/pkg/auth/handler/webapp/viewmodels"
	"github.com/authgear/authgear-server/pkg/auth/webapp"
	"github.com/authgear/authgear-server/pkg/lib/authn"
	"github.com/authgear/authgear-server/pkg/lib/authn/authenticator"
	"github.com/authgear/authgear-server/pkg/lib/authn/mfa"
	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/lib/infra/db"
	"github.com/authgear/authgear-server/pkg/lib/interaction"
	"github.com/authgear/authgear-server/pkg/lib/interaction/intents"
	"github.com/authgear/authgear-server/pkg/lib/session"
	"github.com/authgear/authgear-server/pkg/util/httproute"
	"github.com/authgear/authgear-server/pkg/util/httputil"
	"github.com/authgear/authgear-server/pkg/util/template"
)

var TemplateWebSettingsHTML = template.RegisterHTML(
	"web/settings.html",
	components...,
)

func ConfigureSettingsRoute(route httproute.Route) httproute.Route {
	return route.
		WithMethods("OPTIONS", "GET", "POST").
		WithPathPattern("/settings")
}

type SettingsViewModel struct {
	Authenticators           []*authenticator.Info
	SecondaryTOTPAllowed     bool
	SecondaryOOBOTPAllowed   bool
	SecondaryPasswordAllowed bool
}

type SettingsAuthenticatorService interface {
	List(userID string, filters ...authenticator.Filter) ([]*authenticator.Info, error)
}

type SettingsMFAService interface {
	ListRecoveryCodes(userID string) ([]*mfa.RecoveryCode, error)
	InvalidateAllDeviceTokens(userID string) error
}

type SettingsHandler struct {
	Database       *db.Handle
	BaseViewModel  *viewmodels.BaseViewModeler
	Renderer       Renderer
	WebApp         WebAppService
	Authentication *config.AuthenticationConfig
	Authenticators SettingsAuthenticatorService
	MFA            SettingsMFAService
}

func (h *SettingsHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	redirectURI := httputil.HostRelative(r.URL).String()
	authenticatorID := r.Form.Get("x_authenticator_id")
	userID := *session.GetUserID(r.Context())

	if r.Method == "GET" {
		data := map[string]interface{}{}

		baseViewModel := h.BaseViewModel.ViewModel(r, nil)
		viewmodels.Embed(data, baseViewModel)

		authenticators, err := h.Authenticators.List(userID)
		if err != nil {
			panic(err)
		}

		totp := false
		oobotp := false
		password := false
		for _, typ := range h.Authentication.SecondaryAuthenticators {
			switch typ {
			case authn.AuthenticatorTypePassword:
				password = true
			case authn.AuthenticatorTypeTOTP:
				totp = true
			case authn.AuthenticatorTypeOOB:
				oobotp = true
			}
		}

		viewModel := SettingsViewModel{
			Authenticators:           authenticators,
			SecondaryTOTPAllowed:     totp,
			SecondaryOOBOTPAllowed:   oobotp,
			SecondaryPasswordAllowed: password,
		}
		viewmodels.Embed(data, viewModel)

		h.Renderer.RenderHTML(w, r, TemplateWebSettingsHTML, data)
		return
	}

	if r.Method == "POST" && r.Form.Get("x_action") == "revoke_devices" {
		err := h.Database.WithTx(func() error {
			if err := h.MFA.InvalidateAllDeviceTokens(userID); err != nil {
				return err
			}
			http.Redirect(w, r, redirectURI, http.StatusFound)
			return nil
		})
		if err != nil {
			panic(err)
		}
	}

	if r.Method == "POST" && r.Form.Get("x_action") == "setup_secondary_password" {
		err := h.Database.WithTx(func() error {
			intent := &webapp.Intent{
				RedirectURI: redirectURI,
				Intent: intents.NewIntentAddAuthenticator(
					userID,
					interaction.AuthenticationStageSecondary,
					authn.AuthenticatorTypePassword,
				),
			}
			result, err := h.WebApp.PostIntent(intent, func() (input interface{}, err error) {
				return &InputCreateAuthenticator{}, nil
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

	if r.Method == "POST" && r.Form.Get("x_action") == "remove_secondary_password" {
		err := h.Database.WithTx(func() error {
			intent := &webapp.Intent{
				RedirectURI: redirectURI,
				Intent:      intents.NewIntentRemoveAuthenticator(userID),
			}
			result, err := h.WebApp.PostIntent(intent, func() (input interface{}, err error) {
				return &InputRemoveAuthenticator{
					Type: authn.AuthenticatorTypePassword,
					ID:   authenticatorID,
				}, nil
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
