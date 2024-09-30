package authflowv2

import (
	"net/http"

	handlerwebapp "github.com/authgear/authgear-server/pkg/auth/handler/webapp"
	"github.com/authgear/authgear-server/pkg/auth/handler/webapp/viewmodels"
	"github.com/authgear/authgear-server/pkg/auth/webapp"
	"github.com/authgear/authgear-server/pkg/lib/accountmanagement"
	"github.com/authgear/authgear-server/pkg/lib/session"
	"github.com/authgear/authgear-server/pkg/util/httproute"
	"github.com/authgear/authgear-server/pkg/util/template"
)

var TemplateWebSettingsMFAViewRecoveryCodeHTML = template.RegisterHTML(
	"web/authflowv2/settings_mfa_view_recovery_code.html",
	handlerwebapp.SettingsComponents...,
)

func ConfigureAuthflowV2SettingsMFAViewRecoveryCodeRoute(route httproute.Route) httproute.Route {
	return route.
		WithMethods("OPTIONS", "POST", "GET").
		WithPathPattern(AuthflowV2RouteSettingsMFAViewRecoveryCode)
}

type AuthflowV2SettingsMFAViewRecoveryCodeViewModel struct {
	RecoveryCodes []string
}

type AuthflowV2SettingsMFAViewRecoveryCodeHandler struct {
	ControllerFactory handlerwebapp.ControllerFactory
	BaseViewModel     *viewmodels.BaseViewModeler
	Renderer          handlerwebapp.Renderer

	AccountManagement *accountmanagement.Service
}

func (h *AuthflowV2SettingsMFAViewRecoveryCodeHandler) GetData(r *http.Request, rw http.ResponseWriter, recoveryCodes []string) (map[string]interface{}, error) {
	data := map[string]interface{}{}

	baseViewModel := h.BaseViewModel.ViewModel(r, rw)
	viewmodels.Embed(data, baseViewModel)

	screenViewModel := AuthflowV2SettingsMFAViewRecoveryCodeViewModel{
		RecoveryCodes: handlerwebapp.FormatRecoveryCodes(recoveryCodes),
	}
	viewmodels.Embed(data, screenViewModel)

	return data, nil
}

func (h *AuthflowV2SettingsMFAViewRecoveryCodeHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ctrl, err := h.ControllerFactory.New(r, w)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	defer ctrl.ServeWithoutDBTx()

	ctrl.Get(func() error {
		s := session.GetSession(r.Context())

		tokenString := r.Form.Get("q_token")
		token, err := h.AccountManagement.GetToken(s, tokenString)
		if err != nil {
			return err
		}

		data, err := h.GetData(r, w, token.Authenticator.RecoveryCodes)
		if err != nil {
			return err
		}

		h.Renderer.RenderHTML(w, r, TemplateWebSettingsMFAViewRecoveryCodeHTML, data)
		return nil
	})

	ctrl.PostAction("download", func() error {
		s := session.GetSession(r.Context())

		tokenString := r.Form.Get("q_token")
		token, err := h.AccountManagement.GetToken(s, tokenString)
		if err != nil {
			return err
		}

		data, err := h.GetData(r, w, token.Authenticator.RecoveryCodes)
		if err != nil {
			return err
		}

		handlerwebapp.SetRecoveryCodeAttachmentHeaders(w)
		h.Renderer.Render(w, r, handlerwebapp.TemplateWebDownloadRecoveryCodeTXT, data)
		return nil
	})

	ctrl.PostAction("proceed", func() error {
		s := session.GetSession(r.Context())

		tokenString := r.Form.Get("q_token")
		_, err := h.AccountManagement.GetToken(s, tokenString)
		if err != nil {
			return err
		}

		_, err = h.AccountManagement.FinishAddTOTPAuthenticator(s, tokenString, &accountmanagement.FinishAddTOTPAuthenticatorInput{})
		if err != nil {
			return err
		}

		result := webapp.Result{RedirectURI: AuthflowV2RouteSettingsMFA}
		result.WriteResponse(w, r)

		return nil
	})
}
