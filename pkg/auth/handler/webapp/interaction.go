package webapp

import (
	"github.com/authgear/authgear-server/pkg/lib/authn"
	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/lib/interaction"
	"github.com/authgear/authgear-server/pkg/lib/interaction/nodes"
)

type InputUseOAuth struct {
	ProviderAlias    string
	ErrorRedirectURI string
	Prompt           []string
}

var _ interaction.Input = &InputUseOAuth{}
var _ nodes.InputUseIdentityOAuthProvider = &InputUseOAuth{}

func (*InputUseOAuth) IsInteractive() bool           { return true }
func (i *InputUseOAuth) GetProviderAlias() string    { return i.ProviderAlias }
func (i *InputUseOAuth) GetErrorRedirectURI() string { return i.ErrorRedirectURI }
func (i *InputUseOAuth) GetPrompt() []string         { return i.Prompt }

type InputUseLoginID struct {
	LoginIDKey string
	LoginID    string
}

var _ interaction.Input = &InputUseLoginID{}
var _ nodes.InputUseIdentityLoginID = &InputUseLoginID{}

func (*InputUseLoginID) IsInteractive() bool     { return true }
func (i *InputUseLoginID) GetLoginIDKey() string { return i.LoginIDKey }
func (i *InputUseLoginID) GetLoginID() string    { return i.LoginID }

type InputNewLoginID struct {
	LoginIDType  string
	LoginIDKey   string
	LoginIDValue string
}

var _ interaction.Input = &InputNewLoginID{}
var _ nodes.InputUseIdentityLoginID = &InputNewLoginID{}
var _ nodes.InputCreateAuthenticatorOOBSetup = &InputNewLoginID{}

func (*InputNewLoginID) IsInteractive() bool     { return true }
func (i *InputNewLoginID) GetLoginIDKey() string { return i.LoginIDKey }
func (i *InputNewLoginID) GetLoginID() string    { return i.LoginIDValue }
func (i *InputNewLoginID) GetOOBChannel() authn.AuthenticatorOOBChannel {
	switch i.LoginIDType {
	case string(config.LoginIDKeyTypeEmail):
		return authn.AuthenticatorOOBChannelEmail
	case string(config.LoginIDKeyTypePhone):
		return authn.AuthenticatorOOBChannelSMS
	default:
		return ""
	}
}
func (i *InputNewLoginID) GetOOBTarget() string { return i.LoginIDValue }

type InputCreateAuthenticator struct{}

var _ interaction.Input = &InputCreateAuthenticator{}

func (*InputCreateAuthenticator) IsInteractive() bool     { return true }
func (i *InputCreateAuthenticator) RequestedByUser() bool { return true }

type InputRemoveAuthenticator struct {
	Type authn.AuthenticatorType
	ID   string
}

var _ interaction.Input = &InputRemoveAuthenticator{}
var _ nodes.InputRemoveAuthenticator = &InputRemoveAuthenticator{}

func (*InputRemoveAuthenticator) IsInteractive() bool                             { return true }
func (i *InputRemoveAuthenticator) GetAuthenticatorType() authn.AuthenticatorType { return i.Type }
func (i *InputRemoveAuthenticator) GetAuthenticatorID() string                    { return i.ID }

type InputRemoveIdentity struct {
	Type authn.IdentityType
	ID   string
}

var _ interaction.Input = &InputRemoveIdentity{}
var _ nodes.InputRemoveIdentity = &InputRemoveIdentity{}

func (*InputRemoveIdentity) IsInteractive() bool                   { return true }
func (i *InputRemoveIdentity) GetIdentityType() authn.IdentityType { return i.Type }
func (i *InputRemoveIdentity) GetIdentityID() string               { return i.ID }

type InputTriggerOOB struct {
	AuthenticatorType  string
	AuthenticatorIndex int
}

var _ interaction.Input = &InputTriggerOOB{}
var _ nodes.InputAuthenticationOOBTrigger = &InputTriggerOOB{}

func (*InputTriggerOOB) IsInteractive() bool               { return false }
func (i *InputTriggerOOB) GetOOBAuthenticatorType() string { return i.AuthenticatorType }
func (i *InputTriggerOOB) GetOOBAuthenticatorIndex() int   { return i.AuthenticatorIndex }

type InputSelectTOTP struct{}

var _ interaction.Input = &InputSelectTOTP{}
var _ nodes.InputCreateAuthenticatorTOTPSetup = &InputSelectTOTP{}

func (*InputSelectTOTP) IsInteractive() bool { return false }
func (i *InputSelectTOTP) SetupTOTP()        {}

type InputSetupPassword struct {
	Stage    string
	Password string
}

var _ interaction.Input = &InputSetupPassword{}
var _ nodes.InputCreateAuthenticatorPassword = &InputSetupPassword{}
var _ nodes.InputAuthenticationStage = &InputSetupPassword{}

func (*InputSetupPassword) IsInteractive() bool   { return true }
func (i *InputSetupPassword) GetPassword() string { return i.Password }
func (i *InputSetupPassword) GetAuthenticationStage() authn.AuthenticationStage {
	return authn.AuthenticationStage(i.Stage)
}

type InputResendCode struct{}

var _ interaction.Input = &InputResendCode{}

func (*InputResendCode) IsInteractive() bool { return true }
func (i *InputResendCode) DoResend()         {}

type InputAuthOOB struct {
	Code        string
	DeviceToken bool
}

var _ interaction.Input = &InputAuthOOB{}
var _ nodes.InputAuthenticationOOB = &InputAuthOOB{}
var _ nodes.InputCreateDeviceToken = &InputAuthOOB{}

func (*InputAuthOOB) IsInteractive() bool       { return true }
func (i *InputAuthOOB) GetOOBOTP() string       { return i.Code }
func (i *InputAuthOOB) CreateDeviceToken() bool { return i.DeviceToken }

