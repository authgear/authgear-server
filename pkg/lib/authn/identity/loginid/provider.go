package loginid

import (
	"errors"
	"sort"

	"github.com/iawaknahc/jsonschema/pkg/jsonpointer"

	"github.com/authgear/authgear-server/pkg/api/model"
	"github.com/authgear/authgear-server/pkg/lib/authn/identity"
	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/util/clock"
	"github.com/authgear/authgear-server/pkg/util/uuid"
)

type Provider struct {
	Store             *Store
	Config            *config.LoginIDConfig
	Checker           *Checker
	NormalizerFactory *NormalizerFactory
	Clock             clock.Clock
}

func (p *Provider) List(userID string) ([]*identity.LoginID, error) {
	is, err := p.Store.List(userID)
	if err != nil {
		return nil, err
	}

	sortIdentities(is)
	return is, nil
}

func (p *Provider) ListByClaim(name string, value string) ([]*identity.LoginID, error) {
	is, err := p.Store.ListByClaim(name, value)
	if err != nil {
		return nil, err
	}

	sortIdentities(is)
	return is, nil
}

func (p *Provider) ListByClaimJSONPointer(pointer jsonpointer.T, value string) ([]*identity.LoginID, error) {
	is, err := p.Store.ListByClaimJSONPointer(pointer, value)
	if err != nil {
		return nil, err
	}

	sortIdentities(is)
	return is, nil
}

func (p *Provider) Get(userID, id string) (*identity.LoginID, error) {
	return p.Store.Get(userID, id)
}

func (p *Provider) GetByValue(value string) ([]*identity.LoginID, error) {
	im := map[string]*identity.LoginID{}
	for _, config := range p.Config.Keys {
		// Normalize expects loginID is in correct type so we have to validate it first.
		invalid := p.Checker.ValidateOne(identity.LoginIDSpec{
			Key:   config.Key,
			Type:  config.Type,
			Value: value,
		}, CheckerOptions{
			// Admin can create email login id which bypass domains blocklist allowlist
			// it should not affect getting identity
			// skip the checking when getting identity
			EmailByPassBlocklistAllowlist: true,
		})
		if invalid != nil {
			continue
		}

		normalizer := p.NormalizerFactory.NormalizerWithLoginIDType(config.Type)
		normalizedloginID, err := normalizer.Normalize(value)
		if err != nil {
			return nil, err
		}

		i, err := p.Store.GetByLoginID(config.Key, normalizedloginID)
		if errors.Is(err, identity.ErrIdentityNotFound) {
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

func (p *Provider) GetMany(ids []string) ([]*identity.LoginID, error) {
	return p.Store.GetMany(ids)
}

func (p *Provider) CheckAndNormalize(spec identity.LoginIDSpec) (normalized string, uniqueKey string, err error) {
	err = p.Checker.ValidateOne(spec, CheckerOptions{
		// Bypass blocklist allowlist in checking and normalizing value.
		EmailByPassBlocklistAllowlist: true,
	})
	if err != nil {
		return
	}

	normalizer := p.NormalizerFactory.NormalizerWithLoginIDType(spec.Type)
	normalized, err = normalizer.Normalize(spec.Value)
	if err != nil {
		return
	}

	uniqueKey, err = normalizer.ComputeUniqueKey(normalized)
	if err != nil {
		return
	}

	return
}

func (p *Provider) normalize(typ model.LoginIDKeyType, value string) (normalized string, uniqueKey string, err error) {
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

func (p *Provider) ValidateOne(loginID identity.LoginIDSpec, options CheckerOptions) error {
	return p.Checker.ValidateOne(loginID, options)
}

func (p *Provider) New(userID string, spec identity.LoginIDSpec, options CheckerOptions) (*identity.LoginID, error) {
	err := p.ValidateOne(spec, options)
	if err != nil {
		return nil, err
	}

	normalized, uniqueKey, err := p.normalize(spec.Type, spec.Value)
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
		OriginalLoginID: spec.Value,
		Claims:          claims,
	}

	return iden, nil
}

func (p *Provider) WithValue(iden *identity.LoginID, value string, options CheckerOptions) (*identity.LoginID, error) {
	spec := identity.LoginIDSpec{
		Key:   iden.LoginIDKey,
		Type:  iden.LoginIDType,
		Value: value,
	}

	err := p.ValidateOne(spec, options)
	if err != nil {
		return nil, err
	}

	normalized, uniqueKey, err := p.normalize(spec.Type, spec.Value)
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

func (p *Provider) GetByUniqueKey(uniqueKey string) (*identity.LoginID, error) {
	return p.Store.GetByUniqueKey(uniqueKey)
}

func (p *Provider) Create(i *identity.LoginID) error {
	now := p.Clock.NowUTC()
	i.CreatedAt = now
	i.UpdatedAt = now
	return p.Store.Create(i)
}

func (p *Provider) Update(i *identity.LoginID) error {
	now := p.Clock.NowUTC()
	i.UpdatedAt = now
	return p.Store.Update(i)
}

func (p *Provider) Delete(i *identity.LoginID) error {
	return p.Store.Delete(i)
}

func sortIdentities(is []*identity.LoginID) {
	sort.Slice(is, func(i, j int) bool {
		return is[i].UniqueKey < is[j].UniqueKey
	})
}
