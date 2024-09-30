package authflowv2

import (
	"fmt"
	htmltemplate "html/template"
	"net/http"
	"net/url"
	"time"

	handlerwebapp "github.com/authgear/authgear-server/pkg/auth/handler/webapp"
	"github.com/authgear/authgear-server/pkg/auth/handler/webapp/viewmodels"
	"github.com/authgear/authgear-server/pkg/auth/webapp"

	"github.com/authgear/authgear-server/pkg/lib/accountmanagement"
	"github.com/authgear/authgear-server/pkg/lib/infra/db/appdb"
	"github.com/authgear/authgear-server/pkg/lib/session"
	"github.com/authgear/authgear-server/pkg/util/clock"
	"github.com/authgear/authgear-server/pkg/util/httproute"
	"github.com/authgear/authgear-server/pkg/util/template"
	"github.com/authgear/authgear-server/pkg/util/validation"
)

var TemplateWebSettingsMFAEnterTOTPHTML = template.RegisterHTML(
	"web/authflowv2/settings_mfa_enter_totp.html",
	handlerwebapp.SettingsComponents...,
)

var AuthflowV2SettingsMFAEnterTOTPSchema = validation.NewSimpleSchema(`
	{
		"type": "object",
		"properties": {
			"x_code": { "type": "string" }
		},
		"required": ["x_code"]
	}
`)

func ConfigureAuthflowV2SettingsMFAEnterTOTPRoute(route httproute.Route) httproute.Route {
	return route.
		WithMethods("OPTIONS", "POST", "GET").
		WithPathPattern(AuthflowV2RouteSettingsMFAEnterTOTP)
}

type AuthflowV2SettingsMFAEnterTOTPViewModel struct {
	Secret   string
	ImageURI htmltemplate.URL
}

type AuthflowV2SettingsMFAEnterTOTPHandler struct {
	Database          *appdb.Handle
	ControllerFactory handlerwebapp.ControllerFactory
	BaseViewModel     *viewmodels.BaseViewModeler
	SettingsViewModel *viewmodels.SettingsViewModeler
	Renderer          handlerwebapp.Renderer
	Clock             clock.Clock

	AccountManagement *accountmanagement.Service
}

func (h *AuthflowV2SettingsMFAEnterTOTPHandler) GetData(r *http.Request, rw http.ResponseWriter) (map[string]interface{}, error) {
	data := make(map[string]interface{})

	baseViewModel := h.BaseViewModel.ViewModel(r, rw)
	viewmodels.Embed(data, baseViewModel)

	return data, nil
}

func (h *AuthflowV2SettingsMFAEnterTOTPHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ctrl, err := h.ControllerFactory.New(r, w)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	defer ctrl.ServeWithoutDBTx()

	ctrl.Get(func() error {
		s := session.GetSession(r.Context())

		tokenString := r.Form.Get("q_token")
		_, err := h.AccountManagement.GetToken(s, tokenString)
		if err != nil {
			return err
		}

		data, err := h.GetData(r, w)
		if err != nil {
			return err
		}

		h.Renderer.RenderHTML(w, r, TemplateWebSettingsMFAEnterTOTPHTML, data)
		return nil
	})

	ctrl.PostAction("submit", func() error {
		err := AuthflowV2SettingsMFAEnterTOTPSchema.Validator().ValidateValue(handlerwebapp.FormToJSON(r.Form))
		if err != nil {
			return err
		}

		s := session.GetSession(r.Context())

		tokenString := r.Form.Get("q_token")
		code := r.Form.Get("x_code")

		now := h.Clock.NowUTC()
		displayName := fmt.Sprintf("TOTP @ %s", now.Format(time.RFC3339))

		output, err := h.AccountManagement.ResumeAddTOTPAuthenticator(s, tokenString, &accountmanagement.ResumeAddTOTPAuthenticatorInput{
			DisplayName: displayName,
			Code:        code,
		})
		if err != nil {
			return err
		}

		redirectURI, err := url.Parse(AuthflowV2RouteSettingsMFAViewRecoveryCode)
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
}
