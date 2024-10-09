package authn

import (
	"fmt"

	"github.com/authgear/authgear-server/pkg/lib/config"
)

type AuthenticationType string

const (
	AuthenticationTypeNone            AuthenticationType = "none"
	AuthenticationTypePassword        AuthenticationType = "password"
	AuthenticationTypePasskey         AuthenticationType = "passkey"
	AuthenticationTypeTOTP            AuthenticationType = "totp"
	AuthenticationTypeOOBOTPEmail     AuthenticationType = "oob_otp_email"
	AuthenticationTypeOOBOTPSMS       AuthenticationType = "oob_otp_sms"
	AuthenticationTypeFaceRecognition AuthenticationType = "face_recognition"
	AuthenticationTypeRecoveryCode    AuthenticationType = "recovery_code"
	AuthenticationTypeDeviceToken     AuthenticationType = "device_token"
)

type AuthenticationStage string

const (
	AuthenticationStagePrimary   AuthenticationStage = "primary"
	AuthenticationStageSecondary AuthenticationStage = "secondary"
)

func AuthenticationStageFromAuthenticationMethod(am config.AuthenticationFlowAuthentication) AuthenticationStage {
	switch am {
	case config.AuthenticationFlowAuthenticationPrimaryPassword:
		fallthrough
	case config.AuthenticationFlowAuthenticationPrimaryPasskey:
		fallthrough
	case config.AuthenticationFlowAuthenticationPrimaryOOBOTPEmail:
		fallthrough
	case config.AuthenticationFlowAuthenticationPrimaryOOBOTPSMS:
		return AuthenticationStagePrimary
	case config.AuthenticationFlowAuthenticationSecondaryPassword:
		fallthrough
	case config.AuthenticationFlowAuthenticationSecondaryTOTP:
		fallthrough
	case config.AuthenticationFlowAuthenticationSecondaryFaceRecognition:
		fallthrough
	case config.AuthenticationFlowAuthenticationSecondaryOOBOTPEmail:
		fallthrough
	case config.AuthenticationFlowAuthenticationSecondaryOOBOTPSMS:
		fallthrough
	case config.AuthenticationFlowAuthenticationRecoveryCode:
		// recovery code is considered as secondary
		fallthrough
	case config.AuthenticationFlowAuthenticationDeviceToken:
		// recovery code is considered as secondary
		return AuthenticationStageSecondary
	default:
		panic(fmt.Errorf("unknown authentication method: %v", am))
	}
}
