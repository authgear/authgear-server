package interaction

import (
	gotime "time"

	"github.com/sirupsen/logrus"
	"github.com/skygeario/skygear-server/pkg/core/authn"
	"github.com/skygeario/skygear-server/pkg/core/config"
	"github.com/skygeario/skygear-server/pkg/core/time"
)

type IdentityProvider interface {
	Get(userID string, typ authn.IdentityType, id string) (*IdentityInfo, error)
	// GetByClaims return user ID and information about the identity the matches the provided skygear claims.
	GetByClaims(typ authn.IdentityType, claims map[string]interface{}) (string, *IdentityInfo, error)
	// ListByClaims return list of identities the matches the provided OIDC standard claims.
	ListByClaims(claims map[string]string) ([]*IdentityInfo, error)
	New(userID string, typ authn.IdentityType, claims map[string]interface{}) *IdentityInfo
	CreateAll(userID string, is []*IdentityInfo) error
	Validate(is []*IdentityInfo) error
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
	SendCode(spec AuthenticatorSpec, code string) error
}

// TODO(interaction): configurable lifetime
const interactionIdleTimeout = 5 * gotime.Minute

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
	return i.Token, nil
}
