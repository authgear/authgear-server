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
	LoginID    string
	LoginIDKey string
	IdentityID string
	Channel    model.AuthenticatorOOBChannel
	isUpdate   bool
}

type StartIdentityWithVerificationOutput struct {
	Token            string
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
	Token        string
	Channel      model.AuthenticatorOOBChannel
	Code         string
	IdentityInfo *identity.Info
}

type CreateIdentityWithVerificationInput struct {
	LoginID    string
	LoginIDKey string
	Code       string
	Token      string
	Channel    model.AuthenticatorOOBChannel
}

type CreateIdentityWithVerificationOutput struct {
	IdentityInfo *identity.Info
}

type UpdateIdentityWithVerificationInput struct {
	LoginID    string
	LoginIDKey string
	IdentityID string
	Code       string
	Token      string
	Channel    model.AuthenticatorOOBChannel
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

type AddBiometricInput struct {
	JWTToken string
}

type AddBiometricOutput struct {
	IdentityInfo *identity.Info
}

type RemoveBiometricInput struct {
	IdentityID string
}

type RemoveBiometricOuput struct {
	IdentityInfo *identity.Info
}

type AddUsernameInput struct {
	LoginID    string
	LoginIDKey string
}

type AddUsernameOutput struct {
	IdentityInfo *identity.Info
}

type UpdateUsernameInput struct {
	LoginID    string
	LoginIDKey string
	IdentityID string
}

type UpdateUsernameOutput struct {
	IdentityInfo *identity.Info
}

type RemoveUsernameInput struct {
	IdentityID string
}

type RemoveUsernameOutput struct {
	IdentityInfo *identity.Info
}

type RemoveEmailInput struct {
	IdentityID string
}

type RemoveEmailOutput struct {
	IdentityInfo *identity.Info
}

type RemovePhoneNumberInput struct {
	IdentityID string
}

type RemovePhoneNumberOutput struct {
	IdentityInfo *identity.Info
}
