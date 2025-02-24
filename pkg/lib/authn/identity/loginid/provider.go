package loginid

import (
	"context"
	"errors"
	"sort"

	"github.com/authgear/authgear-server/pkg/api"
	"github.com/authgear/authgear-server/pkg/api/model"
	"github.com/authgear/authgear-server/pkg/lib/authn/identity"
	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/util/clock"
	"github.com/authgear/authgear-server/pkg/util/stringutil"
	"github.com/authgear/authgear-server/pkg/util/uuid"
)

type Provider struct {
	Store             *Store
	Config            *config.LoginIDConfig
	Checker           *Checker
	NormalizerFactory *NormalizerFactory
	Clock             clock.Clock
}

func (p *Provider) List(ctx context.Context, userID string) ([]*identity.LoginID, error) {
	is, err := p.Store.List(ctx, userID)
	if err != nil {
		return nil, err
	}

	sortIdentities(is)
	return is, nil
}

func (p *Provider) ListByClaim(ctx context.Context, name string, value string) ([]*identity.LoginID, error) {
	is, err := p.Store.ListByClaim(ctx, name, value)
	if err != nil {
		return nil, err
	}

	sortIdentities(is)
	return is, nil
}

func (p *Provider) Get(ctx context.Context, userID, id string) (*identity.LoginID, error) {
	return p.Store.Get(ctx, userID, id)
}

func (p *Provider) GetByValue(ctx context.Context, value string) ([]*identity.LoginID, error) {
	im := map[string]*identity.LoginID{}
	for _, config := range p.Config.Keys {
		// Normalize expects loginID is in correct type so we have to validate it first.
		spec := identity.LoginIDSpec{
			Key:   config.Key,
			Type:  config.Type,
			Value: stringutil.NewUserInputString(value),
		}
		invalid := p.Checker.ValidateOne(ctx, spec, CheckerOptions{
			// Admin can create email login id which bypass domains blocklist allowlist
			// it should not affect getting identity
			// skip the checking when getting identity
			EmailByPassBlocklistAllowlist: true,
		})
		if invalid != nil {
			continue
		}

		normalizer := p.NormalizerFactory.NormalizerWithLoginIDType(config.Type)
		normalizedloginID, err := normalizer.Normalize(spec.Value.TrimSpace())
		if err != nil {
			return nil, err
		}
		uniqueKey, err := normalizer.ComputeUniqueKey(normalizedloginID)
		if err != nil {
			return nil, err
		}

		i, err := p.Store.GetByUniqueKey(ctx, uniqueKey)
		if errors.Is(err, api.ErrIdentityNotFound) {
			continue
		} else if err != nil {
			return nil, err
		}

		im[i.ID] = i
	}

	var is []*identity.LoginID
	for _, i := range im {
		is = append(is, i)
	}
	return is, nil
}

func (p *Provider) GetByKeyAndValue(ctx context.Context, key string, value string) (*identity.LoginID, error) {
	cfg, ok := p.Config.GetKeyConfig(key)

	if !ok {
		return nil, api.ErrGetUsersInvalidArgument.New("invalid Login ID key")
	}

	normalizer := p.NormalizerFactory.NormalizerWithLoginIDType(cfg.Type)
	normalizedloginID, err := normalizer.Normalize(value)
	if err != nil {
		return nil, api.ErrGetUsersInvalidArgument.New("invalid Login ID value")
	}
	uniqueKey, err := normalizer.ComputeUniqueKey(normalizedloginID)
	if err != nil {
		return nil, err
	}

	i, err := p.Store.GetByUniqueKey(ctx, uniqueKey)

	if err != nil {
		return nil, err
	}

	return i, nil
}

func (p *Provider) GetMany(ctx context.Context, ids []string) ([]*identity.LoginID, error) {
	return p.Store.GetMany(ctx, ids)
}

