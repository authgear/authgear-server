package webapp

import (
	"github.com/authgear/authgear-server/pkg/api/model"
	"github.com/authgear/authgear-server/pkg/lib/authn"
	"github.com/authgear/authgear-server/pkg/lib/interaction/nodes"
)

type InputUseOAuth struct {
	ProviderAlias    string
	ErrorRedirectURI string
	Prompt           []string
}

var _ nodes.InputUseIdentityOAuthProvider = &InputUseOAuth{}

func (i *InputUseOAuth) GetProviderAlias() string    { return i.ProviderAlias }
func (i *InputUseOAuth) GetErrorRedirectURI() string { return i.ErrorRedirectURI }
func (i *InputUseOAuth) GetPrompt() []string         { return i.Prompt }

type InputUseLoginID struct {
	LoginIDKey string
	LoginID    string
}

var _ nodes.InputUseIdentityLoginID = &InputUseLoginID{}

func (i *InputUseLoginID) GetLoginIDKey() string { return i.LoginIDKey }
func (i *InputUseLoginID) GetLoginID() string    { return i.LoginID }

type InputNewLoginID struct {
	LoginIDType  string
	LoginIDKey   string
	LoginIDValue string
}

var _ nodes.InputUseIdentityLoginID = &InputNewLoginID{}

func (i *InputNewLoginID) GetLoginIDKey() string { return i.LoginIDKey }
func (i *InputNewLoginID) GetLoginID() string    { return i.LoginIDValue }

type InputCreateAuthenticator struct{}

func (i *InputCreateAuthenticator) RequestedByUser() bool { return true }

type InputRemoveAuthenticator struct {
	Type model.AuthenticatorType
	ID   string
}

var _ nodes.InputRemoveAuthenticator = &InputRemoveAuthenticator{}

func (i *InputRemoveAuthenticator) GetAuthenticatorType() model.AuthenticatorType { return i.Type }
func (i *InputRemoveAuthenticator) GetAuthenticatorID() string                    { return i.ID }

type InputRemoveIdentity struct {
	Type model.IdentityType
	ID   string
}

var _ nodes.InputRemoveIdentity = &InputRemoveIdentity{}

func (i *InputRemoveIdentity) GetIdentityType() model.IdentityType { return i.Type }
func (i *InputRemoveIdentity) GetIdentityID() string               { return i.ID }

type InputTriggerOOB struct {
	AuthenticatorType  string
	AuthenticatorIndex int
}

var _ nodes.InputAuthenticationOOBTrigger = &InputTriggerOOB{}

func (i *InputTriggerOOB) GetOOBAuthenticatorType() string { return i.AuthenticatorType }
func (i *InputTriggerOOB) GetOOBAuthenticatorIndex() int   { return i.AuthenticatorIndex }

type InputTriggerWhatsApp struct {
	AuthenticatorIndex int
}

var _ nodes.InputAuthenticationWhatsappTrigger = &InputTriggerWhatsApp{}

func (i *InputTriggerWhatsApp) GetWhatsappAuthenticatorIndex() int { return i.AuthenticatorIndex }

type InputTriggerLoginLink struct {
	AuthenticatorIndex int
}

var _ nodes.InputAuthenticationLoginLinkTrigger = &InputTriggerLoginLink{}

func (i *InputTriggerLoginLink) GetLoginLinkAuthenticatorIndex() int { return i.AuthenticatorIndex }

type InputSelectTOTP struct{}

var _ nodes.InputCreateAuthenticatorTOTPSetup = &InputSelectTOTP{}

func (i *InputSelectTOTP) SetupTOTP() {}

type InputSelectWhatsappOTP struct{}

func (i *InputSelectWhatsappOTP) SetupPrimaryAuthenticatorWhatsappOTP() {}

var _ nodes.InputCreateAuthenticatorWhatsappOTPSetupSelect = &InputSelectWhatsappOTP{}

type InputSelectOOB struct{}

func (i *InputSelectOOB) SetupPrimaryAuthenticatorOOB() {}

var _ nodes.InputCreateAuthenticatorOOBSetupSelect = &InputSelectOOB{}

type InputSelectLoginLink struct{}

func (i *InputSelectLoginLink) SetupPrimaryAuthenticatorLoginLinkOTP() {}

var _ nodes.InputCreateAuthenticatorLoginLinkOTPSetupSelect = &InputSelectLoginLink{}

type InputSetupPassword struct {
	Stage    string
	Password string
}

var _ nodes.InputCreateAuthenticatorPassword = &InputSetupPassword{}
var _ nodes.InputAuthenticationStage = &InputSetupPassword{}

func (i *InputSetupPassword) GetPassword() string { return i.Password }
func (i *InputSetupPassword) GetAuthenticationStage() authn.AuthenticationStage {
	return authn.AuthenticationStage(i.Stage)
}

