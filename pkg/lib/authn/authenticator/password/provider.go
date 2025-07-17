package password

import (
	"context"
	"fmt"
	"log/slog"
	"sort"
	"time"

	"github.com/authgear/authgear-server/pkg/lib/authn/authenticator"
	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/util/clock"
	pwd "github.com/authgear/authgear-server/pkg/util/password"
	"github.com/authgear/authgear-server/pkg/util/slogutil"
	"github.com/authgear/authgear-server/pkg/util/uuid"
)

var ProviderLogger = slogutil.NewLogger("password")

type Provider struct {
	Store           *Store
	Config          *config.AuthenticatorPasswordConfig
	Clock           clock.Clock
	PasswordHistory *HistoryStore
	PasswordChecker *Checker
	Expiry          *Expiry
	Housekeeper     *Housekeeper
}

func (p *Provider) Get(ctx context.Context, userID string, id string) (*authenticator.Password, error) {
	return p.Store.Get(ctx, userID, id)
}

func (p *Provider) GetMany(ctx context.Context, ids []string) ([]*authenticator.Password, error) {
	return p.Store.GetMany(ctx, ids)
}

func (p *Provider) Delete(ctx context.Context, a *authenticator.Password) error {
	return p.Store.Delete(ctx, a.ID)
}

func (p *Provider) List(ctx context.Context, userID string) ([]*authenticator.Password, error) {
	authenticators, err := p.Store.List(ctx, userID)
	if err != nil {
		return nil, err
	}

	sortAuthenticators(authenticators)
	return authenticators, nil
}

func (p *Provider) New(ctx context.Context, id string, userID string, passwordSpec *authenticator.PasswordSpec, isDefault bool, kind string) (*authenticator.Password, error) {
	if id == "" {
		id = uuid.New()
	}
	authen := &authenticator.Password{
		ID:          id,
		UserID:      userID,
		IsDefault:   isDefault,
		Kind:        kind,
		ExpireAfter: passwordSpec.ExpireAfter,
	}

	switch {
	// The input password is plain password.
	case passwordSpec.PlainPassword != "":
		err := p.PasswordChecker.ValidateNewPassword(ctx, userID, passwordSpec.PlainPassword)
		if err != nil {
			return nil, err
		}
		authen = p.populatePasswordHash(authen, passwordSpec.PlainPassword)
		return authen, nil
	// The input password is a bcrypt hash.
	case passwordSpec.PasswordHash != "":
		hash := []byte(passwordSpec.PasswordHash)
		err := pwd.CheckHash(hash)
		if err != nil {
			return nil, TranslateBcryptError(err)
		}
		authen = p.populatePasswordHashWithHash(authen, hash)
		return authen, nil
	default:
		panic(fmt.Errorf("invalid password spec"))
	}
}

type UpdatePasswordOptions struct {
	SetPassword    bool
	PlainPassword  string
	SetExpireAfter bool
	ExpireAfter    *time.Time
}

// UpdatePassword return new authenticator pointer if password or expireAfter is changed
// Otherwise original authenticator will be returned
func (p *Provider) UpdatePassword(ctx context.Context, a *authenticator.Password, options *UpdatePasswordOptions) (bool, *authenticator.Password, error) {
	password := options.PlainPassword

	newAuthen := a
	if options.SetPassword {
		err := p.PasswordChecker.ValidateNewPassword(ctx, a.UserID, password)
		if err != nil {
			return false, nil, err
		}
		if pwd.Compare([]byte(password), a.PasswordHash) != nil {
			newAuthen = p.populatePasswordHash(a, password)
		}
	}

	if options.SetExpireAfter {
		newAuthen = p.populateExpireAfter(newAuthen, options.ExpireAfter)
	}

	changed := newAuthen != a

	return changed, newAuthen, nil
}

func (p *Provider) Create(ctx context.Context, a *authenticator.Password) error {
	now := p.Clock.NowUTC()
	a.CreatedAt = now
	a.UpdatedAt = now

	err := p.Store.Create(ctx, a)
	if err != nil {
		return err
	}

	err = p.PasswordHistory.CreatePasswordHistory(ctx, a.UserID, a.PasswordHash, p.Clock.NowUTC())
	if err != nil {
		return err
	}

	return nil
}

func (p *Provider) Authenticate(ctx context.Context, a *authenticator.Password, password string) (verifyResult *VerifyResult, err error) {
	logger := ProviderLogger.GetLogger(ctx)
	verifyResult = &VerifyResult{}
	err = pwd.Compare([]byte(password), a.PasswordHash)
	if err != nil {
		return
	}

	migrated, err := pwd.TryMigrate([]byte(password), &a.PasswordHash)
	if err != nil {
		logger.WithError(err).Warn(ctx, "Failed to migrate password", slog.String("authenticator_id", a.ID))
		return
	}

	if migrated {
		err = p.Store.UpdatePasswordHash(ctx, a)
		if err != nil {
			logger.WithError(err).Warn(ctx, "Failed to save migrated password", slog.String("authenticator_id", a.ID))
			return
		}
	}

	if validateErr := p.PasswordChecker.ValidateCurrentPassword(password); validateErr != nil {
		if p.Config.ForceChange != nil && *p.Config.ForceChange {
			verifyResult.PolicyForceChange = true
		}
	}

	if expiryErr := p.Expiry.Validate(a); expiryErr != nil {
		verifyResult.ExpiryForceChange = true
	}

	return
}

func (p *Provider) Update(ctx context.Context, a *authenticator.Password) error {
	now := p.Clock.NowUTC()
	a.UpdatedAt = now

	err := p.Store.UpdatePasswordHash(ctx, a)
	if err != nil {
		return err
	}

	err = p.PasswordHistory.CreatePasswordHistory(ctx, a.UserID, a.PasswordHash, p.Clock.NowUTC())
	if err != nil {
		return err
	}

	err = p.Housekeeper.Housekeep(ctx, a.UserID)
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

func (p *Provider) populatePasswordHashWithHash(a *authenticator.Password, hash []byte) *authenticator.Password {
	newAuthn := *a
	newAuthn.PasswordHash = hash
	return &newAuthn
}

func (p *Provider) populateExpireAfter(a *authenticator.Password, expireAfter *time.Time) *authenticator.Password {
	newAuthn := *a
	newAuthn.ExpireAfter = expireAfter
	return &newAuthn
}
