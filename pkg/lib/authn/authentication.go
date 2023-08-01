package authn

import (
	"fmt"

	"github.com/authgear/authgear-server/pkg/lib/config"
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

func AuthenticationStageFromAuthenticationMethod(am config.WorkflowAuthenticationMethod) AuthenticationStage {
	switch am {
	case config.WorkflowAuthenticationMethodPrimaryPassword:
		fallthrough
	case config.WorkflowAuthenticationMethodPrimaryPasskey:
		fallthrough
	case config.WorkflowAuthenticationMethodPrimaryOOBOTPEmail:
		fallthrough
	case config.WorkflowAuthenticationMethodPrimaryOOBOTPSMS:
		return AuthenticationStagePrimary
	case config.WorkflowAuthenticationMethodSecondaryPassword:
		fallthrough
	case config.WorkflowAuthenticationMethodSecondaryTOTP:
		fallthrough
	case config.WorkflowAuthenticationMethodSecondaryOOBOTPEmail:
		fallthrough
	case config.WorkflowAuthenticationMethodSecondaryOOBOTPSMS:
		return AuthenticationStageSecondary
	default:
		panic(fmt.Errorf("unknown authentication method: %v", am))
	}
}
