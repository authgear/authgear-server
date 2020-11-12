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
	SessionStepEnterOOBOTPAuthn    SessionStepKind = "enter-oob-otp-authn"
	SessionStepEnterOOBOTPSetup    SessionStepKind = "enter-oob-otp-setup"
	SessionStepSetupOOBOTP         SessionStepKind = "setup-oob-otp"
	SessionStepEnterTOTP           SessionStepKind = "enter-totp"
	SessionStepSetupTOTP           SessionStepKind = "setup-totp"
	SessionStepEnterRecoveryCode   SessionStepKind = "enter-recovery-code"
	SessionStepSetupRecoveryCode   SessionStepKind = "setup-recovery-code"
	SessionStepVerifyIdentity      SessionStepKind = "verify-identity"
	SessionStepUserBlocked         SessionStepKind = "user-blocked"
)

func (k SessionStepKind) Path() string {
	switch k {
	case SessionStepPromoteUser:
		return "/promote_user"
	case SessionStepEnterPassword:
		return "/enter_password"
	case SessionStepCreatePassword:
		return "/create_password"
	case SessionStepEnterOOBOTPAuthn, SessionStepEnterOOBOTPSetup:
		return "/enter_oob_otp"
	case SessionStepSetupOOBOTP:
		return "/setup_oob_otp"
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
	case SessionStepUserBlocked:
		return "/user_blocked"
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
		return strings.HasPrefix(path, "/sso/oauth2/callback/")
	case SessionStepAuthenticate:
		switch path {
		case "/enter_totp", "/enter_password", "/enter_oob_otp", "/enter_recovery_code":
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
	Kind    SessionStepKind `json:"kind"`
	GraphID string          `json:"graph_id"`
}

func (s SessionStep) URL() *url.URL {
	query := url.Values{}
	query.Set("x_step", s.GraphID)
	u := url.URL{Path: s.Kind.Path(), RawQuery: query.Encode()}
	return &u
}
