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
	"github.com/authgear/authgear-server/pkg/lib/infra/mail"
	"github.com/authgear/authgear-server/pkg/lib/session"
	"github.com/authgear/authgear-server/pkg/util/clock"
	"github.com/authgear/authgear-server/pkg/util/httproute"
	"github.com/authgear/authgear-server/pkg/util/template"
	"github.com/authgear/authgear-server/pkg/util/validation"
)

var TemplateWebSettingsIdentityVerifyEmailHTML = template.RegisterHTML(
	"web/authflowv2/settings_identity_verify_email.html",
	handlerwebapp.SettingsComponents...,
)

var AuthflowV2SettingsIdentityVerifyEmailSchema = validation.NewSimpleSchema(`
	{
		"type": "object",
		"properties": {
			"x_login_id_key": { "type": "string" },
			"x_token": { "type": "string" },
			"x_code": { "type": "string" }
		},
		"required": ["x_login_id_key", "x_token", "x_code"]
	}
`)

var AuthflowV2SettingsIdentityResendEmailSchema = validation.NewSimpleSchema(`
	{
		"type": "object",
		"properties": {
			"x_token": { "type": "string" }
		},
		"required": ["x_token"]
	}
`)

func ConfigureAuthflowV2SettingsIdentityVerifyEmailRoute(route httproute.Route) httproute.Route {
	return route.
		WithMethods("OPTIONS", "POST", "GET").
		WithPathPattern(AuthflowV2RouteSettingsIdentityVerifyEmail)
}

type AuthflowV2SettingsIdentityVerifyEmailViewModel struct {
	LoginIDKey string
	LoginID    string
	Token      string

	CodeLength                     int
	MaskedClaimValue               string
	ResendCooldown                 int
	FailedAttemptRateLimitExceeded bool
}

type AuthflowV2SettingsIdentityVerifyEmailHandler struct {
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

func (h *AuthflowV2SettingsIdentityVerifyEmailHandler) GetData(r *http.Request, rw http.ResponseWriter) (map[string]interface{}, error) {
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

	vm := AuthflowV2SettingsIdentityVerifyEmailViewModel{
		LoginIDKey: loginIDKey,
		LoginID:    token.Identity.Email,
		Token:      tokenString,

		CodeLength:       6,
		MaskedClaimValue: mail.MaskAddress(token.Identity.Email),
	}

	state, err := h.OTPCodeService.InspectState(otp.KindVerification(h.Config, model.AuthenticatorOOBChannelEmail), token.Identity.Email)
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

func (h *AuthflowV2SettingsIdentityVerifyEmailHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
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

		h.Renderer.RenderHTML(w, r, TemplateWebSettingsIdentityVerifyEmailHTML, data)
		return nil
	})

	ctrl.PostAction("submit", func() error {
		err := AuthflowV2SettingsIdentityVerifyEmailSchema.Validator().ValidateValue(handlerwebapp.FormToJSON(r.Form))
		if err != nil {
			return err
		}

		loginIDKey := r.Form.Get("x_login_id_key")
		tokenString := r.Form.Get("x_token")

		code := r.Form.Get("x_code")

		s := session.GetSession(r.Context())
		_, err = h.AccountManagement.ResumeAddOrUpdateIdentityEmail(s, tokenString, &accountmanagement.ResumeAddOrUpdateIdentityEmailInput{
			LoginIDKey: loginIDKey,
			Code:       code,
		})
		if err != nil {
			return err
		}

		redirectURI, err := url.Parse(AuthflowV2RouteSettingsIdentityListEmail)
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
		err := AuthflowV2SettingsIdentityResendEmailSchema.Validator().ValidateValue(handlerwebapp.FormToJSON(r.Form))
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
