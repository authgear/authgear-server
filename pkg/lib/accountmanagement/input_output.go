package accountmanagement

import (
	"github.com/go-webauthn/webauthn/protocol"

	"github.com/authgear/authgear-server/pkg/api/model"
	"github.com/authgear/authgear-server/pkg/lib/authn/identity"
	"github.com/authgear/authgear-server/pkg/util/uuid"
)

type StartAddingInput struct {
	UserID                                          string
	Alias                                           string
	RedirectURI                                     string
	IncludeStateAuthorizationURLAndBindStateToToken bool
}

type StartAddingOutput struct {
	Token            string `json:"token,omitempty"`
	AuthorizationURL string `json:"authorization_url,omitempty"`
}

type FinishAddingInput struct {
	UserID string
	Token  string
	Query  string
}

type FinishAddingOutput struct {
	// It is intentionally empty.
}
type ResendOTPCodeInput struct {
	Channel      model.AuthenticatorOOBChannel
	LoginID      string
	isSwitchPage bool
}

type sendOTPCodeInput struct {
	Channel  model.AuthenticatorOOBChannel
	Target   string
	isResend bool
}

type StartCreateIdentityWithVerificationInput struct {
	LoginID    string
	LoginIDKey string
	Channel    model.AuthenticatorOOBChannel
}

type StartUpdateIdentityWithVerificationInput struct {
	LoginID    string
	LoginIDKey string
	IdentityID string
	Channel    model.AuthenticatorOOBChannel
}

type startIdentityWithVerificationInput struct {
	LoginID      string
	LoginIDKey   string
	IdentityID   string
	IdentitySpec *identity.Spec
	Channel      model.AuthenticatorOOBChannel
	isUpdate     bool
}

type StartIdentityWithVerificationOutput struct {
	IdentityInfo     *identity.Info
	NeedVerification bool
}

type StartCreateIdentityWithVerificationOutput struct {
	Token            string
	IdentityInfo     *identity.Info
	NeedVerification bool
}

type StartUpdateIdentityWithVerificationOutput struct {
	Token            string
	IdentityInfo     *identity.Info
	NeedVerification bool
}

type ResumeAddingIdentityWithVerificationInput struct {
	Token string
}

type ResumeAddingIdentityWithVerificationOutput struct {
	Token          string
	LoginID        string
	LoginIDKeyType model.LoginIDKeyType
	IdentityID     string
}

type verifyIdentityInput struct {
	UserID       string
	Token        *Token
	Channel      model.AuthenticatorOOBChannel
	Code         string
	IdentityInfo *identity.Info
}

type AddIdentityEmailWithVerificationInput struct {
	LoginID    string
	LoginIDKey string
	Code       string
	Token      string
	Channel    model.AuthenticatorOOBChannel
}

type AddIdentityPhoneNumberWithVerificationInput struct {
	LoginID    string
	LoginIDKey string
	Code       string
	Token      string
	Channel    model.AuthenticatorOOBChannel
}

type UpdateIdentityEmailWithVerificationInput struct {
	LoginID    string
	LoginIDKey string
	IdentityID string
	Code       string
	Token      string
	Channel    model.AuthenticatorOOBChannel
}

type UpdateIdentityPhoneNumberWithVerificationInput struct {
	LoginID    string
	LoginIDKey string
	IdentityID string
	Code       string
	Token      string
	Channel    model.AuthenticatorOOBChannel
}

type CreateIdentityWithVerificationInput struct {
	IdentitySpec *identity.Spec
	Code         string
	Token        *Token
	Channel      model.AuthenticatorOOBChannel
}

type CreateIdentityWithVerificationOutput struct {
	IdentityInfo *identity.Info
}

type UpdateIdentityWithVerificationInput struct {
	IdentityID   string
	IdentitySpec *identity.Spec
	Code         string
	Token        *Token
	Channel      model.AuthenticatorOOBChannel
}

type UpdateIdentityWithVerificationOutput struct {
	IdentityInfo *identity.Info
}

type ChangePasswordInput struct {
	OAuthSessionID string
	RedirectURI    string
	OldPassword    string
	NewPassword    string
}

type ChangePasswordOutput struct {
	RedirectURI string
}

type CreateAdditionalPasswordInput struct {
	NewAuthenticatorID string
	UserID             string
	Password           string
}

type AddPasskeyInput struct {
	CreationResponse *protocol.CredentialCreationResponse
}

type AddPasskeyOutput struct {
	IdentityInfo *identity.Info
}

func NewCreateAdditionalPasswordInput(userID string, password string) CreateAdditionalPasswordInput {
	return CreateAdditionalPasswordInput{
		NewAuthenticatorID: uuid.New(),
		UserID:             userID,
		Password:           password,
	}
}

type RemovePasskeyInput struct {
	IdentityID string
}

type RemovePasskeyOutput struct {
	IdentityInfo *identity.Info
}

type AddIdentityBiometricInput struct {
	JWTToken string
}

type AddIdentityBiometricOutput struct {
	IdentityInfo *identity.Info
}

type RemoveIdentityBiometricInput struct {
	IdentityID string
}

type RemoveIdentityBiometricOuput struct {
	IdentityInfo *identity.Info
}

type AddIdentityUsernameInput struct {
	LoginID    string
	LoginIDKey string
}

type AddIdentityUsernameOutput struct {
	IdentityInfo *identity.Info
}

type UpdateIdentityUsernameInput struct {
	LoginID    string
	LoginIDKey string
	IdentityID string
}

type UpdateIdentityUsernameOutput struct {
	IdentityInfo *identity.Info
}

type RemoveIdentityUsernameInput struct {
	IdentityID string
}

type RemoveIdentityUsernameOutput struct {
	IdentityInfo *identity.Info
}

type RemoveIdentityEmailInput struct {
	IdentityID string
}

type RemoveIdentityEmailOutput struct {
	IdentityInfo *identity.Info
}

type RemoveIdentityPhoneNumberInput struct {
	IdentityID string
}

type RemoveIdentityPhoneNumberOutput struct {
	IdentityInfo *identity.Info
}
