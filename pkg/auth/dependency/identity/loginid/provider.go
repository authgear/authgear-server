package loginid

import (
	"sort"

	"github.com/skygeario/skygear-server/pkg/auth/config"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/identity"
	"github.com/skygeario/skygear-server/pkg/core/auth/metadata"
	"github.com/skygeario/skygear-server/pkg/core/errors"
	"github.com/skygeario/skygear-server/pkg/core/uuid"
)

type Provider struct {
	Store             *Store
	Config            *config.LoginIDConfig
	Checker           *Checker
	NormalizerFactory *NormalizerFactory
}

func (p *Provider) List(userID string) ([]*Identity, error) {
	is, err := p.Store.List(userID)
	if err != nil {
		return nil, err
	}

	sortIdentities(is)
	return is, nil
}

func (p *Provider) ListByClaim(name string, value string) ([]*Identity, error) {
	is, err := p.Store.ListByClaim(name, value)
	if err != nil {
		return nil, err
	}

	sortIdentities(is)
	return is, nil
}

func (p *Provider) Get(userID, id string) (*Identity, error) {
	return p.Store.Get(userID, id)
}

func (p *Provider) GetByLoginID(loginID LoginID) ([]*Identity, error) {
	im := map[string]*Identity{}
	for _, config := range p.Config.Keys {
		if !(loginID.Key == "" || config.Key == loginID.Key) {
			continue
		}

		// Normalize expects loginID is in correct type so we have to validate it first.
		invalid := p.Checker.ValidateOne(LoginID{
			Key:   config.Key,
			Value: loginID.Value,
		})
		if invalid != nil {
			continue
		}

		normalizer := p.NormalizerFactory.NormalizerWithLoginIDKey(config.Key)
		normalizedloginID, err := normalizer.Normalize(loginID.Value)
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

	var is []*Identity
	for _, i := range im {
		is = append(is, i)
	}
	return is, nil
}

func (p *Provider) IsLoginIDKeyType(loginIDKey string, loginIDKeyType metadata.StandardKey) bool {
	return p.Checker.CheckType(loginIDKey, loginIDKeyType)
}

func (p *Provider) Normalize(loginID LoginID) (normalized *LoginID, c *config.LoginIDKeyConfig, uniqueKey string, err error) {
	err = p.Validate([]LoginID{loginID})
	if err != nil {
		return
	}

	c = p.lookupLoginIDConfig(loginID)
	if c == nil {
		err = errors.Newf("loginid: unknown login ID key %s", loginID.Key)
		return
	}

	normalizer := p.NormalizerFactory.NormalizerWithLoginIDKey(loginID.Key)
	normalizedloginID, err := normalizer.Normalize(loginID.Value)
	if err != nil {
		return
	}

	normalized = &LoginID{
		Key:   loginID.Key,
		Value: normalizedloginID,
	}

	uniqueKey, err = normalizer.ComputeUniqueKey(normalizedloginID)
	if err != nil {
		return
	}

	return
}

func (p *Provider) Validate(loginIDs []LoginID) error {
	return p.Checker.Validate(loginIDs)
}

func (p *Provider) New(userID string, loginID LoginID) (*Identity, error) {
	iden := &Identity{
		ID:     uuid.New(),
		UserID: userID,
	}

	iden, err := p.populateLoginID(iden, loginID)
	if err != nil {
		return nil, err
	}

	return iden, nil
}

func (p *Provider) WithLoginID(iden *Identity, loginID LoginID) (*Identity, error) {
	newIden, err := p.populateLoginID(iden, loginID)
	if err != nil {
		return nil, err
	}
	return newIden, nil
}

func (p *Provider) CheckDuplicated(uniqueKey string, standardClaims map[string]string, userID string) error {
	// check duplication with unique key
	_, err := p.Store.GetByUniqueKey(uniqueKey)
	if err == nil {
		return identity.ErrIdentityAlreadyExists
	} else if !errors.Is(err, identity.ErrIdentityNotFound) {
		return err
	}

	// check duplication with standard claims
	for name, value := range standardClaims {
		ls, err := p.ListByClaim(name, value)
		if err != nil {
			return err
		}

		for _, i := range ls {
			if i.UserID == userID {
				continue
			}
			return identity.ErrIdentityAlreadyExists
		}
	}

	return nil
}

func (p *Provider) Create(i *Identity) error {
	return p.Store.Create(i)
}

func (p *Provider) Update(i *Identity) error {
	return p.Store.Update(i)
}

func (p *Provider) Delete(i *Identity) error {
	return p.Store.Delete(i)
}

func (p *Provider) lookupLoginIDConfig(loginID LoginID) *config.LoginIDKeyConfig {
	for _, c := range p.Config.Keys {
		if c.Key == loginID.Key {
			return &c
		}
	}
	return nil
}

func sortIdentities(is []*Identity) {
	sort.Slice(is, func(i, j int) bool {
		return is[i].UniqueKey < is[j].UniqueKey
	})
}

func (p *Provider) populateLoginID(i *Identity, loginID LoginID) (newIden *Identity, err error) {
	normalized, _, uniqueKey, err := p.Normalize(loginID)
	if err != nil {
		return
	}

	claims := map[string]string{}
	if standardKey, ok := p.Checker.StandardKey(loginID.Key); ok {
		claims[string(standardKey)] = normalized.Value
	}

	copyI := *i
	newIden = &copyI
	newIden.LoginIDKey = loginID.Key
	newIden.OriginalLoginID = loginID.Value
	newIden.LoginID = normalized.Value
	newIden.UniqueKey = uniqueKey
	newIden.Claims = claims

	return
}
