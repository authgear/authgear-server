package webapp

import (
	"net/url"
	"strings"
)

type SessionStepKind string

const (
	SessionStepOAuthRedirect       SessionStepKind = "oauth-redirect"
	SessionStepPromoteUser         SessionStepKind = "promote-user"
	SessionStepAuthenticate        SessionStepKind = "authenticate"
	SessionStepCreateAuthenticator SessionStepKind = "create-authenticator"
	SessionStepEnterPassword       SessionStepKind = "enter-password"
	SessionStepCreatePassword      SessionStepKind = "create-password"
	SessionStepSendOOBOTPAuthn     SessionStepKind = "send-oob-otp-authn"
	SessionStepEnterOOBOTPAuthn    SessionStepKind = "enter-oob-otp-authn"
	SessionStepEnterOOBOTPSetup    SessionStepKind = "enter-oob-otp-setup"
	SessionStepSetupOOBOTPEmail    SessionStepKind = "setup-oob-otp-email"
	SessionStepSetupOOBOTPSMS      SessionStepKind = "setup-oob-otp-sms"
	SessionStepEnterTOTP           SessionStepKind = "enter-totp"
	SessionStepSetupTOTP           SessionStepKind = "setup-totp"
	SessionStepEnterRecoveryCode   SessionStepKind = "enter-recovery-code"
	SessionStepSetupRecoveryCode   SessionStepKind = "setup-recovery-code"
	SessionStepVerifyIdentity      SessionStepKind = "verify-identity"
	SessionStepUserDisabled        SessionStepKind = "user-disabled"
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
		return "/promote_user"
	case SessionStepEnterPassword:
		return "/enter_password"
	case SessionStepCreatePassword:
		return "/create_password"
	case SessionStepSendOOBOTPAuthn:
		return "/send_oob_otp"
	case SessionStepEnterOOBOTPAuthn, SessionStepEnterOOBOTPSetup:
		return "/enter_oob_otp"
	case SessionStepSetupOOBOTPEmail:
		return "/setup_oob_otp_email"
	case SessionStepSetupOOBOTPSMS:
		return "/setup_oob_otp_sms"
	case SessionStepEnterTOTP:
		return "/enter_totp"
	case SessionStepSetupTOTP:
		return "/setup_totp"
	case SessionStepEnterRecoveryCode:
		return "/enter_recovery_code"
	case SessionStepSetupRecoveryCode:
		return "/setup_recovery_code"
	case SessionStepVerifyIdentity:
		return "/verify_identity"
	case SessionStepUserDisabled:
		return "/user_disabled"
	case SessionStepOAuthRedirect,
		SessionStepAuthenticate,
		SessionStepCreateAuthenticator:
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
		case "/enter_totp", "/enter_password", "/enter_oob_otp", "/enter_recovery_code", "/send_oob_otp":
			return true
		default:
			return false
		}
	case SessionStepCreateAuthenticator:
		switch path {
		case "/setup_totp", "/setup_oob_otp", "/create_password":
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
