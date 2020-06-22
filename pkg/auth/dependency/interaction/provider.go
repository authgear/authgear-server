package interaction

import (
	"time"

	"github.com/sirupsen/logrus"

	"github.com/skygeario/skygear-server/pkg/auth/dependency/authenticator"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/authenticator/oob"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/identity"
	"github.com/skygeario/skygear-server/pkg/auth/event"
	"github.com/skygeario/skygear-server/pkg/auth/model"
	"github.com/skygeario/skygear-server/pkg/clock"
	"github.com/skygeario/skygear-server/pkg/core/authn"
	"github.com/skygeario/skygear-server/pkg/core/config"
)

//go:generate mockgen -source=provider.go -destination=provider_mock_test.go -package interaction_test

type Store interface {
	Create(i *Interaction) error
	Get(token string) (*Interaction, error)
	Update(i *Interaction) error
	Delete(i *Interaction) error
}

type IdentityProvider interface {
	Get(userID string, typ authn.IdentityType, id string) (*identity.Info, error)
	// GetByClaims return user ID and information about the identity the matches the provided skygear claims.
	GetByClaims(typ authn.IdentityType, claims map[string]interface{}) (string, *identity.Info, error)
	// GetByUserAndClaims return user's identity that matches the provide skygear claims.
	//
	// Given that user id is provided, the matching rule of this function is less strict than GetByClaims.
	// For example, login id identity needs match both key and value and oauth identity only needs to match provider id.
	// This function is currently in used by remove identity interaction.
	GetByUserAndClaims(typ authn.IdentityType, userID string, claims map[string]interface{}) (*identity.Info, error)
	// ListByClaims return list of identities the matches the provided OIDC standard claims.
	ListByClaims(claims map[string]string) ([]*identity.Info, error)
	ListByUser(userID string) ([]*identity.Info, error)
	New(userID string, typ authn.IdentityType, claims map[string]interface{}) (*identity.Info, error)
	WithClaims(userID string, ii *identity.Info, claims map[string]interface{}) (*identity.Info, error)
	CreateAll(userID string, is []*identity.Info) error
	UpdateAll(userID string, is []*identity.Info) error
	DeleteAll(userID string, is []*identity.Info) error
	Validate(is []*identity.Info) error
	// RelateIdentityToAuthenticator tells if authenticatorSpec is compatible with and related to identitySpec.
	//
	// A authenticatorSpec is compatible with identitySpec if authenticator can be used as authentication for the identity.
	// For example, OAuth identity is incompatible with any authenticator because the identity itself implicit authenticates.
	// For example, login ID identity of login ID type username is incompatible with OOB OTP authenticator because
	// OOB OTP authenticator requires email or phone.
	//
	// If authenticatorSpec is incompatible with identitySpec, nil is returned.
	//
	// Otherwise authenticatorSpec is further checked if it is related to identitySpec.
	// If authenticatorSpec is related to identitySpec, authenticatorSpec.Props is mutated in-place.
	// Currently on the following case mutation would occur.
	//
	//   - login ID identity of login ID type email or phone and OOB OTP authenticator.
	RelateIdentityToAuthenticator(identitySpec identity.Spec, authenticatorSpec *authenticator.Spec) *authenticator.Spec
	CheckIdentityDuplicated(is *identity.Info, userID string) error
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
	Create(userID string, metadata map[string]interface{}, identities []*identity.Info) error
	Get(userID string) (*model.User, error)
}

type OOBProvider interface {
	GenerateCode() string
	SendCode(opts oob.SendCodeOptions) error
}

type HookProvider interface {
	DispatchEvent(payload event.Payload, user *model.User) error
}

// TODO(interaction): configurable lifetime
const interactionIdleTimeout = 5 * time.Minute

// NOTE(interaction): save-commit
// SaveInteraction and Commit are mutually exclusively within a request.
// You either do something with the interaction, SaveInteraction and return the token.
// Or do something with the interaction, Commit and discard the interaction.
//
// Mixing SaveInteraction and Commit may lead to data corruption.
// For example, given the following call sequence in a function.
//
// PerformAction
// SaveInteraction
// Commit
//
// If Commit fails for some reason, the interaction has already been mutated by SaveInteraction.
// If the function is retried, PerformAction is applied twice, leading to data corruption.

type Provider struct {
	Store         Store
	Clock         clock.Clock
	Logger        *logrus.Entry
	Identity      IdentityProvider
	Authenticator AuthenticatorProvider
	User          UserProvider
	OOB           OOBProvider
	Hooks         HookProvider
	Config        *config.AuthenticationConfiguration
}

func (p *Provider) GetInteraction(token string) (*Interaction, error) {
	i, err := p.Store.Get(token)
	if err != nil {
		return nil, err
	}

	return i, nil
}

func (p *Provider) SaveInteraction(i *Interaction) (string, error) {
	if i.committed {
		panic("interaction: see NOTE(interaction): save-commit")
	}

	if i.Token == "" {
		i.Token = generateToken()
		i.CreatedAt = p.Clock.NowUTC()
		i.ExpireAt = i.CreatedAt.Add(interactionIdleTimeout)
		if err := p.Store.Create(i); err != nil {
			return "", err
		}
	} else {
		i.ExpireAt = p.Clock.NowUTC().Add(interactionIdleTimeout)
		if err := p.Store.Update(i); err != nil {
			return "", err
		}
	}

	i.saved = true

	return i.Token, nil
}
