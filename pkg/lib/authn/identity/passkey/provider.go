package passkey

import (
	"fmt"
	"sort"

	"github.com/authgear/authgear-server/pkg/api/model"
	"github.com/authgear/authgear-server/pkg/lib/authn/identity"
	"github.com/authgear/authgear-server/pkg/util/clock"
	"github.com/authgear/authgear-server/pkg/util/uuid"
)

// nolint: golint
type PasskeyService interface {
	PeekAttestationResponse(attestationResponse []byte) (creationOptions *model.WebAuthnCreationOptions, credentialID string, signCount int64, err error)
	GetCredentialIDFromAssertionResponse(assertionResponse []byte) (credentialID string, err error)
}

type Provider struct {
	Store   *Store
	Clock   clock.Clock
	Passkey PasskeyService
}

func (p *Provider) List(userID string) ([]*identity.Passkey, error) {
	is, err := p.Store.List(userID)
	if err != nil {
		return nil, err
	}

	sortIdentities(is)
	return is, nil
}

func (p *Provider) Get(userID, id string) (*identity.Passkey, error) {
	return p.Store.Get(userID, id)
}

func (p *Provider) GetBySpec(spec *identity.PasskeySpec) (*identity.Passkey, error) {
	switch {
	case spec.AttestationResponse != nil:
		_, credentialID, _, err := p.Passkey.PeekAttestationResponse(spec.AttestationResponse)
		if err != nil {
			return nil, err
		}
		return p.Store.GetByCredentialID(credentialID)
	case spec.AssertionResponse != nil:
		credentialID, err := p.Passkey.GetCredentialIDFromAssertionResponse(spec.AssertionResponse)
		if err != nil {
			return nil, err
		}
		return p.Store.GetByCredentialID(credentialID)
	default:
		panic(fmt.Errorf("passkey: expect either attestation response or assert response in passkey spec"))
	}
}

func (p *Provider) GetMany(ids []string) ([]*identity.Passkey, error) {
	return p.Store.GetMany(ids)
}

func (p *Provider) New(
	userID string,
	attestationResponse []byte,
) (*identity.Passkey, error) {
	creationOptions, credentialID, _, err := p.Passkey.PeekAttestationResponse(attestationResponse)
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

func (p *Provider) Create(i *identity.Passkey) error {
	now := p.Clock.NowUTC()
	i.CreatedAt = now
	i.UpdatedAt = now
	return p.Store.Create(i)
}

func (p *Provider) Delete(i *identity.Passkey) error {
	return p.Store.Delete(i)
}

func sortIdentities(is []*identity.Passkey) {
	sort.Slice(is, func(i, j int) bool {
		return is[i].CreatedAt.Before(is[j].CreatedAt)
	})
}
