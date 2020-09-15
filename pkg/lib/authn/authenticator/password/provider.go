package password

import (
	"fmt"
	"sort"

	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/lib/infra/task"
	"github.com/authgear/authgear-server/pkg/lib/tasks"
	"github.com/authgear/authgear-server/pkg/util/clock"
	"github.com/authgear/authgear-server/pkg/util/log"
	pwd "github.com/authgear/authgear-server/pkg/util/password"
	"github.com/authgear/authgear-server/pkg/util/uuid"
)

type Logger struct{ *log.Logger }

func NewLogger(lf *log.Factory) Logger { return Logger{lf.New("password")} }

type Provider struct {
	Store           *Store
	Config          *config.AuthenticatorPasswordConfig
	Clock           clock.Clock
	Logger          Logger
	PasswordHistory *HistoryStore
	PasswordChecker *Checker
	TaskQueue       task.Queue
}

func (p *Provider) Get(userID string, id string) (*Authenticator, error) {
	return p.Store.Get(userID, id)
}

func (p *Provider) GetMany(ids []string) ([]*Authenticator, error) {
	return p.Store.GetMany(ids)
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

func (p *Provider) New(userID string, password string, isDefault bool, kind string) (*Authenticator, error) {
	authen := &Authenticator{
		ID:        uuid.New(),
		Labels:    make(map[string]interface{}),
		UserID:    userID,
		IsDefault: isDefault,
		Kind:      kind,
	}
	// Empty password is not supported in password authenticator
	// If the password is empty string means no password for this password authenticator
	// In this case, the authenticator cannot be used to authenticate successfully
	if password != "" {
		err := p.isPasswordAllowed(userID, password)
		if err != nil {
			return nil, err
		}

		authen = p.populatePasswordHash(authen, password)
	} else {
		authen.PasswordHash = nil
	}
	return authen, nil
}

// WithPassword return new authenticator pointer if password is changed
// Otherwise original authenticator will be returned
func (p *Provider) WithPassword(a *Authenticator, password string) (*Authenticator, error) {
	var newAuthen *Authenticator
	if password != "" {
		err := p.isPasswordAllowed(a.UserID, password)
		if err != nil {
			return nil, err
		}

		// If password is not changed, skip the logic.
		// Return original authenticator pointer
		if pwd.Compare([]byte(password), a.PasswordHash) == nil {
			return a, nil
		}

		newAuthen = p.populatePasswordHash(a, password)
	} else {
		c := *a
		c.PasswordHash = nil
		newAuthen = &c
	}

	return newAuthen, nil
}

func (p *Provider) Create(a *Authenticator) error {
	now := p.Clock.NowUTC()
	a.CreatedAt = now
	a.UpdatedAt = now

	err := p.Store.Create(a)
	if err != nil {
		return err
	}

	err = p.PasswordHistory.CreatePasswordHistory(a.UserID, a.PasswordHash, p.Clock.NowUTC())
	if err != nil {
		return err
	}

	return nil
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
	return p.PasswordChecker.ValidatePassword(ValidatePayload{
		AuthID:        userID,
		PlainPassword: password,
	})
}

func (p *Provider) UpdatePassword(a *Authenticator) error {
	now := p.Clock.NowUTC()
	a.UpdatedAt = now

	err := p.Store.UpdatePasswordHash(a)
	if err != nil {
		return err
	}

	err = p.PasswordHistory.CreatePasswordHistory(a.UserID, a.PasswordHash, p.Clock.NowUTC())
	if err != nil {
		return err
	}

	p.TaskQueue.Enqueue(&tasks.PwHousekeeperParam{
		UserID: a.UserID,
	})

	return nil
}

func sortAuthenticators(as []*Authenticator) {
	sort.Slice(as, func(i, j int) bool {
		return as[i].ID < as[j].ID
	})
}

func (p *Provider) populatePasswordHash(a *Authenticator, password string) *Authenticator {
	hash, err := pwd.Hash([]byte(password))
	if err != nil {
		panic(fmt.Errorf("password: failed to hash password: %w", err))
	}

	newAuthn := *a
	newAuthn.PasswordHash = hash

	return &newAuthn
}
