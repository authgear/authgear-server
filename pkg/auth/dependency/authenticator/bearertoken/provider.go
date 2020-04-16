package bearertoken

import (
	"errors"
	gotime "time"

	"github.com/skygeario/skygear-server/pkg/core/config"
	"github.com/skygeario/skygear-server/pkg/core/time"
	"github.com/skygeario/skygear-server/pkg/core/uuid"
)

type Provider struct {
	Store  *Store
	Config *config.AuthenticatorBearerTokenConfiguration
	Time   time.Provider
}

func (p *Provider) GetByToken(userID string, token string) (*Authenticator, error) {
	return p.Store.GetByToken(userID, token)
}

func (p *Provider) RevokeAll(userID string) error {
	return p.Store.DeleteAll(userID)
}

func (p *Provider) DeleteByParentID(parentID string) error {
	return p.Store.DeleteAllByParentID(parentID)
}

func (p *Provider) CleanupExpiredAuthenticators(userID string) error {
	return p.Store.DeleteAllExpired(userID, p.Time.NowUTC())
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
	now := p.Time.NowUTC()
	expireAt := now.Add(gotime.Duration(p.Config.ExpireInDays) * gotime.Hour * 24)
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
