package authflowv2

import (
	"math"
	"net/http"
	"net/url"

	"github.com/authgear/authgear-server/pkg/api/model"
	handlerwebapp "github.com/authgear/authgear-server/pkg/auth/handler/webapp"
	"github.com/authgear/authgear-server/pkg/auth/handler/webapp/viewmodels"
	"github.com/authgear/authgear-server/pkg/auth/webapp"
	"github.com/authgear/authgear-server/pkg/lib/accountmanagement"
	"github.com/authgear/authgear-server/pkg/lib/authn/otp"
	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/lib/infra/db/appdb"
	"github.com/authgear/authgear-server/pkg/lib/session"
	"github.com/authgear/authgear-server/pkg/util/clock"
	"github.com/authgear/authgear-server/pkg/util/httproute"
	"github.com/authgear/authgear-server/pkg/util/phone"
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
			"x_code": { "type": "string" }
		},
		"required": ["x_login_id_key", "x_code"]
	}
`)

var AuthflowV2SettingsIdentityResendPhoneSchema = validation.NewSimpleSchema(`
	{
		"type": "object",
		"properties": {
			"x_token": { "type": "string" }
		},
		"required": ["x_token"]
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
	Token      string

	CodeLength                     int
	MaskedClaimValue               string
	ResendCooldown                 int
	FailedAttemptRateLimitExceeded bool
}

type AuthflowV2SettingsIdentityVerifyPhoneHandler struct {
	Database            *appdb.Handle
	ControllerFactory   handlerwebapp.ControllerFactory
	BaseViewModel       *viewmodels.BaseViewModeler
	OTPCodeService      handlerwebapp.OTPCodeService
	Renderer            handlerwebapp.Renderer
	AccountManagement   accountmanagement.Service
	Clock               clock.Clock
	Config              *config.AppConfig
	AuthenticatorConfig *config.AuthenticatorConfig
}

func (h *AuthflowV2SettingsIdentityVerifyPhoneHandler) GetData(r *http.Request, rw http.ResponseWriter) (map[string]interface{}, error) {
	data := map[string]interface{}{}

	loginIDKey := r.Form.Get("q_login_id_key")
	tokenString := r.Form.Get("q_token")

	baseViewModel := h.BaseViewModel.ViewModel(r, rw)
	viewmodels.Embed(data, baseViewModel)

	s := session.GetSession(r.Context())
	token, err := h.AccountManagement.GetToken(s, tokenString)
	if err != nil {
		return nil, err
	}

	var channel model.AuthenticatorOOBChannel
	if h.AuthenticatorConfig.OOB.SMS.PhoneOTPMode.IsWhatsappEnabled() {
		channel = model.AuthenticatorOOBChannelWhatsapp
	} else {
		channel = model.AuthenticatorOOBChannelSMS
	}

	vm := AuthflowV2SettingsIdentityVerifyPhoneViewModel{
		LoginIDKey: loginIDKey,
		LoginID:    token.Identity.PhoneNumber,
		Token:      tokenString,

		CodeLength:       6,
		MaskedClaimValue: phone.Mask(token.Identity.PhoneNumber),
	}

	state, err := h.OTPCodeService.InspectState(otp.KindVerification(h.Config, channel), token.Identity.PhoneNumber)
	if err != nil {
		return nil, err
	}
	cooldown := int(math.Ceil(state.CanResendAt.Sub(h.Clock.NowUTC()).Seconds())) // Use ceil, because int conversion truncates decimal and can lead to Please Wait Before Resending error.
	if cooldown < 0 {
		vm.ResendCooldown = 0
	} else {
		vm.ResendCooldown = cooldown
	}

	vm.FailedAttemptRateLimitExceeded = state.TooManyAttempts

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
		err := AuthflowV2SettingsIdentityVerifyPhoneSchema.Validator().ValidateValue(handlerwebapp.FormToJSON(r.Form))
		if err != nil {
			return err
		}

		loginIDKey := r.Form.Get("x_login_id_key")
		token := r.Form.Get("q_token")
		code := r.Form.Get("x_code")

		s := session.GetSession(r.Context())
		_, err = h.AccountManagement.ResumeAddOrUpdateIdentityPhone(s, token, &accountmanagement.ResumeAddOrUpdateIdentityPhoneInput{
			LoginIDKey: loginIDKey,
			Code:       code,
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

	ctrl.PostAction("resend", func() error {
		err := AuthflowV2SettingsIdentityResendPhoneSchema.Validator().ValidateValue(handlerwebapp.FormToJSON(r.Form))
		if err != nil {
			return err
		}

		tokenString := r.Form.Get("x_token")
		err = h.AccountManagement.ResendOTPCode(session.GetSession(r.Context()), tokenString)
		if err != nil {
			return err
		}

		result := webapp.Result{}
		result.WriteResponse(w, r)
		return nil
	})
}
