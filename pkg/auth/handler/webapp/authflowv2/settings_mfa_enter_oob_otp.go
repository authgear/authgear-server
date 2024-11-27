package authflowv2

import (
	"context"
	"math"
	"net/http"

	"net/url"
	"time"

	"github.com/authgear/authgear-server/pkg/api/model"
	handlerwebapp "github.com/authgear/authgear-server/pkg/auth/handler/webapp"
	"github.com/authgear/authgear-server/pkg/auth/handler/webapp/viewmodels"
	"github.com/authgear/authgear-server/pkg/auth/webapp"
	"github.com/authgear/authgear-server/pkg/lib/accountmanagement"
	"github.com/authgear/authgear-server/pkg/lib/authn/otp"
	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/lib/infra/mail"
	"github.com/authgear/authgear-server/pkg/lib/session"
	"github.com/authgear/authgear-server/pkg/util/clock"
	"github.com/authgear/authgear-server/pkg/util/httproute"
	"github.com/authgear/authgear-server/pkg/util/phone"
	"github.com/authgear/authgear-server/pkg/util/template"
	"github.com/authgear/authgear-server/pkg/util/validation"
)

var TemplateWebSettingsMFAEnterOOBOTPHTML = template.RegisterHTML(
	"web/authflowv2/settings_mfa_enter_oob_otp.html",
	handlerwebapp.SettingsComponents...,
)

var AuthflowV2SettingsMFAEnterOOBOTP = validation.NewSimpleSchema(`
	{
		"type": "object",
		"properties": {
			"x_code": {
				"type": "string",
				"format": "x_oob_otp_code"
			}
		},
		"required": ["x_code"]
	}
`)

var AuthflowV2SettingsIdentityResendOOBOTPSchema = validation.NewSimpleSchema(`
	{
		"type": "object",
		"properties": {
			"q_token": { "type": "string" }
		},
		"required": ["q_token"]
	}
`)

func ConfigureAuthflowV2SettingsMFAEnterOOBOTPRoute(route httproute.Route) httproute.Route {
	return route.
		WithMethods("OPTIONS", "POST", "GET").
		WithPathPattern(AuthflowV2RouteSettingsMFAEnterOOBOTP)
}

type SettingsMFAEnterOOBOTPViewModel struct {
	AuthenticatorType              string
	Channels                       []model.AuthenticatorOOBChannel
	Channel                        model.AuthenticatorOOBChannel
	Target                         string
	MaskedClaimValue               string
	CodeLength                     int
	FailedAttemptRateLimitExceeded bool
	ResendCooldown                 int
}

func NewSettingsMFAEnterOOBOTPViewModel(oobConfig *config.AuthenticatorOOBConfig, tokenAuthenticator *accountmanagement.TokenAuthenticator, now time.Time, state *otp.State) SettingsMFAEnterOOBOTPViewModel {
	var target string
	var maskedClaimValue string
	var resendCooldown int
	var failedAttemptRateLimitExceeded bool
	var channels []model.AuthenticatorOOBChannel

	switch tokenAuthenticator.OOBOTPChannel {
	case model.AuthenticatorOOBChannelWhatsapp:
		fallthrough
	case model.AuthenticatorOOBChannelSMS:
		target = tokenAuthenticator.OOBOTPTarget
		maskedClaimValue = phone.Mask(tokenAuthenticator.OOBOTPTarget)
		switch oobConfig.SMS.PhoneOTPMode {
		case config.AuthenticatorPhoneOTPModeSMSOnly:
			channels = append(channels, model.AuthenticatorOOBChannelSMS)
		case config.AuthenticatorPhoneOTPModeWhatsappOnly:
			channels = append(channels, model.AuthenticatorOOBChannelWhatsapp)
		case config.AuthenticatorPhoneOTPModeWhatsappSMS:
			channels = append(channels, model.AuthenticatorOOBChannelWhatsapp)
			channels = append(channels, model.AuthenticatorOOBChannelSMS)
		}
	case model.AuthenticatorOOBChannelEmail:
		target = tokenAuthenticator.OOBOTPTarget
		maskedClaimValue = mail.MaskAddress(tokenAuthenticator.OOBOTPTarget)
	}

	cooldown := int(math.Ceil(state.CanResendAt.Sub(now).Seconds()))
	if cooldown > 0 {
		resendCooldown = cooldown
	}

	failedAttemptRateLimitExceeded = state.TooManyAttempts

	return SettingsMFAEnterOOBOTPViewModel{
		AuthenticatorType:              tokenAuthenticator.AuthenticatorType,
		Channels:                       channels,
		Channel:                        tokenAuthenticator.OOBOTPChannel,
		Target:                         target,
		MaskedClaimValue:               maskedClaimValue,
		CodeLength:                     6,
		FailedAttemptRateLimitExceeded: failedAttemptRateLimitExceeded,
		ResendCooldown:                 resendCooldown,
	}
}

