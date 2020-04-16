package password

import (
	"sort"

	"github.com/sirupsen/logrus"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/audit"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/passwordhistory"
	"github.com/skygeario/skygear-server/pkg/core/config"
	"github.com/skygeario/skygear-server/pkg/core/errors"
	pwd "github.com/skygeario/skygear-server/pkg/core/password"
	"github.com/skygeario/skygear-server/pkg/core/time"
	"github.com/skygeario/skygear-server/pkg/core/uuid"
)

type Provider struct {
	Store           *Store
	Config          *config.AuthenticatorPasswordConfiguration
	Time            time.Provider
	Logger          *logrus.Entry
	PasswordHistory passwordhistory.Store
	PasswordChecker *audit.PasswordChecker
}

func (p *Provider) Get(userID string, id string) (*Authenticator, error) {
	return p.Store.Get(userID, id)
}

func (p *Provider) Delete(a *Authenticator) error {
	return p.Store.Delete(a.ID)
}

func (p *Provider) List(userID string) ([]*Authenticator, error) {
	authenticators, err := p.Store.List(userID)
	if err != nil {
		return nil, err
	}

	sortAuthenticators(authenticators)
	return authenticators, nil
}

func (p *Provider) New(userID string, password string) (*Authenticator, error) {
	err := p.isPasswordAllowed(userID, password)
	if err != nil {
		return nil, err
	}

	hash, err := pwd.Hash([]byte(password))
	if err != nil {
		panic(errors.Newf("password: failed to hash password: %w", err))
	}

	a := &Authenticator{
		ID:           uuid.New(),
		UserID:       userID,
		PasswordHash: hash,
	}
	return a, nil
}

func (p *Provider) Create(a *Authenticator) error {
	return p.Store.Create(a)
}

func (p *Provider) Authenticate(a *Authenticator, password string) error {
	err := pwd.Compare([]byte(password), a.PasswordHash)
	if err != nil {
		return err
	}

	migrated, err := pwd.TryMigrate([]byte(password), &a.PasswordHash)
	if err != nil {
		p.Logger.WithError(err).WithField("authenticator_id", a.ID).
			Warn("Failed to migrate password")
		return nil
	}

	if migrated {
		err = p.Store.UpdatePasswordHash(a)
		if err != nil {
			p.Logger.WithError(err).WithField("authenticator_id", a.ID).
				Warn("Failed to save migrated password")
			return nil
		}
	}

	return nil
}

func (p *Provider) isPasswordAllowed(userID string, password string) error {
	return p.PasswordChecker.ValidatePassword(audit.ValidatePasswordPayload{
		AuthID:        userID,
		PlainPassword: password,
	})
}

func (p *Provider) UpdatePassword(a *Authenticator, password string) error {
	err := p.isPasswordAllowed(a.UserID, password)
	if err != nil {
		return err
	}

	// If password is not changed, skip the logic.
	if pwd.Compare([]byte(password), a.PasswordHash) == nil {
		return nil
	}

	hash, err := pwd.Hash([]byte(password))
	if err != nil {
		return err
	}

	a.PasswordHash = hash
	err = p.Store.UpdatePasswordHash(a)
	if err != nil {
		return err
	}

	err = p.PasswordHistory.CreatePasswordHistory(a.UserID, hash, p.Time.NowUTC())
	if err != nil {
		return err
	}

	return nil
}

func sortAuthenticators(as []*Authenticator) {
	sort.Slice(as, func(i, j int) bool {
		return as[i].ID < as[j].ID
	})
}
