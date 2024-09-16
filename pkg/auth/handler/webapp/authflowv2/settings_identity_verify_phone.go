package authflowv2

import (
	"fmt"
	"net/http"
	"net/url"

	"github.com/authgear/authgear-server/pkg/api/model"
	handlerwebapp "github.com/authgear/authgear-server/pkg/auth/handler/webapp"
	"github.com/authgear/authgear-server/pkg/auth/handler/webapp/viewmodels"
	"github.com/authgear/authgear-server/pkg/auth/webapp"
	"github.com/authgear/authgear-server/pkg/lib/accountmanagement"
	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/lib/session"
	"github.com/authgear/authgear-server/pkg/util/httproute"
	"github.com/authgear/authgear-server/pkg/util/template"
	"github.com/authgear/authgear-server/pkg/util/validation"
)

var TemplateWebSettingsIdentityVerifyPhoneHTML = template.RegisterHTML(
	"web/authflowv2/settings_identity_verify_phone.html",
	handlerwebapp.SettingsComponents...,
)

var AuthflowV2SettingsIdentityVerifyPhoneSchema = validation.NewSimpleSchema(`
	{
		"type": "object",
		"properties": {
			"x_login_id_key": { "type": "string" },
			"x_login_id": { "type": "string" },
			"x_token": { "type": "string" },
			"x_code": { "type": "string" }
		},
		"required": ["x_login_id_key", "x_login_id", "x_token", "x_code"]
	}
`)

func ConfigureAuthflowV2SettingsIdentityVerifyPhoneRoute(route httproute.Route) httproute.Route {
	return route.
		WithMethods("OPTIONS", "POST", "GET").
		WithPathPattern(AuthflowV2RouteSettingsIdentityVerifyPhone)
}

type AuthflowV2SettingsIdentityVerifyPhoneViewModel struct {
	LoginIDKey string
	LoginID    string
	TokenID    string

	CodeLength       int
	MaskedClaimValue string
}

type AuthflowV2SettingsIdentityVerifyPhoneHandler struct {
	ControllerFactory   handlerwebapp.ControllerFactory
	BaseViewModel       *viewmodels.BaseViewModeler
	Renderer            handlerwebapp.Renderer
	AuthenticatorConfig *config.AuthenticatorConfig
	AccountManagement   accountmanagement.Service
}

func (h *AuthflowV2SettingsIdentityVerifyPhoneHandler) GetData(r *http.Request, rw http.ResponseWriter) (map[string]interface{}, error) {
	data := map[string]interface{}{}

	loginIDKey := r.Form.Get("q_login_id_key")
	tokenID := r.Form.Get("q_token")

	baseViewModel := h.BaseViewModel.ViewModel(r, rw)
	viewmodels.Embed(data, baseViewModel)

	s := session.GetSession(r.Context())
	output, err := h.AccountManagement.ResumeCreatingPhoneNumberIdentityWithVerification(s, &accountmanagement.ResumeAddingIdentityWithVerificationInput{
		Token: tokenID,
	})
	if err != nil {
		return nil, err
	}

	vm := AuthflowV2SettingsIdentityVerifyPhoneViewModel{
		LoginIDKey: loginIDKey,
		LoginID:    output.LoginID,
		TokenID:    tokenID,

		CodeLength:       6,
		MaskedClaimValue: output.LoginID,
	}
	viewmodels.Embed(data, vm)

	return data, nil
}

func (h *AuthflowV2SettingsIdentityVerifyPhoneHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ctrl, err := h.ControllerFactory.New(r, w)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	defer ctrl.ServeWithoutDBTx()

	ctrl.Get(func() error {
		data, err := h.GetData(r, w)
		if err != nil {
			return err
		}

		h.Renderer.RenderHTML(w, r, TemplateWebSettingsIdentityVerifyPhoneHTML, data)
		return nil
	})

	ctrl.PostAction("submit", func() error {

		fmt.Printf("%s\n", r.Form)

		err := AuthflowV2SettingsIdentityVerifyPhoneSchema.Validator().ValidateValue(handlerwebapp.FormToJSON(r.Form))
		if err != nil {
			return err
		}

		loginID := r.Form.Get("x_login_id")
		loginIDKey := r.Form.Get("x_login_id_key")
		tokenID := r.Form.Get("x_token")

		code := r.Form.Get("x_code")

		var channel model.AuthenticatorOOBChannel
		if h.AuthenticatorConfig.OOB.SMS.PhoneOTPMode.IsWhatsappEnabled() {
			channel = model.AuthenticatorOOBChannelWhatsapp
		} else {
			channel = model.AuthenticatorOOBChannelSMS
		}

		s := session.GetSession(r.Context())
		_, err = h.AccountManagement.AddIdentityPhoneNumberWithVerification(s, &accountmanagement.AddIdentityPhoneNumberWithVerificationInput{
			LoginID:    loginID,
			LoginIDKey: loginIDKey,
			Code:       code,
			Token:      tokenID,
			Channel:    channel,
		})
		if err != nil {
			return err
		}

		redirectURI, err := url.Parse(AuthflowV2RouteSettingsIdentityListPhone)
		if err != nil {
			return err
		}
		q := redirectURI.Query()
		q.Set("q_login_id_key", loginIDKey)
		redirectURI.RawQuery = q.Encode()

		result := webapp.Result{RedirectURI: redirectURI.String()}
		result.WriteResponse(w, r)
		return nil
	})
}