type InputPasskeyAttestationResponse struct {
	Stage               string
	AttestationResponse []byte
}

var _ nodes.InputCreateAuthenticatorPasskey = &InputPasskeyAttestationResponse{}
var _ nodes.InputAuthenticationStage = &InputPasskeyAttestationResponse{}

type InputPromptCreatePasskeyAttestationResponse struct {
	Skipped             bool
	AttestationResponse []byte
}

var _ nodes.InputPromptCreatePasskey = &InputPromptCreatePasskeyAttestationResponse{}

func (i *InputPromptCreatePasskeyAttestationResponse) IsSkipped() bool {
	return i.Skipped
}

func (i *InputPromptCreatePasskeyAttestationResponse) GetAttestationResponse() []byte {
	return i.AttestationResponse
}

func (i *InputPasskeyAttestationResponse) GetAttestationResponse() []byte {
	return i.AttestationResponse
}

func (i *InputPasskeyAttestationResponse) GetAuthenticationStage() authn.AuthenticationStage {
	return authn.AuthenticationStage(i.Stage)
}

type InputPasskeyAssertionResponse struct {
	Stage             string
	AssertionResponse []byte
}

var _ nodes.InputAuthenticationPasskey = &InputPasskeyAssertionResponse{}
var _ nodes.InputAuthenticationStage = &InputPasskeyAssertionResponse{}

func (i *InputPasskeyAssertionResponse) GetAssertionResponse() []byte {
	return i.AssertionResponse
}

func (i *InputPasskeyAssertionResponse) GetAuthenticationStage() authn.AuthenticationStage {
	return authn.AuthenticationStage(i.Stage)
}

type InputResendCode struct{}

func (i *InputResendCode) DoResend() {}

type InputAuthOOB struct {
	Code        string
	DeviceToken bool
}

var _ nodes.InputAuthenticationOOB = &InputAuthOOB{}
var _ nodes.InputCreateDeviceToken = &InputAuthOOB{}

func (i *InputAuthOOB) GetOOBOTP() string       { return i.Code }
func (i *InputAuthOOB) CreateDeviceToken() bool { return i.DeviceToken }

type InputAuthTOTP struct {
	Code        string
	DeviceToken bool
}

var _ nodes.InputAuthenticationTOTP = &InputAuthTOTP{}
var _ nodes.InputCreateDeviceToken = &InputAuthTOTP{}

func (i *InputAuthTOTP) GetTOTP() string         { return i.Code }
func (i *InputAuthTOTP) CreateDeviceToken() bool { return i.DeviceToken }

type InputAuthPassword struct {
	Stage       string
	Password    string
	DeviceToken bool
}

var _ nodes.InputAuthenticationPassword = &InputAuthPassword{}
var _ nodes.InputCreateDeviceToken = &InputAuthPassword{}
var _ nodes.InputAuthenticationStage = &InputAuthPassword{}

func (i *InputAuthPassword) GetPassword() string     { return i.Password }
func (i *InputAuthPassword) CreateDeviceToken() bool { return i.DeviceToken }
func (i *InputAuthPassword) GetAuthenticationStage() authn.AuthenticationStage {
	return authn.AuthenticationStage(i.Stage)
}

type InputAuthRecoveryCode struct {
	Code        string
	DeviceToken bool
}

var _ nodes.InputConsumeRecoveryCode = &InputAuthRecoveryCode{}
var _ nodes.InputCreateDeviceToken = &InputAuthRecoveryCode{}

func (i *InputAuthRecoveryCode) GetRecoveryCode() string { return i.Code }
func (i *InputAuthRecoveryCode) CreateDeviceToken() bool { return i.DeviceToken }

type InputSetupOOB struct {
	InputType string
	Target    string
}

var _ nodes.InputCreateAuthenticatorOOBSetup = &InputSetupOOB{}

func (i *InputSetupOOB) GetOOBChannel() model.AuthenticatorOOBChannel {
	switch i.InputType {
	case "email":
		return model.AuthenticatorOOBChannelEmail
	case "phone":
		return model.AuthenticatorOOBChannelSMS
	default:
		panic("webapp: unknown input type: " + i.InputType)
	}
}
func (i *InputSetupOOB) GetOOBTarget() string { return i.Target }

type InputSetupRecoveryCode struct{}

var _ nodes.InputGenerateRecoveryCodeEnd = &InputSetupRecoveryCode{}

func (i *InputSetupRecoveryCode) ViewedRecoveryCodes() {}

type InputSetupTOTP struct {
	Code        string
	DisplayName string
}

var _ nodes.InputCreateAuthenticatorTOTP = &InputSetupTOTP{}

