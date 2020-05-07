package interaction

import (
	gotime "time"

	"github.com/sirupsen/logrus"

	"github.com/skygeario/skygear-server/pkg/auth/dependency/authenticator/oob"
	"github.com/skygeario/skygear-server/pkg/core/authn"
	"github.com/skygeario/skygear-server/pkg/core/config"
	"github.com/skygeario/skygear-server/pkg/core/time"
)

//go:generate mockgen -source=provider.go -destination=provider_mock_test.go -package interaction_test

type Store interface {
	Create(i *Interaction) error
	Get(token string) (*Interaction, error)
	Update(i *Interaction) error
	Delete(i *Interaction) error
}

type IdentityProvider interface {
	Get(userID string, typ authn.IdentityType, id string) (*IdentityInfo, error)
	// GetByClaims return user ID and information about the identity the matches the provided skygear claims.
	GetByClaims(typ authn.IdentityType, claims map[string]interface{}) (string, *IdentityInfo, error)
	// ListByClaims return list of identities the matches the provided OIDC standard claims.
	ListByClaims(claims map[string]string) ([]*IdentityInfo, error)
	ListByUser(userID string) ([]*IdentityInfo, error)
	New(userID string, typ authn.IdentityType, claims map[string]interface{}) *IdentityInfo
	WithClaims(userID string, ii *IdentityInfo, claims map[string]interface{}) *IdentityInfo
	CreateAll(userID string, is []*IdentityInfo) error
	UpdateAll(userID string, is []*IdentityInfo) error
	Validate(is []*IdentityInfo) error
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
	RelateIdentityToAuthenticator(identitySpec IdentitySpec, authenticatorSpec *AuthenticatorSpec) *AuthenticatorSpec
}

type AuthenticatorProvider interface {
	Get(userID string, typ authn.AuthenticatorType, id string) (*AuthenticatorInfo, error)
	List(userID string, typ authn.AuthenticatorType) ([]*AuthenticatorInfo, error)
	ListByIdentity(userID string, ii *IdentityInfo) ([]*AuthenticatorInfo, error)
	New(userID string, spec AuthenticatorSpec, secret string) ([]*AuthenticatorInfo, error)
	CreateAll(userID string, ais []*AuthenticatorInfo) error
	Authenticate(userID string, spec AuthenticatorSpec, state *map[string]string, secret string) (*AuthenticatorInfo, error)
}

type UserProvider interface {
	Create(userID string, metadata map[string]interface{}, identities []*IdentityInfo) error
}

type OOBProvider interface {
	GenerateCode() string
	SendCode(opts oob.SendCodeOptions) error
}

// TODO(interaction): configurable lifetime
const interactionIdleTimeout = 5 * gotime.Minute

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
	Time          time.Provider
	Logger        *logrus.Entry
	Identity      IdentityProvider
	Authenticator AuthenticatorProvider
	User          UserProvider
	OOB           OOBProvider
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
		i.CreatedAt = p.Time.NowUTC()
		i.ExpireAt = i.CreatedAt.Add(interactionIdleTimeout)
		if err := p.Store.Create(i); err != nil {
			return "", err
		}
	} else {
		i.ExpireAt = p.Time.NowUTC().Add(interactionIdleTimeout)
		if err := p.Store.Update(i); err != nil {
			return "", err
		}
	}

	i.saved = true

	return i.Token, nil
}
