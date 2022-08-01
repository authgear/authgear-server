package password

import (
	"fmt"
	"sort"

	"github.com/authgear/authgear-server/pkg/lib/authn/authenticator"
	"github.com/authgear/authgear-server/pkg/lib/config"
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
	Housekeeper     *Housekeeper
}

func (p *Provider) Get(userID string, id string) (*authenticator.Password, error) {
	return p.Store.Get(userID, id)
}

func (p *Provider) GetMany(ids []string) ([]*authenticator.Password, error) {
	return p.Store.GetMany(ids)
}

func (p *Provider) Delete(a *authenticator.Password) error {
	return p.Store.Delete(a.ID)
}

func (p *Provider) List(userID string) ([]*authenticator.Password, error) {
	authenticators, err := p.Store.List(userID)
	if err != nil {
		return nil, err
	}

	sortAuthenticators(authenticators)
	return authenticators, nil
}

func (p *Provider) New(id string, userID string, password string, isDefault bool, kind string) (*authenticator.Password, error) {
	if id == "" {
		id = uuid.New()
	}
	authen := &authenticator.Password{
		ID:        id,
		UserID:    userID,
		IsDefault: isDefault,
		Kind:      kind,
	}
	err := p.PasswordChecker.ValidateNewPassword(userID, password)
	if err != nil {
		return nil, err
	}
	authen = p.populatePasswordHash(authen, password)
	return authen, nil
}

// WithPassword return new authenticator pointer if password is changed
// Otherwise original authenticator will be returned
func (p *Provider) WithPassword(a *authenticator.Password, password string) (*authenticator.Password, error) {
	err := p.PasswordChecker.ValidateNewPassword(a.UserID, password)
	if err != nil {
		return nil, err
	}

	// If password is not changed, skip the logic.
	// Return original authenticator pointer
	if pwd.Compare([]byte(password), a.PasswordHash) == nil {
		return a, nil
	}

	newAuthen := p.populatePasswordHash(a, password)
	return newAuthen, nil
}

func (p *Provider) Create(a *authenticator.Password) error {
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

func (p *Provider) Authenticate(a *authenticator.Password, password string) (requireUpdate bool, err error) {
	err = pwd.Compare([]byte(password), a.PasswordHash)
	if err != nil {
		return
	}

	migrated, err := pwd.TryMigrate([]byte(password), &a.PasswordHash)
	if err != nil {
		p.Logger.WithError(err).WithField("authenticator_id", a.ID).
			Warn("Failed to migrate password")
		return
	}

	if migrated {
		err = p.Store.UpdatePasswordHash(a)
		if err != nil {
			p.Logger.WithError(err).WithField("authenticator_id", a.ID).
				Warn("Failed to save migrated password")
			return
		}
	}

	if notAllowedErr := p.PasswordChecker.ValidateCurrentPassword(password); notAllowedErr != nil {
		if p.Config.ForceChange != nil && *p.Config.ForceChange {
			requireUpdate = true
		}
	}

	return
}

func (p *Provider) UpdatePassword(a *authenticator.Password) error {
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

	err = p.Housekeeper.Housekeep(a.UserID)
	if err != nil {
		return err
	}

	return nil
}

func sortAuthenticators(as []*authenticator.Password) {
	sort.Slice(as, func(i, j int) bool {
		return as[i].ID < as[j].ID
	})
}

func (p *Provider) populatePasswordHash(a *authenticator.Password, password string) *authenticator.Password {
	hash, err := pwd.Hash([]byte(password))
	if err != nil {
		panic(fmt.Errorf("password: failed to hash password: %w", err))
	}

	newAuthn := *a
	newAuthn.PasswordHash = hash

	return &newAuthn
}
