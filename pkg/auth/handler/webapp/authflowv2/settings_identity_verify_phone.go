package authflowv2

import (
	"context"
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
			"q_token": { "type": "string" }
		},
		"required": ["q_token"]
	}
`)

func ConfigureAuthflowV2SettingsIdentityVerifyPhoneRoute(route httproute.Route) httproute.Route {
	return route.
		WithMethods("OPTIONS", "POST", "GET").
		WithPathPattern(AuthflowV2RouteSettingsIdentityVerifyPhone)
}

type AuthflowV2SettingsIdentityVerifyPhoneViewModel struct {
	Channels   []model.AuthenticatorOOBChannel
	Channel    model.AuthenticatorOOBChannel
	LoginIDKey string
	LoginID    string
	IdentityID string
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

func (h *AuthflowV2SettingsIdentityVerifyPhoneHandler) GetData(ctx context.Context, r *http.Request, rw http.ResponseWriter) (map[string]interface{}, error) {
	data := map[string]interface{}{}

	loginIDKey := r.Form.Get("q_login_id_key")
	tokenString := r.Form.Get("q_token")

	baseViewModel := h.BaseViewModel.ViewModel(r, rw)
	viewmodels.Embed(data, baseViewModel)

	s := session.GetSession(ctx)
	token, err := h.AccountManagement.GetToken(ctx, s, tokenString)
	if err != nil {
		return nil, err
	}

	oobConfig := h.AuthenticatorConfig.OOB
	channel := model.AuthenticatorOOBChannel(token.Identity.Channel)

	var channels []model.AuthenticatorOOBChannel
	switch oobConfig.SMS.PhoneOTPMode {
	case config.AuthenticatorPhoneOTPModeSMSOnly:
		channels = append(channels, model.AuthenticatorOOBChannelSMS)
	case config.AuthenticatorPhoneOTPModeWhatsappOnly:
		channels = append(channels, model.AuthenticatorOOBChannelWhatsapp)
	case config.AuthenticatorPhoneOTPModeWhatsappSMS:
		channels = append(channels, model.AuthenticatorOOBChannelWhatsapp)
		channels = append(channels, model.AuthenticatorOOBChannelSMS)
	}

	vm := AuthflowV2SettingsIdentityVerifyPhoneViewModel{
		Channels:   channels,
		Channel:    channel,
		LoginIDKey: loginIDKey,
		LoginID:    token.Identity.PhoneNumber,
		IdentityID: token.Identity.IdentityID,
		Token:      tokenString,

		CodeLength:       6,
		MaskedClaimValue: phone.Mask(token.Identity.PhoneNumber),
	}

	state, err := h.OTPCodeService.InspectState(ctx, otp.KindVerification(h.Config, channel), token.Identity.PhoneNumber)
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
	defer ctrl.ServeWithoutDBTx(r.Context())

	ctrl.Get(func(ctx context.Context) error {
		data, err := h.GetData(ctx, r, w)
		if err != nil {
			return err
		}

		h.Renderer.RenderHTML(w, r, TemplateWebSettingsIdentityVerifyPhoneHTML, data)
		return nil
	})

	ctrl.PostAction("submit", func(ctx context.Context) error {
		err := AuthflowV2SettingsIdentityVerifyPhoneSchema.Validator().ValidateValue(handlerwebapp.FormToJSON(r.Form))
		if err != nil {
			return err
		}

		loginIDKey := r.Form.Get("x_login_id_key")
		token := r.Form.Get("q_token")
		code := r.Form.Get("x_code")

		s := session.GetSession(ctx)
		_, err = h.AccountManagement.ResumeAddOrUpdateIdentityPhone(ctx, s, token, &accountmanagement.ResumeAddOrUpdateIdentityPhoneInput{
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

	ctrl.PostAction("resend", func(ctx context.Context) error {
		err := AuthflowV2SettingsIdentityResendPhoneSchema.Validator().ValidateValue(handlerwebapp.FormToJSON(r.Form))
		if err != nil {
			return err
		}

		tokenString := r.Form.Get("q_token")
		err = h.AccountManagement.ResendOTPCode(ctx, session.GetSession(ctx), tokenString)
		if err != nil {
			return err
		}

		result := webapp.Result{}
		result.WriteResponse(w, r)
		return nil
	})
}