type AuthflowV2SettingsMFAEnterOOBOTPHandler struct {
	ControllerFactory handlerwebapp.ControllerFactory
	BaseViewModel     *viewmodels.BaseViewModeler
	Renderer          handlerwebapp.Renderer
	FlashMessage      handlerwebapp.FlashMessage
	Clock             clock.Clock
	Config            *config.AppConfig
	OTPCode           handlerwebapp.OTPCodeService
	AccountManagement *accountmanagement.Service
}

func (h *AuthflowV2SettingsMFAEnterOOBOTPHandler) GetData(ctx context.Context, r *http.Request, w http.ResponseWriter, tokenAuthenticator *accountmanagement.TokenAuthenticator) (map[string]interface{}, error) {
	now := h.Clock.NowUTC()
	data := make(map[string]interface{})

	channel := tokenAuthenticator.OOBOTPChannel
	target := tokenAuthenticator.OOBOTPTarget

	state, err := h.OTPCode.InspectState(ctx, otp.KindVerification(h.Config, channel), target)
	if err != nil {
		return nil, err
	}

	baseViewModel := h.BaseViewModel.ViewModelForAuthFlow(r, w)
	viewmodels.Embed(data, baseViewModel)

	oobConfig := h.Config.Authenticator.OOB

	screenViewModel := NewSettingsMFAEnterOOBOTPViewModel(oobConfig, tokenAuthenticator, now, state)
	viewmodels.Embed(data, screenViewModel)

	return data, nil
}

func (h *AuthflowV2SettingsMFAEnterOOBOTPHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ctrl, err := h.ControllerFactory.New(r, w)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	defer ctrl.ServeWithoutDBTx(r.Context())

	ctrl.Get(func(ctx context.Context) error {
		tokenString := r.Form.Get("q_token")
		token, err := h.AccountManagement.GetToken(ctx, session.GetSession(ctx), tokenString)
		if err != nil {
			return err
		}

		data, err := h.GetData(ctx, r, w, token.Authenticator)
		if err != nil {
			return err
		}

		h.Renderer.RenderHTML(w, r, TemplateWebSettingsMFAEnterOOBOTPHTML, data)

		return nil
	})

	ctrl.PostAction("resend", func(ctx context.Context) error {
		err := AuthflowV2SettingsIdentityResendOOBOTPSchema.Validator().ValidateValue(handlerwebapp.FormToJSON(r.Form))
		if err != nil {
			return err
		}

		tokenString := r.Form.Get("q_token")
		err = h.AccountManagement.ResendOTPCode(ctx, session.GetSession(ctx), tokenString)
		if err != nil {
			return err
		}

		h.FlashMessage.Flash(w, string(webapp.FlashMessageTypeResendCodeSuccess))
		result := webapp.Result{}
		result.WriteResponse(w, r)
		return nil
	})

	ctrl.PostAction("submit", func(ctx context.Context) error {
		err := AuthflowV2SettingsMFAEnterOOBOTP.Validator().ValidateValue(handlerwebapp.FormToJSON(r.Form))
		if err != nil {
			return err
		}

		s := session.GetSession(ctx)

		tokenString := r.Form.Get("q_token")
		code := r.Form.Get("x_code")

		output, err := h.AccountManagement.ResumeAddOOBOTPAuthenticator(ctx, session.GetSession(ctx), tokenString, &accountmanagement.ResumeAddOOBOTPAuthenticatorInput{
			Code: code,
		})
		if err != nil {
			return err
		}

		var redirectURI *url.URL
		if output.RecoveryCodesCreated {
			redirectURI, err = url.Parse(AuthflowV2RouteSettingsMFAViewRecoveryCode)
			if err != nil {
				return err
			}
			q := redirectURI.Query()
			q.Set("q_token", output.Token)
			redirectURI.RawQuery = q.Encode()
		} else {
			_, err = h.AccountManagement.FinishAddOOBOTPAuthenticator(ctx, s, output.Token, &accountmanagement.FinishAddOOBOTPAuthenticatorInput{})
			if err != nil {
				return err
			}

			redirectURI, err = url.Parse(AuthflowV2RouteSettingsMFA)
			if err != nil {
				return err
			}
			q := redirectURI.Query()
			q.Set("q_token", output.Token)
			redirectURI.RawQuery = q.Encode()
		}

		result := webapp.Result{RedirectURI: redirectURI.String()}
		result.WriteResponse(w, r)

		return nil
	})
}
