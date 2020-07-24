package interaction

import (
	"github.com/authgear/authgear-server/pkg/auth/config"
	"github.com/authgear/authgear-server/pkg/auth/dependency/authenticator"
	"github.com/authgear/authgear-server/pkg/auth/dependency/identity"
	"github.com/authgear/authgear-server/pkg/auth/event"
	"github.com/authgear/authgear-server/pkg/auth/model"
	"github.com/authgear/authgear-server/pkg/clock"
	"github.com/authgear/authgear-server/pkg/core/authn"
	"github.com/authgear/authgear-server/pkg/log"
)

//go:generate mockgen -source=provider.go -destination=provider_mock_test.go -package interaction_test

type IdentityProvider interface {
	Get(userID string, typ authn.IdentityType, id string) (*identity.Info, error)
	// GetByClaims return user ID and information about the identity the matches the provided authgear claims.
	GetByClaims(typ authn.IdentityType, claims map[string]interface{}) (string, *identity.Info, error)
	// GetByUserAndClaims return user's identity that matches the provided authgear claims.
	//
	// Given that user id is provided, the matching rule of this function is less strict than GetByClaims.
	// For example, login id identity needs match both key and value and oauth identity only needs to match provider id.
	// This function is currently in used by remove identity interaction.
	GetByUserAndClaims(typ authn.IdentityType, userID string, claims map[string]interface{}) (*identity.Info, error)
	// ListByClaims return list of identities the matches the provided OIDC standard claims.
	ListByClaims(claims map[string]string) ([]*identity.Info, error)
	ListByUser(userID string) ([]*identity.Info, error)
	New(userID string, typ authn.IdentityType, claims map[string]interface{}) (*identity.Info, error)
	WithClaims(ii *identity.Info, claims map[string]interface{}) (*identity.Info, error)
	CreateAll(is []*identity.Info) error
	UpdateAll(is []*identity.Info) error
	DeleteAll(is []*identity.Info) error
	Validate(is []*identity.Info) error
	// RelateIdentityToAuthenticator tells if authenticatorSpec is compatible with and related to identityInfo.
	//
	// A authenticatorSpec is compatible with identityInfo if authenticator can be used as authentication for the identity.
	// For example, OAuth identity is incompatible with any authenticator because the identity itself implicit authenticates.
	// For example, login ID identity of login ID type username is incompatible with OOB OTP authenticator because
	// OOB OTP authenticator requires email or phone.
	//
	// If authenticatorSpec is incompatible with identityInfo, nil is returned.
	//
	// Otherwise authenticatorSpec is further checked if it is related to identityInfo.
	// If authenticatorSpec is related to identityInfo, authenticatorSpec.Props is mutated in-place.
	// Currently on the following case mutation would occur.
	//
	//   - login ID identity of login ID type email or phone and OOB OTP authenticator.
	RelateIdentityToAuthenticator(identityInfo *identity.Info, authenticatorSpec *authenticator.Spec) *authenticator.Spec
	CheckIdentityDuplicated(is *identity.Info) error
}

type AuthenticatorProvider interface {
	Get(userID string, typ authn.AuthenticatorType, id string) (*authenticator.Info, error)
	List(userID string, typ authn.AuthenticatorType) ([]*authenticator.Info, error)
	ListByIdentity(userID string, ii *identity.Info) ([]*authenticator.Info, error)
	New(userID string, spec authenticator.Spec, secret string) ([]*authenticator.Info, error)
	// WithSecret returns bool to indicate if the authenticator changed
	WithSecret(userID string, a *authenticator.Info, secret string) (bool, *authenticator.Info, error)
	CreateAll(userID string, ais []*authenticator.Info) error
	UpdateAll(userID string, ais []*authenticator.Info) error
	DeleteAll(userID string, ais []*authenticator.Info) error
	Authenticate(userID string, spec authenticator.Spec, state *map[string]string, secret string) (*authenticator.Info, error)
	VerifySecret(userID string, a *authenticator.Info, secret string) error
}

type UserProvider interface {
	Create(userID string, metadata map[string]interface{}, identities []*identity.Info, authenticators []*authenticator.Info) error
	Get(userID string) (*model.User, error)
}

type OOBProvider interface {
	GenerateCode(channel authn.AuthenticatorOOBChannel) string
}

type HookProvider interface {
	DispatchEvent(payload event.Payload, user *model.User) error
}

type Logger struct{ *log.Logger }

func NewLogger(lf *log.Factory) Logger {
	return Logger{lf.New("interaction")}
}

type Provider struct {
	Clock         clock.Clock
	Logger        Logger
	Identity      IdentityProvider
	Authenticator AuthenticatorProvider
	User          UserProvider
	OOB           OOBProvider `wire:"-"`
	Hooks         HookProvider
	Config        *config.AuthenticationConfig
}
