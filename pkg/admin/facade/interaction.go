package facade

import (
	"github.com/authgear/authgear-server/pkg/admin/model"
	apimodel "github.com/authgear/authgear-server/pkg/api/model"
	"github.com/authgear/authgear-server/pkg/lib/authn"
	"github.com/authgear/authgear-server/pkg/lib/authn/authenticator"
	"github.com/authgear/authgear-server/pkg/lib/authn/identity"
	"github.com/authgear/authgear-server/pkg/lib/interaction"
	"github.com/authgear/authgear-server/pkg/lib/interaction/nodes"
)

type InteractionService interface {
	Perform(intent interaction.Intent, input interface{}) (*interaction.Graph, error)
}

// adminAPIOp represents common characteristics applicable to Admin API operations.
type adminAPIOp struct{}

// BypassInteractionIPRateLimit indicates Admin API operations does not check
// for interaction rate limits on IP; Admin API requests are commonly issued
// from single IP.
func (adminAPIOp) BypassInteractionIPRateLimit() bool { return true }

// BypassMFARequirement indicates Admin API operations does not check
// for MFA requirement invariant; Admin API can create/delete secondary
// authenticators freely.
func (adminAPIOp) BypassMFARequirement() bool { return true }

// SkipVerification indicates Admin API operations does not check
// for verification requirement; user would be prompt to verify identities
// if required next login.
func (adminAPIOp) SkipVerification() bool { return true }

// SkipMFASetup indicates Admin API operations does not setup required
// secondary authenticator; user would be prompt to setup MFA if required
// next login.
func (adminAPIOp) SkipMFASetup() bool { return true }

// BypassPublicSignupDisabled indicates Admin API operations bypass
// disabled public signup; creating users through admin APIs do not count as
// public signup.
func (adminAPIOp) BypassPublicSignupDisabled() bool {
	return true
}

// BypassLoginIDEmailBlocklistAllowlist indicates Admin API operations
// bypass email domains blocklist allowlist checking; Admin API can create
// email login id that doesn't have blocklist allowlist restriction
func (adminAPIOp) BypassLoginIDEmailBlocklistAllowlist() bool {
	return true
}

// IsAdminAPI indicates this is admin operation
func (adminAPIOp) IsAdminAPI() bool {
	return true
}

type removeIdentityInput struct {
	adminAPIOp
	identityInfo *identity.Info
}

var _ nodes.InputRemoveIdentity = &removeIdentityInput{}

func (i *removeIdentityInput) GetIdentityType() apimodel.IdentityType {
	return i.identityInfo.Type
}
func (i *removeIdentityInput) GetIdentityID() string {
	return i.identityInfo.ID
}

type addIdentityInput struct {
	adminAPIOp
	identityDef model.IdentityDef
}

func (i *addIdentityInput) Input() interface{} {
	return i.identityDef
}

type updateIdentityInput struct {
	adminAPIOp
	identityDef model.IdentityDef
}

func (i *updateIdentityInput) Input() interface{} {
	return i.identityDef
}

type removeAuthenticatorInput struct {
	adminAPIOp
	authenticatorInfo *authenticator.Info
}

var _ nodes.InputRemoveAuthenticator = &removeAuthenticatorInput{}

func (i *removeAuthenticatorInput) GetAuthenticatorType() apimodel.AuthenticatorType {
	return i.authenticatorInfo.Type
}
func (i *removeAuthenticatorInput) GetAuthenticatorID() string {
	return i.authenticatorInfo.ID
}

type addPasswordInput struct {
	adminAPIOp
	inner    interface{}
	password string
}

var _ nodes.InputCreateAuthenticatorPassword = &addPasswordInput{}
var _ nodes.InputAuthenticationStage = &addPasswordInput{}

func (i *addPasswordInput) GetPassword() string {
	return i.password
}
func (i *addPasswordInput) Input() interface{} {
	return i.inner
}

func (i *addPasswordInput) GetAuthenticationStage() authn.AuthenticationStage {
	return authn.AuthenticationStagePrimary
}

type resetPasswordInput struct {
	adminAPIOp
	userID           string
	password         string
	generatePassword bool
	sendPassword     bool
	changeOnLogin    bool
}

var _ nodes.InputResetPassword = &resetPasswordInput{}

func (i *resetPasswordInput) GetResetPasswordUserID() string {
	return i.userID
}
func (i *resetPasswordInput) GetNewPassword() string {
	return i.password
}

func (i *resetPasswordInput) GeneratePassword() bool {
	return i.generatePassword
}

func (i *resetPasswordInput) SendPassword() bool {
	return i.sendPassword
}

func (i *resetPasswordInput) ChangeOnLogin() bool {
	return i.changeOnLogin
}

type createUserInput struct {
	adminAPIOp
	identityDef model.IdentityDef
	password    string
}

var _ nodes.InputCreateAuthenticatorPassword = &createUserInput{}
var _ nodes.InputAuthenticationStage = &createUserInput{}
var _ nodes.InputPromptCreatePasskey = &createUserInput{}

func (i *createUserInput) GetPassword() string {
	return i.password
}
func (i *createUserInput) Input() interface{} {
	return i.identityDef
}
func (i *createUserInput) GetAuthenticationStage() authn.AuthenticationStage {
	return authn.AuthenticationStagePrimary
}

// IsSkipped implements InputPromptCreatePasskey
func (i *createUserInput) IsSkipped() bool {
	return true
}

// GetAttestationResponse implements InputPromptCreatePasskey
func (i *createUserInput) GetAttestationResponse() []byte {
	return nil
}
