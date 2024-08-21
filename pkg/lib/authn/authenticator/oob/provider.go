package oob

import (
	"fmt"
	"sort"

	"github.com/authgear/authgear-server/pkg/api/internalinterface"
	"github.com/authgear/authgear-server/pkg/api/model"
	"github.com/authgear/authgear-server/pkg/lib/authn/authenticator"
	"github.com/authgear/authgear-server/pkg/util/clock"
	"github.com/authgear/authgear-server/pkg/util/uuid"
	"github.com/authgear/authgear-server/pkg/util/validation"
)

type LoginIDNormalizerFactory interface {
	NormalizerWithLoginIDType(loginIDKeyType model.LoginIDKeyType) internalinterface.LoginIDNormalizer
}

type Provider struct {
	Store                    *Store
	LoginIDNormalizerFactory LoginIDNormalizerFactory
	Clock                    clock.Clock
}

func (p *Provider) Get(userID string, id string) (*authenticator.OOBOTP, error) {
	return p.Store.Get(userID, id)
}

func (p *Provider) GetMany(ids []string) ([]*authenticator.OOBOTP, error) {
	return p.Store.GetMany(ids)
}

func (p *Provider) Delete(a *authenticator.OOBOTP) error {
	return p.Store.Delete(a.ID)
}

func (p *Provider) List(userID string) ([]*authenticator.OOBOTP, error) {
	authenticators, err := p.Store.List(userID)
	if err != nil {
		return nil, err
	}

	sortAuthenticators(authenticators)
	return authenticators, nil
}

func (p *Provider) New(id string, userID string, oobAuthenticatorType model.AuthenticatorType, target string, isDefault bool, kind string) (*authenticator.OOBOTP, error) {
	if id == "" {
		id = uuid.New()
	}
	a := &authenticator.OOBOTP{
		ID:                   id,
		UserID:               userID,
		OOBAuthenticatorType: oobAuthenticatorType,
		IsDefault:            isDefault,
		Kind:                 kind,
	}

	// Validate and normalize the target.
	switch oobAuthenticatorType {
	case model.AuthenticatorTypeOOBEmail:
		validationCtx := &validation.Context{}
		err := validation.FormatEmail{AllowName: false}.CheckFormat(target)
		if err != nil {
			validationCtx.EmitError("format", map[string]interface{}{"format": "email"})
		}
		err = validationCtx.Error("invalid target")
		if err != nil {
			return nil, err
		}

		target, err = p.LoginIDNormalizerFactory.NormalizerWithLoginIDType(model.LoginIDKeyTypeEmail).Normalize(target)
		if err != nil {
			return nil, err
		}

		a.Email = target
	case model.AuthenticatorTypeOOBSMS:
		validationCtx := &validation.Context{}
		err := validation.FormatPhone{}.CheckFormat(target)
		if err != nil {
			validationCtx.EmitError("format", map[string]interface{}{"format": "phone"})
		}
		err = validationCtx.Error("invalid target")
		if err != nil {
			return nil, err
		}

		target, err = p.LoginIDNormalizerFactory.NormalizerWithLoginIDType(model.LoginIDKeyTypePhone).Normalize(target)
		if err != nil {
			return nil, err
		}

		a.Phone = target
	default:
		panic(fmt.Errorf("authenticator: unexpected OOBOTP authenticator type: %v", oobAuthenticatorType))
	}

	return a, nil
}

type UpdateTargetOption struct {
	Email string
	Phone string
}

// UpdateTarget return new authenticator pointer if target is changed
// Otherwise original authenticator will be returned
func (p *Provider) UpdateTarget(a *authenticator.OOBOTP, option UpdateTargetOption) (bool, *authenticator.OOBOTP, error) {
	switch a.OOBAuthenticatorType {
	case model.AuthenticatorTypeOOBEmail:
		if a.ToTarget() == option.Email {
			return false, a, nil
		}
		newAuth := *a
		newAuth.Email = option.Email
		return true, &newAuth, nil
	case model.AuthenticatorTypeOOBSMS:
		if a.ToTarget() == option.Phone {
			return false, a, nil
		}
		newAuth := *a
		newAuth.Phone = option.Phone
		return true, &newAuth, nil
	default:
		panic("oob: incompatible authenticator type:" + a.OOBAuthenticatorType)
	}
}

func (p *Provider) Create(a *authenticator.OOBOTP) error {
	now := p.Clock.NowUTC()
	a.CreatedAt = now
	a.UpdatedAt = now
	return p.Store.Create(a)
}

func (p *Provider) Update(a *authenticator.OOBOTP) error {
	now := p.Clock.NowUTC()
	a.UpdatedAt = now
	return p.Store.Update(a)
}

func sortAuthenticators(as []*authenticator.OOBOTP) {
	sort.Slice(as, func(i, j int) bool {
		return as[i].CreatedAt.Before(as[j].CreatedAt)
	})
}
