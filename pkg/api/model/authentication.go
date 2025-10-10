package model

import (
	"fmt"
)

type AuthenticationFlowAuthentication string

const (
	AuthenticationFlowAuthenticationPrimaryPassword      AuthenticationFlowAuthentication = "primary_password"
	AuthenticationFlowAuthenticationPrimaryPasskey       AuthenticationFlowAuthentication = "primary_passkey"
	AuthenticationFlowAuthenticationPrimaryOOBOTPEmail   AuthenticationFlowAuthentication = "primary_oob_otp_email"
	AuthenticationFlowAuthenticationPrimaryOOBOTPSMS     AuthenticationFlowAuthentication = "primary_oob_otp_sms"
	AuthenticationFlowAuthenticationSecondaryPassword    AuthenticationFlowAuthentication = "secondary_password"
	AuthenticationFlowAuthenticationSecondaryTOTP        AuthenticationFlowAuthentication = "secondary_totp"
	AuthenticationFlowAuthenticationSecondaryOOBOTPEmail AuthenticationFlowAuthentication = "secondary_oob_otp_email"
	AuthenticationFlowAuthenticationSecondaryOOBOTPSMS   AuthenticationFlowAuthentication = "secondary_oob_otp_sms"
	AuthenticationFlowAuthenticationRecoveryCode         AuthenticationFlowAuthentication = "recovery_code"
	AuthenticationFlowAuthenticationDeviceToken          AuthenticationFlowAuthentication = "device_token"
)

func (m AuthenticationFlowAuthentication) MaybeAuthenticatorKind() (AuthenticatorKind, bool) {
	switch m {
	case AuthenticationFlowAuthenticationPrimaryPassword:
		fallthrough
	case AuthenticationFlowAuthenticationPrimaryPasskey:
		fallthrough
	case AuthenticationFlowAuthenticationPrimaryOOBOTPEmail:
		fallthrough
	case AuthenticationFlowAuthenticationPrimaryOOBOTPSMS:
		return AuthenticatorKindPrimary, true
	case AuthenticationFlowAuthenticationSecondaryPassword:
		fallthrough
	case AuthenticationFlowAuthenticationSecondaryTOTP:
		fallthrough
	case AuthenticationFlowAuthenticationSecondaryOOBOTPEmail:
		fallthrough
	case AuthenticationFlowAuthenticationSecondaryOOBOTPSMS:
		return AuthenticatorKindSecondary, true
	case AuthenticationFlowAuthenticationRecoveryCode:
		fallthrough
	case AuthenticationFlowAuthenticationDeviceToken:
		return "", false
	default:
		panic(fmt.Errorf("unknown authentication: %v", m))
	}
}

func (m AuthenticationFlowAuthentication) AuthenticatorKind() AuthenticatorKind {
	kind, ok := m.MaybeAuthenticatorKind()
	if ok {
		return kind
	}
	panic(fmt.Errorf("%v is not an authenticator", m))
}

func (a AuthenticationFlowAuthentication) AMR() []string {
	switch a {
	case AuthenticationFlowAuthenticationPrimaryPassword:
		return []string{AMRPWD, AMRXPrimaryPassword}
	case AuthenticationFlowAuthenticationPrimaryOOBOTPEmail:
		return []string{AMROTP, AMRXPrimaryOOBOTPEmail}
	case AuthenticationFlowAuthenticationPrimaryOOBOTPSMS:
		return []string{AMROTP, AMRSMS, AMRXPrimaryOOBOTPSMS}
	case AuthenticationFlowAuthenticationPrimaryPasskey:
		return []string{AMRXPasskey, AMRXPrimaryPasskey}
	case AuthenticationFlowAuthenticationSecondaryPassword:
		return []string{AMRPWD, AMRXSecondaryPassword}
	case AuthenticationFlowAuthenticationSecondaryTOTP:
		return []string{AMROTP, AMRXSecondaryTOTP}
	case AuthenticationFlowAuthenticationSecondaryOOBOTPEmail:
		return []string{AMROTP, AMRXSecondaryOOBOTPEmail}
	case AuthenticationFlowAuthenticationSecondaryOOBOTPSMS:
		return []string{AMROTP, AMRSMS, AMRXSecondaryOOBOTPSMS}
	case AuthenticationFlowAuthenticationRecoveryCode:
		return []string{AMRXRecoveryCode}
	case AuthenticationFlowAuthenticationDeviceToken:
		return []string{AMRXDeviceToken}
	default:
		panic(fmt.Errorf("unknown authentication: %v", a))
	}
}

type Authentication struct {
	Authentication AuthenticationFlowAuthentication `json:"authentication"`
	Authenticator  *Authenticator                   `json:"authenticator"`
}
