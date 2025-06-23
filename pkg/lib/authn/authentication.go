package authn

import (
	"fmt"

	"github.com/authgear/authgear-server/pkg/api/model"
)

type AuthenticationType string

const (
	AuthenticationTypeNone         AuthenticationType = "none"
	AuthenticationTypePassword     AuthenticationType = "password"
	AuthenticationTypePasskey      AuthenticationType = "passkey"
	AuthenticationTypeTOTP         AuthenticationType = "totp"
	AuthenticationTypeOOBOTPEmail  AuthenticationType = "oob_otp_email"
	AuthenticationTypeOOBOTPSMS    AuthenticationType = "oob_otp_sms"
	AuthenticationTypeRecoveryCode AuthenticationType = "recovery_code"
	AuthenticationTypeDeviceToken  AuthenticationType = "device_token"
)

type AuthenticationStage string

const (
	AuthenticationStagePrimary   AuthenticationStage = "primary"
	AuthenticationStageSecondary AuthenticationStage = "secondary"
)

func AuthenticationStageFromAuthenticationMethod(am model.AuthenticationFlowAuthentication) AuthenticationStage {
	switch am {
	case model.AuthenticationFlowAuthenticationPrimaryPassword:
		fallthrough
	case model.AuthenticationFlowAuthenticationPrimaryPasskey:
		fallthrough
	case model.AuthenticationFlowAuthenticationPrimaryOOBOTPEmail:
		fallthrough
	case model.AuthenticationFlowAuthenticationPrimaryOOBOTPSMS:
		return AuthenticationStagePrimary
	case model.AuthenticationFlowAuthenticationSecondaryPassword:
		fallthrough
	case model.AuthenticationFlowAuthenticationSecondaryTOTP:
		fallthrough
	case model.AuthenticationFlowAuthenticationSecondaryOOBOTPEmail:
		fallthrough
	case model.AuthenticationFlowAuthenticationSecondaryOOBOTPSMS:
		fallthrough
	case model.AuthenticationFlowAuthenticationRecoveryCode:
		// recovery code is considered as secondary
		fallthrough
	case model.AuthenticationFlowAuthenticationDeviceToken:
		// recovery code is considered as secondary
		return AuthenticationStageSecondary
	default:
		panic(fmt.Errorf("unknown authentication method: %v", am))
	}
}
