package authflowv2

import (
	"context"
	"net/http"

	"net/url"

	"github.com/authgear/authgear-server/pkg/api/model"
	handlerwebapp "github.com/authgear/authgear-server/pkg/auth/handler/webapp"
	"github.com/authgear/authgear-server/pkg/auth/handler/webapp/viewmodels"
	"github.com/authgear/authgear-server/pkg/auth/webapp"
	"github.com/authgear/authgear-server/pkg/lib/accountmanagement"
	"github.com/authgear/authgear-server/pkg/lib/authn/authenticator"
	authenticatorservice "github.com/authgear/authgear-server/pkg/lib/authn/authenticator/service"
	"github.com/authgear/authgear-server/pkg/lib/infra/db/appdb"
	"github.com/authgear/authgear-server/pkg/lib/session"
	"github.com/authgear/authgear-server/pkg/util/httproute"
	"github.com/authgear/authgear-server/pkg/util/httputil"
	"github.com/authgear/authgear-server/pkg/util/template"
)

var TemplateWebSettingsTOTPHTML = template.RegisterHTML(
	"web/authflowv2/settings_mfa_totp.html",
	handlerwebapp.SettingsComponents...,
)

func ConfigureAuthflowV2SettingsTOTPRoute(route httproute.Route) httproute.Route {
	return route.
		WithMethods("OPTIONS", "POST", "GET").
		WithPathPattern("/settings/mfa/totp")
}

type AuthflowV2SettingsTOTPViewModel struct {
	TOTPAuthenticators []*authenticator.TOTP
}

type AuthflowV2SettingsTOTPHandler struct {
	Database          *appdb.Handle
	ControllerFactory handlerwebapp.ControllerFactory
	BaseViewModel     *viewmodels.BaseViewModeler
	SettingsViewModel *viewmodels.SettingsViewModeler
	Renderer          handlerwebapp.Renderer

	AccountManagement *accountmanagement.Service
	Authenticators    authenticatorservice.Service
}

func (h *AuthflowV2SettingsTOTPHandler) GetData(ctx context.Context, w http.ResponseWriter, r *http.Request) (map[string]interface{}, error) {
	data := map[string]interface{}{}
	userID := session.GetUserID(ctx)

	// BaseViewModel
	baseViewModel := h.BaseViewModel.ViewModel(r, w)
	viewmodels.Embed(data, baseViewModel)

	// SettingsViewModel
	settingsViewModel, err := h.SettingsViewModel.ViewModel(ctx, *userID)
	if err != nil {
		return nil, err
	}
	viewmodels.Embed(data, *settingsViewModel)

	authenticators, err := h.Authenticators.List(
		ctx,
		*userID,
		authenticator.KeepKind(authenticator.KindSecondary),
		authenticator.KeepType(model.AuthenticatorTypeTOTP),
	)

	if err != nil {
		return nil, err
	}

	var totpAuthenticators []*authenticator.TOTP
	for _, a := range authenticators {
		totpAuthenticators = append(totpAuthenticators, a.TOTP)
	}

	vm := AuthflowV2SettingsTOTPViewModel{
		TOTPAuthenticators: totpAuthenticators,
	}
	viewmodels.Embed(data, vm)

	return data, nil
}

func (h *AuthflowV2SettingsTOTPHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ctrl, err := h.ControllerFactory.New(r, w)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	defer ctrl.ServeWithoutDBTx(r.Context())

	ctrl.Get(func(ctx context.Context) error {
		var data map[string]interface{}
		err := h.Database.WithTx(ctx, func(ctx context.Context) error {
			data, err = h.GetData(ctx, w, r)
			return err
		})
		if err != nil {
			return err
		}

		h.Renderer.RenderHTML(w, r, TemplateWebSettingsTOTPHTML, data)
		return nil
	})

	ctrl.PostAction("create_totp", func(ctx context.Context) error {
		s := session.GetSession(ctx)
		output, err := h.AccountManagement.StartAddTOTPAuthenticator(ctx, s, &accountmanagement.StartAddTOTPAuthenticatorInput{})
		if err != nil {
			return err
		}

		redirectURI, err := url.Parse(AuthflowV2RouteSettingsMFACreateTOTP)
		if err != nil {
			return err
		}

		q := redirectURI.Query()
		q.Set("q_token", output.Token)

		redirectURI.RawQuery = q.Encode()

		result := webapp.Result{RedirectURI: redirectURI.String()}
		result.WriteResponse(w, r)

		return nil
	})

	ctrl.PostAction("remove", func(ctx context.Context) error {
		authenticatorID := r.Form.Get("x_authenticator_id")

		s := session.GetSession(ctx)

		input := &accountmanagement.DeleteTOTPAuthenticatorInput{
			AuthenticatorID: authenticatorID,
		}
		_, err = h.AccountManagement.DeleteTOTPAuthenticator(ctx, s, input)
		if err != nil {
			return err
		}

		redirectURI := httputil.HostRelative(r.URL).String()
		result := webapp.Result{RedirectURI: redirectURI}
		result.WriteResponse(w, r)

		return nil
	})
}