func (p *Provider) CheckAndNormalize(ctx context.Context, spec identity.LoginIDSpec) (normalized string, uniqueKey string, err error) {
	err = p.Checker.ValidateOne(ctx, spec, CheckerOptions{
		// Bypass blocklist allowlist in checking and normalizing value.
		EmailByPassBlocklistAllowlist: true,
	})
	if err != nil {
		return
	}

	normalizer := p.NormalizerFactory.NormalizerWithLoginIDType(spec.Type)
	normalized, err = normalizer.Normalize(spec.Value.TrimSpace())
	if err != nil {
		return
	}

	uniqueKey, err = normalizer.ComputeUniqueKey(normalized)
	if err != nil {
		return
	}

	return
}

func (p *Provider) Normalize(typ model.LoginIDKeyType, value string) (normalized string, uniqueKey string, err error) {
	normalizer := p.NormalizerFactory.NormalizerWithLoginIDType(typ)
	normalized, err = normalizer.Normalize(value)
	if err != nil {
		return
	}

	uniqueKey, err = normalizer.ComputeUniqueKey(normalized)
	if err != nil {
		return
	}

	return
}

func (p *Provider) validateOne(ctx context.Context, loginID identity.LoginIDSpec, options CheckerOptions) error {
	return p.Checker.ValidateOne(ctx, loginID, options)
}

func (p *Provider) New(ctx context.Context, userID string, spec identity.LoginIDSpec, options CheckerOptions) (*identity.LoginID, error) {
	err := p.validateOne(ctx, spec, options)
	if err != nil {
		return nil, err
	}

	normalized, uniqueKey, err := p.Normalize(spec.Type, spec.Value.TrimSpace())
	if err != nil {
		return nil, err
	}

	claims := make(map[string]interface{})
	if claimName, ok := model.GetLoginIDKeyTypeClaim(spec.Type); ok {
		claims[string(claimName)] = normalized
	}

	iden := &identity.LoginID{
		ID:              uuid.New(),
		UserID:          userID,
		LoginIDKey:      spec.Key,
		LoginIDType:     spec.Type,
		LoginID:         normalized,
		UniqueKey:       uniqueKey,
		OriginalLoginID: spec.Value.TrimSpace(),
		Claims:          claims,
	}

	return iden, nil
}

func (p *Provider) WithValue(ctx context.Context, iden *identity.LoginID, value string, options CheckerOptions) (*identity.LoginID, error) {
	spec := identity.LoginIDSpec{
		Key:   iden.LoginIDKey,
		Type:  iden.LoginIDType,
		Value: stringutil.NewUserInputString(value),
	}

	err := p.validateOne(ctx, spec, options)
	if err != nil {
		return nil, err
	}

	normalized, uniqueKey, err := p.Normalize(spec.Type, spec.Value.TrimSpace())
	if err != nil {
		return nil, err
	}

	claims := make(map[string]interface{})
	if claimName, ok := model.GetLoginIDKeyTypeClaim(spec.Type); ok {
		claims[string(claimName)] = normalized
	}

	newIden := *iden
	newIden.LoginID = normalized
	newIden.UniqueKey = uniqueKey
	newIden.OriginalLoginID = value
	newIden.Claims = claims

	return &newIden, nil
}

func (p *Provider) GetByUniqueKey(ctx context.Context, uniqueKey string) (*identity.LoginID, error) {
	return p.Store.GetByUniqueKey(ctx, uniqueKey)
}

func (p *Provider) Create(ctx context.Context, i *identity.LoginID) error {
	now := p.Clock.NowUTC()
	i.CreatedAt = now
	i.UpdatedAt = now
	return p.Store.Create(ctx, i)
}

func (p *Provider) Update(ctx context.Context, i *identity.LoginID) error {
	now := p.Clock.NowUTC()
	i.UpdatedAt = now
	return p.Store.Update(ctx, i)
}

func (p *Provider) Delete(ctx context.Context, i *identity.LoginID) error {
	return p.Store.Delete(ctx, i)
}

func sortIdentities(is []*identity.LoginID) {
	sort.Slice(is, func(i, j int) bool {
		return is[i].UniqueKey < is[j].UniqueKey
	})
}