func (i *InputSetupTOTP) GetTOTP() string            { return i.Code }
func (i *InputSetupTOTP) GetTOTPDisplayName() string { return i.DisplayName }

type InputOAuthCallback struct {
	ProviderAlias string
	Query         string
}

var _ nodes.InputUseIdentityOAuthUserInfo = &InputOAuthCallback{}

func (i *InputOAuthCallback) GetProviderAlias() string { return i.ProviderAlias }
func (i *InputOAuthCallback) GetQuery() string         { return i.Query }

type InputVerificationCode struct {
	Code string
}

var _ nodes.InputVerifyIdentityCheckCode = &InputVerificationCode{}

func (i *InputVerificationCode) GetVerificationCode() string { return i.Code }

type InputChangePassword struct {
	AuthenticationStage authn.AuthenticationStage
	OldPassword         string
	NewPassword         string
}

var _ nodes.InputChangePassword = &InputChangePassword{}

func (i *InputChangePassword) GetAuthenticationStage() authn.AuthenticationStage {
	return i.AuthenticationStage
}
func (i *InputChangePassword) GetOldPassword() string { return i.OldPassword }
func (i *InputChangePassword) GetNewPassword() string { return i.NewPassword }

type InputResetPassword struct {
	Code     string
	Password string
}

var _ nodes.InputResetPasswordByCode = &InputResetPassword{}

func (i *InputResetPassword) GetCode() string        { return i.Code }
func (i *InputResetPassword) GetNewPassword() string { return i.Password }

type InputVerifyWhatsappOTP struct {
	DeviceToken bool
	WhatsappOTP string
}

func (i *InputVerifyWhatsappOTP) GetWhatsappOTP() string  { return i.WhatsappOTP }
func (i *InputVerifyWhatsappOTP) CreateDeviceToken() bool { return i.DeviceToken }

var _ nodes.InputCreateAuthenticatorWhatsappOTP = &InputVerifyWhatsappOTP{}
var _ nodes.InputAuthenticationWhatsapp = &InputVerifyWhatsappOTP{}
var _ nodes.InputVerifyIdentityViaWhatsappCheckCode = &InputVerifyWhatsappOTP{}
var _ nodes.InputCreateDeviceToken = &InputVerifyWhatsappOTP{}

type InputVerifyLoginLinkOTP struct {
	DeviceToken bool
}

func (i *InputVerifyLoginLinkOTP) VerifyLoginLink()        {}
func (i *InputVerifyLoginLinkOTP) CreateDeviceToken() bool { return i.DeviceToken }

var _ nodes.InputCreateAuthenticatorLoginLinkOTP = &InputVerifyLoginLinkOTP{}
var _ nodes.InputAuthenticationLoginLink = &InputVerifyLoginLinkOTP{}
var _ nodes.InputCreateDeviceToken = &InputVerifyLoginLinkOTP{}

type InputSetupWhatsappOTP struct {
	Phone string
}

func (i *InputSetupWhatsappOTP) GetWhatsappPhone() string { return i.Phone }

var _ nodes.InputCreateAuthenticatorWhatsappOTPSetup = &InputSetupWhatsappOTP{}

type InputSetupLoginLinkOTP struct {
	InputType string
	Target    string
}

func (i *InputSetupLoginLinkOTP) GetLoginLinkOTPTarget() string { return i.Target }

var _ nodes.InputCreateAuthenticatorLoginLinkOTPSetup = &InputSetupLoginLinkOTP{}

type InputSelectVerifyIdentityViaOOBOTP struct{}

func (i *InputSelectVerifyIdentityViaOOBOTP) SelectVerifyIdentityViaOOBOTP() {}

var _ nodes.InputVerifyIdentity = &InputSelectVerifyIdentityViaOOBOTP{}

type InputSelectVerifyIdentityViaWhatsapp struct{}

func (i *InputSelectVerifyIdentityViaWhatsapp) SelectVerifyIdentityViaWhatsapp() {}

var _ nodes.InputVerifyIdentityViaWhatsapp = &InputSelectVerifyIdentityViaWhatsapp{}

type InputConfirmWeb3AccountRequest struct {
	Message   string
	Signature string
}

func (i *InputConfirmWeb3AccountRequest) GetMessage() string   { return i.Message }
func (i *InputConfirmWeb3AccountRequest) GetSignature() string { return i.Signature }

var _ nodes.InputUseIdentitySIWE = &InputConfirmWeb3AccountRequest{}

type InputConfirmTerminateOtherSessions struct {
	IsConfirm bool
}

func (i *InputConfirmTerminateOtherSessions) GetIsConfirmed() bool { return i.IsConfirm }

var _ nodes.InputConfirmTerminateOtherSessionsEnd = &InputConfirmTerminateOtherSessions{}