type InputAuthTOTP struct {
	Code        string
	DeviceToken bool
}

var _ interaction.Input = &InputAuthTOTP{}
var _ nodes.InputAuthenticationTOTP = &InputAuthTOTP{}
var _ nodes.InputCreateDeviceToken = &InputAuthTOTP{}

func (*InputAuthTOTP) IsInteractive() bool       { return true }
func (i *InputAuthTOTP) GetTOTP() string         { return i.Code }
func (i *InputAuthTOTP) CreateDeviceToken() bool { return i.DeviceToken }

type InputAuthPassword struct {
	Stage       string
	Password    string
	DeviceToken bool
}

var _ interaction.Input = &InputAuthPassword{}
var _ nodes.InputAuthenticationPassword = &InputAuthPassword{}
var _ nodes.InputCreateDeviceToken = &InputAuthPassword{}
var _ nodes.InputAuthenticationStage = &InputAuthPassword{}

func (*InputAuthPassword) IsInteractive() bool       { return true }
func (i *InputAuthPassword) GetPassword() string     { return i.Password }
func (i *InputAuthPassword) CreateDeviceToken() bool { return i.DeviceToken }
func (i *InputAuthPassword) GetAuthenticationStage() authn.AuthenticationStage {
	return authn.AuthenticationStage(i.Stage)
}

type InputAuthRecoveryCode struct {
	Code string
}

var _ interaction.Input = &InputAuthRecoveryCode{}
var _ nodes.InputConsumeRecoveryCode = &InputAuthRecoveryCode{}

func (*InputAuthRecoveryCode) IsInteractive() bool       { return true }
func (i *InputAuthRecoveryCode) GetRecoveryCode() string { return i.Code }

type InputSetupOOB struct {
	InputType string
	Target    string
}

var _ interaction.Input = &InputSetupOOB{}
var _ nodes.InputCreateAuthenticatorOOBSetup = &InputSetupOOB{}

func (*InputSetupOOB) IsInteractive() bool { return true }
func (i *InputSetupOOB) GetOOBChannel() authn.AuthenticatorOOBChannel {
	switch i.InputType {
	case "email":
		return authn.AuthenticatorOOBChannelEmail
	case "phone":
		return authn.AuthenticatorOOBChannelSMS
	default:
		panic("webapp: unknown input type: " + i.InputType)
	}
}
func (i *InputSetupOOB) GetOOBTarget() string { return i.Target }

type InputSetupRecoveryCode struct{}

var _ interaction.Input = &InputSetupRecoveryCode{}
var _ nodes.InputGenerateRecoveryCodeEnd = &InputSetupRecoveryCode{}

func (*InputSetupRecoveryCode) IsInteractive() bool    { return true }
func (i *InputSetupRecoveryCode) ViewedRecoveryCodes() {}

type InputSetupTOTP struct {
	Code        string
	DisplayName string
}

var _ interaction.Input = &InputSetupTOTP{}
var _ nodes.InputCreateAuthenticatorTOTP = &InputSetupTOTP{}

func (*InputSetupTOTP) IsInteractive() bool          { return true }
func (i *InputSetupTOTP) GetTOTP() string            { return i.Code }
func (i *InputSetupTOTP) GetTOTPDisplayName() string { return i.DisplayName }

type InputOAuthCallback struct {
	ProviderAlias string

	Code             string
	Scope            string
	Error            string
	ErrorDescription string
	ErrorURI         string
}

var _ interaction.Input = &InputOAuthCallback{}
var _ nodes.InputUseIdentityOAuthUserInfo = &InputOAuthCallback{}

func (*InputOAuthCallback) IsInteractive() bool           { return false }
func (i *InputOAuthCallback) GetProviderAlias() string    { return i.ProviderAlias }
func (i *InputOAuthCallback) GetCode() string             { return i.Code }
func (i *InputOAuthCallback) GetScope() string            { return i.Scope }
func (i *InputOAuthCallback) GetError() string            { return i.Error }
func (i *InputOAuthCallback) GetErrorDescription() string { return i.ErrorDescription }
func (i *InputOAuthCallback) GetErrorURI() string         { return i.ErrorURI }

type InputVerificationCode struct {
	Code string
}

var _ interaction.Input = &InputVerificationCode{}
var _ nodes.InputVerifyIdentityCheckCode = &InputVerificationCode{}

func (*InputVerificationCode) IsInteractive() bool           { return true }
func (i *InputVerificationCode) GetVerificationCode() string { return i.Code }

type InputChangePassword struct {
	OldPassword string
	NewPassword string
}

var _ interaction.Input = &InputChangePassword{}
var _ nodes.InputChangePassword = &InputChangePassword{}

func (*InputChangePassword) IsInteractive() bool      { return true }
func (i *InputChangePassword) GetOldPassword() string { return i.OldPassword }
func (i *InputChangePassword) GetNewPassword() string { return i.NewPassword }

type InputForgotPassword struct {
	LoginID string
}

var _ interaction.Input = &InputForgotPassword{}
var _ nodes.InputForgotPasswordSelectLoginID = &InputForgotPassword{}

func (*InputForgotPassword) IsInteractive() bool  { return true }
func (i *InputForgotPassword) GetLoginID() string { return i.LoginID }

type InputResetPassword struct {
	Code     string
	Password string
}

var _ interaction.Input = &InputResetPassword{}
var _ nodes.InputResetPasswordByCode = &InputResetPassword{}

func (*InputResetPassword) IsInteractive() bool      { return true }
func (i *InputResetPassword) GetCode() string        { return i.Code }
func (i *InputResetPassword) GetNewPassword() string { return i.Password }
