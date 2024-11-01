package passkey

import (
	"context"
	"sort"

	"github.com/authgear/authgear-server/pkg/api/model"
	"github.com/authgear/authgear-server/pkg/lib/authn/authenticator"
	"github.com/authgear/authgear-server/pkg/util/clock"
	"github.com/authgear/authgear-server/pkg/util/uuid"
)

// nolint: golint
type PasskeyService interface {
	PeekAttestationResponse(ctx context.Context, attestationResponse []byte) (creationOptions *model.WebAuthnCreationOptions, credentialID string, signCount int64, err error)
	PeekAssertionResponse(ctx context.Context, assertionResponse []byte, attestationResponse []byte) (signCount int64, err error)
}

type Provider struct {
	Store   *Store
	Clock   clock.Clock
	Passkey PasskeyService
}

func (p *Provider) Get(ctx context.Context, userID string, id string) (*authenticator.Passkey, error) {
	return p.Store.Get(ctx, userID, id)
}

func (p *Provider) GetMany(ctx context.Context, ids []string) ([]*authenticator.Passkey, error) {
	return p.Store.GetMany(ctx, ids)
}

func (p *Provider) Delete(ctx context.Context, a *authenticator.Passkey) error {
	return p.Store.Delete(ctx, a.ID)
}

func (p *Provider) Create(ctx context.Context, a *authenticator.Passkey) error {
	now := p.Clock.NowUTC()
	a.CreatedAt = now
	a.UpdatedAt = now
	return p.Store.Create(ctx, a)
}

func (p *Provider) Update(ctx context.Context, a *authenticator.Passkey) error {
	now := p.Clock.NowUTC()
	a.UpdatedAt = now

	err := p.Store.UpdateSignCount(ctx, a)
	if err != nil {
		return err
	}

	return nil
}

func (p *Provider) List(ctx context.Context, userID string) ([]*authenticator.Passkey, error) {
	authenticators, err := p.Store.List(ctx, userID)
	if err != nil {
		return nil, err
	}

	sortAuthenticators(authenticators)
	return authenticators, nil
}

func (p *Provider) New(
	ctx context.Context,
	id string,
	userID string,
	attestationResponse []byte,
	isDefault bool,
	kind string,
) (*authenticator.Passkey, error) {
	creationOptions, credentialID, signCount, err := p.Passkey.PeekAttestationResponse(ctx, attestationResponse)
	if err != nil {
		return nil, err
	}
	if id == "" {
		id = uuid.New()
	}
	a := &authenticator.Passkey{
		ID:                  id,
		UserID:              userID,
		IsDefault:           isDefault,
		Kind:                kind,
		CredentialID:        credentialID,
		CreationOptions:     creationOptions,
		AttestationResponse: attestationResponse,
		SignCount:           signCount,
	}
	return a, nil
}

func (p *Provider) Authenticate(ctx context.Context, a *authenticator.Passkey, assertionResponse []byte) (requireUpdate bool, err error) {
	signCount, err := p.Passkey.PeekAssertionResponse(ctx, assertionResponse, a.AttestationResponse)
	if err != nil {
		return
	}

	if signCount != a.SignCount {
		a.SignCount = signCount
		requireUpdate = true
	}

	return
}

func sortAuthenticators(as []*authenticator.Passkey) {
	sort.Slice(as, func(i, j int) bool {
		return as[i].CreatedAt.Before(as[j].CreatedAt)
	})
}
