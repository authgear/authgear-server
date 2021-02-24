package webapp

import (
	"net/http"

	"github.com/authgear/authgear-server/pkg/auth/handler/webapp/viewmodels"
	"github.com/authgear/authgear-server/pkg/auth/webapp"
	"github.com/authgear/authgear-server/pkg/lib/authn"
	"github.com/authgear/authgear-server/pkg/lib/authn/authenticator"
	"github.com/authgear/authgear-server/pkg/lib/authn/mfa"
	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/lib/interaction"
	"github.com/authgear/authgear-server/pkg/lib/interaction/intents"
	"github.com/authgear/authgear-server/pkg/util/httproute"
	"github.com/authgear/authgear-server/pkg/util/httputil"
	"github.com/authgear/authgear-server/pkg/util/template"
)

var TemplateWebSettingsMFAHTML = template.RegisterHTML(
	"web/settings_mfa.html",
	components...,
)

func ConfigureSettingsMFARoute(route httproute.Route) httproute.Route {
	return route.
		WithMethods("OPTIONS", "POST", "GET").
		WithPathPattern("/settings/mfa")
}

type SettingsMFAViewModel struct {
	Authenticators              []*authenticator.Info
	SecondaryTOTPAllowed        bool
	SecondaryOOBOTPEmailAllowed bool
	SecondaryOOBOTPSMSAllowed   bool
	SecondaryPasswordAllowed    bool
	HasDeviceTokens             bool
	ListRecoveryCodesAllowed    bool
}

type SettingsMFAService interface {
	ListRecoveryCodes(userID string) ([]*mfa.RecoveryCode, error)
	InvalidateAllDeviceTokens(userID string) error
	HasDeviceTokens(userID string) (bool, error)
}

type SettingsMFAHandler struct {
	ControllerFactory ControllerFactory
	BaseViewModel     *viewmodels.BaseViewModeler
	Renderer          Renderer
	Authentication    *config.AuthenticationConfig
	Authenticators    SettingsAuthenticatorService
	MFA               SettingsMFAService
	CSRFCookie        webapp.CSRFCookieDef
}

func (h *SettingsMFAHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ctrl, err := h.ControllerFactory.New(r, w)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	defer ctrl.Serve()

	redirectURI := httputil.HostRelative(r.URL).String()
	authenticatorID := r.Form.Get("x_authenticator_id")
	userID := ctrl.RequireUserID()

	ctrl.Get(func() error {
		data := map[string]interface{}{}

		baseViewModel := h.BaseViewModel.ViewModel(r, w)
		viewmodels.Embed(data, baseViewModel)

		authenticators, err := h.Authenticators.List(userID)
		if err != nil {
			panic(err)
		}

		totp := false
		oobotpemail := false
		oobotpsms := false
		password := false
		for _, typ := range h.Authentication.SecondaryAuthenticators {
			switch typ {
			case authn.AuthenticatorTypePassword:
				password = true
			case authn.AuthenticatorTypeTOTP:
				totp = true
			case authn.AuthenticatorTypeOOBEmail:
				oobotpemail = true
			case authn.AuthenticatorTypeOOBSMS:
				oobotpsms = true
			}
		}

		hasDeviceTokens, err := h.MFA.HasDeviceTokens(userID)
		if err != nil {
			return err
		}

		viewModel := SettingsMFAViewModel{
			Authenticators:              authenticators,
			SecondaryTOTPAllowed:        totp,
			SecondaryOOBOTPEmailAllowed: oobotpemail,
			SecondaryOOBOTPSMSAllowed:   oobotpsms,
			SecondaryPasswordAllowed:    password,
			HasDeviceTokens:             hasDeviceTokens,
			ListRecoveryCodesAllowed:    h.Authentication.RecoveryCode.ListEnabled,
		}
		viewmodels.Embed(data, viewModel)

		h.Renderer.RenderHTML(w, r, TemplateWebSettingsMFAHTML, data)
		return nil
	})

	ctrl.PostAction("revoke_devices", func() error {
		if err := h.MFA.InvalidateAllDeviceTokens(userID); err != nil {
			return err
		}

		result := webapp.Result{RedirectURI: redirectURI}
		result.WriteResponse(w, r)
		return nil
	})

	ctrl.PostAction("setup_secondary_password", func() error {
		opts := webapp.SessionOptions{
			RedirectURI: redirectURI,
		}
		intent := intents.NewIntentAddAuthenticator(
			userID,
			interaction.AuthenticationStageSecondary,
			authn.AuthenticatorTypePassword,
		)

		result, err := ctrl.EntryPointPost(opts, intent, func() (input interface{}, err error) {
			return &InputCreateAuthenticator{}, nil
		})
		if err != nil {
			return err
		}

		result.WriteResponse(w, r)
		return nil
	})

	ctrl.PostAction("remove_secondary_password", func() error {
		opts := webapp.SessionOptions{
			RedirectURI: redirectURI,
		}
		intent := intents.NewIntentRemoveAuthenticator(userID)

		result, err := ctrl.EntryPointPost(opts, intent, func() (input interface{}, err error) {
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

	ctrl.PostAction("add_secondary_totp", func() error {
		opts := webapp.SessionOptions{
			RedirectURI: redirectURI,
		}
		intent := intents.NewIntentAddAuthenticator(
			userID,
			interaction.AuthenticationStageSecondary,
			authn.AuthenticatorTypeTOTP,
		)

		result, err := ctrl.EntryPointPost(opts, intent, func() (input interface{}, err error) {
			input = &InputCreateAuthenticator{}
			return
		})
		if err != nil {
			return err
		}

		result.WriteResponse(w, r)
		return nil
	})

	ctrl.PostAction("add_secondary_oob_otp_email", func() error {
		opts := webapp.SessionOptions{
			RedirectURI: redirectURI,
		}
		intent := intents.NewIntentAddAuthenticator(
			userID,
			interaction.AuthenticationStageSecondary,
			authn.AuthenticatorTypeOOBEmail,
		)

		result, err := ctrl.EntryPointPost(opts, intent, func() (input interface{}, err error) {
			input = &InputCreateAuthenticator{}
			return
		})
		if err != nil {
			return err
		}

		result.WriteResponse(w, r)
		return nil
	})

	ctrl.PostAction("add_secondary_oob_otp_sms", func() error {
		opts := webapp.SessionOptions{
			RedirectURI: redirectURI,
		}
		intent := intents.NewIntentAddAuthenticator(
			userID,
			interaction.AuthenticationStageSecondary,
			authn.AuthenticatorTypeOOBSMS,
		)

		result, err := ctrl.EntryPointPost(opts, intent, func() (input interface{}, err error) {
			input = &InputCreateAuthenticator{}
			return
		})
		if err != nil {
			return err
		}

		result.WriteResponse(w, r)
		return nil
	})
}
