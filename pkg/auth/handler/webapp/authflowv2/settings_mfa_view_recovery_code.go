package authflowv2

import (
	"net/http"

	handlerwebapp "github.com/authgear/authgear-server/pkg/auth/handler/webapp"
	"github.com/authgear/authgear-server/pkg/auth/handler/webapp/viewmodels"
	"github.com/authgear/authgear-server/pkg/auth/webapp"
	"github.com/authgear/authgear-server/pkg/lib/accountmanagement"
	"github.com/authgear/authgear-server/pkg/lib/infra/db/appdb"
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
	CanProceed    bool
	CanRegenerate bool
}

type AuthflowV2SettingsMFAViewRecoveryCodeHandler struct {
	Database          *appdb.Handle
	ControllerFactory handlerwebapp.ControllerFactory
	BaseViewModel     *viewmodels.BaseViewModeler
	Renderer          handlerwebapp.Renderer

	AccountManagement *accountmanagement.Service
	MFA               handlerwebapp.SettingsMFAService
}

func (h *AuthflowV2SettingsMFAViewRecoveryCodeHandler) GetData(r *http.Request, rw http.ResponseWriter) (map[string]interface{}, error) {
	s := session.GetSession(r.Context())
	userID := session.GetUserID(r.Context())

	data := map[string]interface{}{}

	baseViewModel := h.BaseViewModel.ViewModel(r, rw)
	viewmodels.Embed(data, baseViewModel)

	var tokenAuthenticator *accountmanagement.TokenAuthenticator
	var recoveryCodesString []string

	tokenString := r.Form.Get("q_token")
	if tokenString != "" {
		token, err := h.AccountManagement.GetToken(s, tokenString)
		if err != nil {
			return nil, err
		}

		tokenAuthenticator = token.Authenticator
		recoveryCodesString = tokenAuthenticator.RecoveryCodes
	} else {
		recoveryCodes, err := h.MFA.ListRecoveryCodes(*userID)
		if err != nil {
			return nil, err
		}

		recoveryCodesString = make([]string, len(recoveryCodes))
		for i, code := range recoveryCodes {
			recoveryCodesString[i] = code.Code
		}
	}

	screenViewModel := AuthflowV2SettingsMFAViewRecoveryCodeViewModel{
		RecoveryCodes: handlerwebapp.FormatRecoveryCodes(recoveryCodesString),
		CanProceed:    tokenAuthenticator != nil,
		CanRegenerate: tokenAuthenticator == nil,
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
		var data map[string]interface{}
		err := h.Database.WithTx(func() error {
			data, err = h.GetData(r, w)
			if err != nil {
				return err
			}
			return nil
		})
		if err != nil {
			return err
		}

		h.Renderer.RenderHTML(w, r, TemplateWebSettingsMFAViewRecoveryCodeHTML, data)
		return nil
	})

	ctrl.PostAction("download", func() error {
		var data map[string]interface{}
		err := h.Database.WithTx(func() error {
			data, err = h.GetData(r, w)
			if err != nil {
				return err
			}
			return nil
		})
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
		token, err := h.AccountManagement.GetToken(s, tokenString)
		if err != nil {
			return err
		}

		if token.Authenticator.TOTPVerified {
			_, err = h.AccountManagement.FinishAddTOTPAuthenticator(s, tokenString, &accountmanagement.FinishAddTOTPAuthenticatorInput{})
			if err != nil {
				return err
			}
		} else if token.Authenticator.OOBOTPVerified {
			_, err = h.AccountManagement.FinishAddOOBOTPAuthenticator(s, tokenString, &accountmanagement.FinishAddOOBOTPAuthenticatorInput{})
			if err != nil {
				return err
			}
		} else {
			panic("authflowv2: unexpected authenticator type")
		}

		result := webapp.Result{RedirectURI: AuthflowV2RouteSettingsMFA}
		result.WriteResponse(w, r)

		return nil
	})

	ctrl.PostAction("regenerate", func() error {
		s := session.GetSession(r.Context())

		_, err := h.AccountManagement.GenerateRecoveryCodes(s, &accountmanagement.GenerateRecoveryCodesInput{})
		if err != nil {
			return err
		}

		result := webapp.Result{RedirectURI: AuthflowV2RouteSettingsMFAViewRecoveryCode}
		result.WriteResponse(w, r)

		return nil
	})
}
