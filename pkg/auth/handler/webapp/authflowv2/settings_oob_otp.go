package authflowv2

import (
	"context"
	"net/http"

	"github.com/authgear/authgear-server/pkg/api/model"
	handlerwebapp "github.com/authgear/authgear-server/pkg/auth/handler/webapp"
	"github.com/authgear/authgear-server/pkg/auth/handler/webapp/viewmodels"
	"github.com/authgear/authgear-server/pkg/auth/webapp"
	"github.com/authgear/authgear-server/pkg/lib/accountmanagement"
	"github.com/authgear/authgear-server/pkg/lib/authn/authenticator"
	authenticatorservice "github.com/authgear/authgear-server/pkg/lib/authn/authenticator/service"
	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/lib/infra/db/appdb"
	"github.com/authgear/authgear-server/pkg/lib/session"
	"github.com/authgear/authgear-server/pkg/util/httproute"
	"github.com/authgear/authgear-server/pkg/util/httputil"
	"github.com/authgear/authgear-server/pkg/util/template"
)

var TemplateWebSettingsOOBOTPHTML = template.RegisterHTML(
	"web/authflowv2/settings_oob_otp.html",
	handlerwebapp.SettingsComponents...,
)

func ConfigureAuthflowV2SettingsOOBOTPRoute(route httproute.Route) httproute.Route {
	return route.
		WithMethods("OPTIONS", "POST", "GET").
		WithPathPattern("/settings/mfa/oob_otp_:channel")
}

type AuthflowV2SettingsOOBOTPViewModel struct {
	OOBOTPType           model.AuthenticatorType
	OOBOTPAuthenticators []*authenticator.OOBOTP
	OOBOTPChannel        model.AuthenticatorOOBChannel
}

type AuthflowV2SettingsOOBOTPHandler struct {
	Database             *appdb.Handle
	ControllerFactory    handlerwebapp.ControllerFactory
	BaseViewModel        *viewmodels.BaseViewModeler
	SettingsViewModel    *viewmodels.SettingsViewModeler
	Renderer             handlerwebapp.Renderer
	AuthenticatiorConfig *config.AuthenticatorConfig
	AccountManagement    *accountmanagement.Service
	Authenticators       authenticatorservice.Service
}

func (h *AuthflowV2SettingsOOBOTPHandler) GetData(ctx context.Context, w http.ResponseWriter, r *http.Request) (map[string]interface{}, error) {
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

	email_or_sms := httproute.GetParam(r, "channel")

	t, err := model.ParseOOBAuthenticatorType(email_or_sms)
	if err != nil {
		return nil, err
	}
	authenticators, err := h.Authenticators.List(
		ctx,
		*userID,
		authenticator.KeepKind(authenticator.KindSecondary),
		authenticator.KeepType(t),
	)
	if err != nil {
		return nil, err
	}

	var OOBOTPAuthenticators []*authenticator.OOBOTP
	for _, a := range authenticators {
		OOBOTPAuthenticators = append(OOBOTPAuthenticators, a.OOBOTP)
	}

	channel := h.Authenticators.Config.Authenticator.OOB.GetDefaultChannelFor(t)

	vm := AuthflowV2SettingsOOBOTPViewModel{
		OOBOTPType:           t,
		OOBOTPAuthenticators: OOBOTPAuthenticators,
		OOBOTPChannel:        channel,
	}
	viewmodels.Embed(data, vm)

	return data, nil
}

func (h *AuthflowV2SettingsOOBOTPHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
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

		h.Renderer.RenderHTML(w, r, TemplateWebSettingsOOBOTPHTML, data)
		return nil
	})

	ctrl.PostAction("remove", func(ctx context.Context) error {
		authenticatorID := r.Form.Get("x_authenticator_id")

		s := session.GetSession(ctx)

		input := &accountmanagement.DeleteOOBOTPAuthenticatorInput{
			AuthenticatorID: authenticatorID,
		}
		_, err = h.AccountManagement.DeleteOOBOTPAuthenticator(ctx, s, input)
		if err != nil {
			return err
		}

		redirectURI := httputil.HostRelative(r.URL).String()
		result := webapp.Result{RedirectURI: redirectURI}
		result.WriteResponse(w, r)

		return nil
	})
}
