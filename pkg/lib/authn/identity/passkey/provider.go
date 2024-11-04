package passkey

import (
	"context"
	"fmt"
	"sort"

	"github.com/authgear/authgear-server/pkg/api/model"
	"github.com/authgear/authgear-server/pkg/lib/authn/identity"
	"github.com/authgear/authgear-server/pkg/util/clock"
	"github.com/authgear/authgear-server/pkg/util/uuid"
)

// nolint: golint
type PasskeyService interface {
	PeekAttestationResponse(ctx context.Context, attestationResponse []byte) (creationOptions *model.WebAuthnCreationOptions, credentialID string, signCount int64, err error)
	GetCredentialIDFromAssertionResponse(ctx context.Context, assertionResponse []byte) (credentialID string, err error)
}

type Provider struct {
	Store   *Store
	Clock   clock.Clock
	Passkey PasskeyService
}

func (p *Provider) List(ctx context.Context, userID string) ([]*identity.Passkey, error) {
	is, err := p.Store.List(ctx, userID)
	if err != nil {
		return nil, err
	}

	sortIdentities(is)
	return is, nil
}

func (p *Provider) Get(ctx context.Context, userID, id string) (*identity.Passkey, error) {
	return p.Store.Get(ctx, userID, id)
}

func (p *Provider) GetBySpec(ctx context.Context, spec *identity.PasskeySpec) (*identity.Passkey, error) {
	switch {
	case spec.AttestationResponse != nil:
		_, credentialID, _, err := p.Passkey.PeekAttestationResponse(ctx, spec.AttestationResponse)
		if err != nil {
			return nil, err
		}
		return p.Store.GetByCredentialID(ctx, credentialID)
	case spec.AssertionResponse != nil:
		credentialID, err := p.Passkey.GetCredentialIDFromAssertionResponse(ctx, spec.AssertionResponse)
		if err != nil {
			return nil, err
		}
		return p.Store.GetByCredentialID(ctx, credentialID)
	default:
		panic(fmt.Errorf("passkey: expect either attestation response or assert response in passkey spec"))
	}
}

func (p *Provider) GetMany(ctx context.Context, ids []string) ([]*identity.Passkey, error) {
	return p.Store.GetMany(ctx, ids)
}

func (p *Provider) New(
	ctx context.Context,
	userID string,
	attestationResponse []byte,
) (*identity.Passkey, error) {
	creationOptions, credentialID, _, err := p.Passkey.PeekAttestationResponse(ctx, attestationResponse)
	if err != nil {
		return nil, err
	}

	i := &identity.Passkey{
		ID:                  uuid.New(),
		UserID:              userID,
		CredentialID:        credentialID,
		CreationOptions:     creationOptions,
		AttestationResponse: attestationResponse,
	}
	return i, nil
}

func (p *Provider) Create(ctx context.Context, i *identity.Passkey) error {
	now := p.Clock.NowUTC()
	i.CreatedAt = now
	i.UpdatedAt = now
	return p.Store.Create(ctx, i)
}

func (p *Provider) Delete(ctx context.Context, i *identity.Passkey) error {
	return p.Store.Delete(ctx, i)
}

func sortIdentities(is []*identity.Passkey) {
	sort.Slice(is, func(i, j int) bool {
		return is[i].CreatedAt.Before(is[j].CreatedAt)
	})
}
