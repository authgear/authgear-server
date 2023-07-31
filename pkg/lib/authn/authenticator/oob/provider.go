package oob

import (
	"fmt"
	"sort"

	"github.com/authgear/authgear-server/pkg/api/model"
	"github.com/authgear/authgear-server/pkg/lib/authn/authenticator"
	"github.com/authgear/authgear-server/pkg/util/clock"
	"github.com/authgear/authgear-server/pkg/util/uuid"
	"github.com/authgear/authgear-server/pkg/util/validation"
)

type Provider struct {
	Store *Store
	Clock clock.Clock
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

	// Validate the target.
	validationCtx := &validation.Context{}
	switch oobAuthenticatorType {
	case model.AuthenticatorTypeOOBEmail:
		err := validation.FormatEmail{AllowName: false}.CheckFormat(target)
		if err != nil {
			validationCtx.EmitError("format", map[string]interface{}{"format": "email"})
		}

		a.Email = target
	case model.AuthenticatorTypeOOBSMS:
		err := validation.FormatPhone{}.CheckFormat(target)
		if err != nil {
			validationCtx.EmitError("format", map[string]interface{}{"format": "phone"})
		}

		a.Phone = target
	default:
		panic(fmt.Errorf("authenticator: unexpected OOBOTP authenticator type: %v", oobAuthenticatorType))
	}
	err := validationCtx.Error("invalid target")
	if err != nil {
		return nil, err
	}

	return a, nil
}

// WithSpec return new authenticator pointer if target is changed
// Otherwise original authenticator will be returned
func (p *Provider) WithSpec(a *authenticator.OOBOTP, spec *authenticator.OOBOTPSpec) (*authenticator.OOBOTP, error) {
	switch a.OOBAuthenticatorType {
	case model.AuthenticatorTypeOOBEmail:
		if spec.Email == a.ToTarget() {
			return a, nil
		}
		newAuth := *a
		newAuth.Email = spec.Email
		return &newAuth, nil
	case model.AuthenticatorTypeOOBSMS:
		if spec.Phone == a.ToTarget() {
			return a, nil
		}
		newAuth := *a
		newAuth.Phone = spec.Phone
		return &newAuth, nil
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
