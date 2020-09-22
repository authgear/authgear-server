package loader

import (
	"github.com/authgear/authgear-server/pkg/admin/model"
	"github.com/authgear/authgear-server/pkg/lib/authn"
	"github.com/authgear/authgear-server/pkg/lib/authn/authenticator"
	"github.com/authgear/authgear-server/pkg/lib/authn/identity"
	"github.com/authgear/authgear-server/pkg/lib/interaction"
)

type InteractionService interface {
	Perform(intent interaction.Intent, input interface{}) (*interaction.Graph, error)
}

type removeIdentityInput struct {
	identityInfo *identity.Info
}

func (i *removeIdentityInput) GetIdentityType() authn.IdentityType {
	return i.identityInfo.Type
}
func (i *removeIdentityInput) GetIdentityID() string {
	return i.identityInfo.ID
}

type addIdentityInput struct {
	identityDef model.IdentityDef
}

func (i *addIdentityInput) SkipVerification() bool {
	return true
}
func (i *addIdentityInput) Input() interface{} {
	return i.identityDef
}

type removeAuthenticatorInput struct {
	authenticatorInfo *authenticator.Info
}

func (i *removeAuthenticatorInput) GetAuthenticatorType() authn.AuthenticatorType {
	return i.authenticatorInfo.Type
}
func (i *removeAuthenticatorInput) GetAuthenticatorID() string {
	return i.authenticatorInfo.ID
}
func (i *removeAuthenticatorInput) BypassMFARequirement() bool {
	return true
}

type addPasswordInput struct {
	inner    interface{}
	password string
}

func (i *addPasswordInput) GetPassword() string {
	return i.password
}
func (i *addPasswordInput) Input() interface{} {
	return i.inner
}

type resetPasswordInput struct {
	userID   string
	password string
}

func (i *resetPasswordInput) GetResetPasswordUserID() string {
	return i.userID
}

func (i *resetPasswordInput) GetNewPassword() string {
	return i.password
}

type createUserInput struct {
	identityDef model.IdentityDef
	password    string
}

func (i *createUserInput) SkipVerification() bool {
	return true
}
func (i *createUserInput) SkipMFASetup() bool {
	return true
}
func (i *createUserInput) GetPassword() string {
	return i.password
}
func (i *createUserInput) Input() interface{} {
	return i.identityDef
}
