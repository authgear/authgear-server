package webapp

import (
	"net/http"

	"github.com/authgear/authgear-server/pkg/api/model"
	"github.com/authgear/authgear-server/pkg/auth/handler/webapp/viewmodels"
	"github.com/authgear/authgear-server/pkg/auth/webapp"
	"github.com/authgear/authgear-server/pkg/lib/authn"
	"github.com/authgear/authgear-server/pkg/lib/authn/mfa"
	"github.com/authgear/authgear-server/pkg/lib/interaction/intents"
	"github.com/authgear/authgear-server/pkg/util/httproute"
	"github.com/authgear/authgear-server/pkg/util/httputil"
	"github.com/authgear/authgear-server/pkg/util/template"
)

var TemplateWebSettingsMFAHTML = template.RegisterHTML(
	"web/settings_mfa.html",
	Components...,
)

func ConfigureSettingsMFARoute(route httproute.Route) httproute.Route {
	return route.
		WithMethods("OPTIONS", "POST", "GET").
		WithPathPattern("/settings/mfa")
}

type SettingsMFAService interface {
	ListRecoveryCodes(userID string) ([]*mfa.RecoveryCode, error)
	InvalidateAllDeviceTokens(userID string) error
}

type SettingsMFAHandler struct {
	ControllerFactory ControllerFactory
	BaseViewModel     *viewmodels.BaseViewModeler
	SettingsViewModel *viewmodels.SettingsViewModeler
	Renderer          Renderer
	MFA               SettingsMFAService
}

func (h *SettingsMFAHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ctrl, err := h.ControllerFactory.New(r, w)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	defer ctrl.ServeWithDBTx()

	redirectURI := httputil.HostRelative(r.URL).String()
	authenticatorID := r.Form.Get("x_authenticator_id")
	userID := ctrl.RequireUserID()

	ctrl.Get(func() error {
		data := map[string]interface{}{}

		baseViewModel := h.BaseViewModel.ViewModel(r, w)
		viewmodels.Embed(data, baseViewModel)

		viewModelPtr, err := h.SettingsViewModel.ViewModel(userID)
		if err != nil {
			return err
		}
		viewmodels.Embed(data, *viewModelPtr)

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
			authn.AuthenticationStageSecondary,
			model.AuthenticatorTypePassword,
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
				Type: model.AuthenticatorTypePassword,
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
			authn.AuthenticationStageSecondary,
			model.AuthenticatorTypeTOTP,
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
			authn.AuthenticationStageSecondary,
			model.AuthenticatorTypeOOBEmail,
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
			authn.AuthenticationStageSecondary,
			model.AuthenticatorTypeOOBSMS,
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
