package oob

import (
	"context"
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

func (p *Provider) Get(ctx context.Context, userID string, id string) (*authenticator.OOBOTP, error) {
	return p.Store.Get(ctx, userID, id)
}

func (p *Provider) GetMany(ctx context.Context, ids []string) ([]*authenticator.OOBOTP, error) {
	return p.Store.GetMany(ctx, ids)
}

func (p *Provider) Delete(ctx context.Context, a *authenticator.OOBOTP) error {
	return p.Store.Delete(ctx, a.ID)
}

func (p *Provider) List(ctx context.Context, userID string) ([]*authenticator.OOBOTP, error) {
	authenticators, err := p.Store.List(ctx, userID)
	if err != nil {
		return nil, err
	}

	sortAuthenticators(authenticators)
	return authenticators, nil
}

func (p *Provider) New(ctx context.Context, id string, userID string, oobAuthenticatorType model.AuthenticatorType, target string, isDefault bool, kind string) (*authenticator.OOBOTP, error) {
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
		err := validation.FormatEmail{AllowName: false}.CheckFormat(ctx, target)
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
		err := validation.FormatPhone{}.CheckFormat(ctx, target)
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
	NewTarget string
}

// UpdateTarget return new authenticator pointer if target is changed
// Otherwise original authenticator will be returned
func (p *Provider) UpdateTarget(a *authenticator.OOBOTP, option UpdateTargetOption) (*authenticator.OOBOTP, bool) {
	switch a.OOBAuthenticatorType {
	case model.AuthenticatorTypeOOBEmail:
		if a.ToTarget() == option.NewTarget {
			return a, false
		}
		newAuth := *a
		newAuth.Email = option.NewTarget
		return &newAuth, true
	case model.AuthenticatorTypeOOBSMS:
		if a.ToTarget() == option.NewTarget {
			return a, false
		}
		newAuth := *a
		newAuth.Phone = option.NewTarget
		return &newAuth, true
	default:
		panic("oob: incompatible authenticator type:" + a.OOBAuthenticatorType)
	}
}

func (p *Provider) Create(ctx context.Context, a *authenticator.OOBOTP) error {
	now := p.Clock.NowUTC()
	a.CreatedAt = now
	a.UpdatedAt = now
	return p.Store.Create(ctx, a)
}

func (p *Provider) Update(ctx context.Context, a *authenticator.OOBOTP) error {
	now := p.Clock.NowUTC()
	a.UpdatedAt = now
	return p.Store.Update(ctx, a)
}

func sortAuthenticators(as []*authenticator.OOBOTP) {
	sort.Slice(as, func(i, j int) bool {
		return as[i].CreatedAt.Before(as[j].CreatedAt)
	})
}
