package authflowclient

import (
	"encoding/json"
)

type FlowType string

const (
	FlowTypeSignup          FlowType = "signup"
	FlowTypePromote         FlowType = "promote"
	FlowTypeLogin           FlowType = "login"
	FlowTypeSignupLogin     FlowType = "signup_login"
	FlowTypeReauth          FlowType = "reauth"
	FlowTypeAccountRecovery FlowType = "account_recovery"
)

type FlowReference struct {
	Type FlowType `json:"type"`
	Name string   `json:"name"`
}

type FlowActionType string

const (
	FlowActionTypeFinished                  FlowActionType = "finished"
	FlowActionTypeIdentify                  FlowActionType = "identify"
	FlowActionTypeAuthenticate              FlowActionType = "authenticate"
	FlowActionTypeCreateAuthenticator       FlowActionType = "create_authenticator"
	FlowActionTypeVerify                    FlowActionType = "verify"
	FlowActionTypeFillInUserProfile         FlowActionType = "fill_in_user_profile"
	FlowActionTypeViewRecoveryCode          FlowActionType = "view_recovery_code"
	FlowActionTypePromptCreatePasskey       FlowActionType = "prompt_create_passkey"
	FlowActionTypeTerminateOtherSessions    FlowActionType = "terminate_other_sessions"
	FlowActionTypeCheckAccountStatus        FlowActionType = "check_account_status"
	FlowActionTypeChangePassword            FlowActionType = "change_password"
	FlowActionTypeSelectDestination         FlowActionType = "select_destination"
	FlowActionTypeVerifyAccountRecoveryCode FlowActionType = "verify_account_recovery_code"
	FlowActionTypeResetPassword             FlowActionType = "reset_password"
)

type Identification string

const (
	IdentificationEmail    Identification = "email"
	IdentificationPhone    Identification = "phone"
	IdentificationUsername Identification = "username"
	IdentificationOAuth    Identification = "oauth"
	IdentificationPasskey  Identification = "passkey"
	IdentificationIDToken  Identification = "id_token"
)

type AccountRecoveryIdentification string

const (
	AccountRecoveryIdentificationEmail AccountRecoveryIdentification = "email"
	AccountRecoveryIdentificationPhone AccountRecoveryIdentification = "phone"
)

type Authentication string

const (
	AuthenticationPrimaryPassword      Authentication = "primary_password"
	AuthenticationPrimaryPasskey       Authentication = "primary_passkey"
	AuthenticationPrimaryOOBOTPEmail   Authentication = "primary_oob_otp_email"
	AuthenticationPrimaryOOBOTPSMS     Authentication = "primary_oob_otp_sms"
	AuthenticationSecondaryPassword    Authentication = "secondary_password"
	AuthenticationSecondaryTOTP        Authentication = "secondary_totp"
	AuthenticationSecondaryOOBOTPEmail Authentication = "secondary_oob_otp_email"
	AuthenticationSecondaryOOBOTPSMS   Authentication = "secondary_oob_otp_sms"
	AuthenticationRecoveryCode         Authentication = "recovery_code"
	AuthenticationDeviceToken          Authentication = "device_token"
)

type FlowAction struct {
	Type           FlowActionType  `json:"type"`
	Identification Identification  `json:"identification,omitempty"`
	Authentication Authentication  `json:"authentication,omitempty"`
	Data           json.RawMessage `json:"data,omitempty"`
}

type FlowResponse struct {
	StateToken string      `json:"state_token"`
	Type       FlowType    `json:"type,omitempty"`
	Name       string      `json:"name,omitempty"`
	Action     *FlowAction `json:"action,omitempty"`
}
