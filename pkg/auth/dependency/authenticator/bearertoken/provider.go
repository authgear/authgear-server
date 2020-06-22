package bearertoken

import (
	"errors"

	"github.com/skygeario/skygear-server/pkg/auth/config"
	"github.com/skygeario/skygear-server/pkg/clock"
	"github.com/skygeario/skygear-server/pkg/core/uuid"
)

type Provider struct {
	Store  *Store
	Config *config.AuthenticatorBearerTokenConfig
	Clock  clock.Clock
}

func (p *Provider) Get(userID string, id string) (*Authenticator, error) {
	return p.Store.Get(userID, id)
}

func (p *Provider) GetByToken(userID string, token string) (*Authenticator, error) {
	return p.Store.GetByToken(userID, token)
}

func (p *Provider) List(userID string) ([]*Authenticator, error) {
	return p.Store.List(userID)
}

func (p *Provider) RevokeAll(userID string) error {
	return p.Store.DeleteAll(userID)
}

func (p *Provider) DeleteByParentID(parentID string) error {
	return p.Store.DeleteAllByParentID(parentID)
}

func (p *Provider) CleanupExpiredAuthenticators(userID string) error {
	return p.Store.DeleteAllExpired(userID, p.Clock.NowUTC())
}

func (p *Provider) New(userID string, parentID string) *Authenticator {
	a := &Authenticator{
		ID:       uuid.New(),
		UserID:   userID,
		ParentID: parentID,
		Token:    GenerateToken(),
	}
	return a
}

func (p *Provider) Create(a *Authenticator) error {
	now := p.Clock.NowUTC()
	expireAt := now.Add(p.Config.ExpireIn.Duration())
	a.CreatedAt = now
	a.ExpireAt = expireAt

	return p.Store.Create(a)
}

func (p *Provider) Authenticate(authenticator *Authenticator, token string) error {
	ok := VerifyToken(authenticator.Token, token)
	if !ok {
		return errors.New("invalid bearer token")
	}

	return nil
}
