package webapp

import (
	"net/url"
	"strings"
)

type SessionStepKind string

const (
	SessionStepOAuthRedirect             SessionStepKind = "oauth-redirect"
	SessionStepPromoteUser               SessionStepKind = "promote-user"
	SessionStepAuthenticate              SessionStepKind = "authenticate"
	SessionStepCreateAuthenticator       SessionStepKind = "create-authenticator"
	SessionStepEnterPassword             SessionStepKind = "enter-password"
	SessionStepUsePasskey                SessionStepKind = "use-passkey"
	SessionStepCreatePassword            SessionStepKind = "create-password"
	SessionStepCreatePasskey             SessionStepKind = "create-passkey"
	SessionStepPromptCreatePasskey       SessionStepKind = "prompt-create-passkey"
	SessionStepChangePrimaryPassword     SessionStepKind = "change-primary-password"
	SessionStepChangeSecondaryPassword   SessionStepKind = "change-secondary-password"
	SessionStepEnterOOBOTPAuthnEmail     SessionStepKind = "enter-oob-otp-authn-email"
	SessionStepEnterOOBOTPAuthnSMS       SessionStepKind = "enter-oob-otp-authn-sms"
	SessionStepEnterOOBOTPSetupEmail     SessionStepKind = "enter-oob-otp-setup-email"
	SessionStepEnterOOBOTPSetupSMS       SessionStepKind = "enter-oob-otp-setup-sms"
	SessionStepSetupOOBOTPEmail          SessionStepKind = "setup-oob-otp-email"
	SessionStepSetupOOBOTPSMS            SessionStepKind = "setup-oob-otp-sms"
	SessionStepSetupWhatsappOTP          SessionStepKind = "setup-whatsapp-otp"
	SessionStepVerifyWhatsappOTPAuthn    SessionStepKind = "verify-whatsapp-otp-authn"
	SessionStepVerifyWhatsappOTPSetup    SessionStepKind = "verify-whatsapp-otp-setup"
	SessionStepEnterTOTP                 SessionStepKind = "enter-totp"
	SessionStepSetupTOTP                 SessionStepKind = "setup-totp"
	SessionStepEnterRecoveryCode         SessionStepKind = "enter-recovery-code"
	SessionStepSetupRecoveryCode         SessionStepKind = "setup-recovery-code"
	SessionStepVerifyIdentityBegin       SessionStepKind = "verify-identity-begin"
	SessionStepVerifyIdentityViaOOBOTP   SessionStepKind = "verify-identity"
	SessionStepVerifyIdentityViaWhatsapp SessionStepKind = "verify-identity-via-whatsapp"
	SessionStepAccountStatus             SessionStepKind = "account-status"
)

func NewSessionStep(kind SessionStepKind, graphID string) SessionStep {
	return SessionStep{
		Kind:     kind,
		GraphID:  graphID,
		FormData: make(map[string]interface{}),
	}
}

func (k SessionStepKind) Path() string {
	switch k {
	case SessionStepPromoteUser:
		return "/flows/promote_user"
	case SessionStepEnterPassword:
		return "/flows/enter_password"
	case SessionStepUsePasskey:
		return "/flows/use_passkey"
	case SessionStepCreatePassword:
		return "/flows/create_password"
	case SessionStepCreatePasskey:
		return "/flows/create_passkey"
	case SessionStepPromptCreatePasskey:
		return "/flows/prompt_create_passkey"
	case SessionStepChangePrimaryPassword:
		return "/flows/change_password"
	case SessionStepChangeSecondaryPassword:
		return "/flows/change_secondary_password"
	case SessionStepEnterOOBOTPAuthnEmail,
		SessionStepEnterOOBOTPAuthnSMS,
		SessionStepEnterOOBOTPSetupEmail,
		SessionStepEnterOOBOTPSetupSMS:
		return "/flows/enter_oob_otp"
	case SessionStepSetupOOBOTPEmail:
		return "/flows/setup_oob_otp_email"
	case SessionStepSetupOOBOTPSMS:
		return "/flows/setup_oob_otp_sms"
	case SessionStepSetupWhatsappOTP:
		return "/flows/setup_whatsapp_otp"
	case SessionStepVerifyWhatsappOTPSetup,
		SessionStepVerifyWhatsappOTPAuthn,
		SessionStepVerifyIdentityViaWhatsapp:
		return "/flows/whatsapp_otp"
	case SessionStepEnterTOTP:
		return "/flows/enter_totp"
	case SessionStepSetupTOTP:
		return "/flows/setup_totp"
	case SessionStepEnterRecoveryCode:
		return "/flows/enter_recovery_code"
	case SessionStepSetupRecoveryCode:
		return "/setup_recovery_code"
	case SessionStepVerifyIdentityViaOOBOTP:
		return "/flows/verify_identity"
	case SessionStepAccountStatus:
		return "/flows/account_status"
	case SessionStepOAuthRedirect,
		SessionStepAuthenticate,
		SessionStepCreateAuthenticator,
		SessionStepVerifyIdentityBegin:
		// No path for step.
		return ""
	default:
		panic("webapp: unknown step " + string(k))
	}
}

func (k SessionStepKind) MatchPath(path string) bool {
	switch k {
	case SessionStepOAuthRedirect:
		// In Wechat authorize flow, instead of redirect user to provider authorization page
		// redirect user to page that display qr code
		// https://developers.weixin.qq.com/doc/offiaccount/OA_Web_Apps/Wechat_webpage_authorization.html
		return strings.HasPrefix(path, "/sso/wechat/auth/") ||
			strings.HasPrefix(path, "/sso/oauth2/callback/")
	case SessionStepAuthenticate:
		switch path {
		case "/flows/enter_totp", "/flows/enter_password", "/flows/enter_oob_otp", "/flows/enter_recovery_code":
			return true
		default:
			return false
		}
	case SessionStepCreateAuthenticator:
		switch path {
		case "/flows/setup_totp", "/flows/setup_oob_otp_email", "/flows/setup_oob_otp_sms", "/flows/create_password", "/flows/setup_whatsapp_otp":
			return true
		default:
			return false
		}
	default:
		return k.Path() == path
	}
}

type SessionStep struct {
	// Kind is the kind of the step.
	Kind SessionStepKind `json:"kind"`
	// GraphID is the graph ID of the step.
	GraphID string `json:"graph_id"`
	// FormData is the place to store shared form data across different user agents.
	// The only use case currently is verification email being opened in another user agent.
	// In that case, the form submitted by the other user agent will update FormData.
	// The original user agent will then read from it to fill in its form.
	FormData map[string]interface{} `json:"form_data"`
}

func (s SessionStep) URL() *url.URL {
	query := url.Values{}
	query.Set("x_step", s.GraphID)
	u := url.URL{Path: s.Kind.Path(), RawQuery: query.Encode()}
	return &u
}
