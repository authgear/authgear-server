package passkey

import (
	"sort"

	"github.com/authgear/authgear-server/pkg/lib/webauthn"
	"github.com/authgear/authgear-server/pkg/util/clock"
	"github.com/authgear/authgear-server/pkg/util/uuid"
)

type Provider struct {
	Store *Store
	Clock clock.Clock
}

func (p *Provider) Get(userID string, id string) (*Authenticator, error) {
	return p.Store.Get(userID, id)
}

func (p *Provider) GetMany(ids []string) ([]*Authenticator, error) {
	return p.Store.GetMany(ids)
}

func (p *Provider) Delete(a *Authenticator) error {
	return p.Store.Delete(a.ID)
}

func (p *Provider) Create(a *Authenticator) error {
	now := p.Clock.NowUTC()
	a.CreatedAt = now
	a.UpdatedAt = now
	return p.Store.Create(a)
}

func (p *Provider) Update(a *Authenticator) error {
	now := p.Clock.NowUTC()
	a.UpdatedAt = now

	err := p.Store.UpdateSignCount(a)
	if err != nil {
		return err
	}

	return nil
}

func (p *Provider) List(userID string) ([]*Authenticator, error) {
	authenticators, err := p.Store.List(userID)
	if err != nil {
		return nil, err
	}

	sortAuthenticators(authenticators)
	return authenticators, nil
}

func (p *Provider) New(
	id string,
	userID string,
	creationOptions *webauthn.CreationOptions,
	attestationResponse []byte,
	isDefault bool,
	kind string,
) (*Authenticator, error) {
	credentialID, signCount, err := webauthn.VerifyAttestationResponse(attestationResponse)
	if err != nil {
		return nil, err
	}
	if id == "" {
		id = uuid.New()
	}
	a := &Authenticator{
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

func (p *Provider) Authenticate(a *Authenticator, assertionResponse []byte) (requireUpdate bool, err error) {
	_, signCount, err := webauthn.ParseAssertionResponse(assertionResponse)
	if err != nil {
		return
	}

	if signCount != a.SignCount {
		a.SignCount = signCount
		requireUpdate = true
	}

	return
}

func sortAuthenticators(as []*Authenticator) {
	sort.Slice(as, func(i, j int) bool {
		return as[i].CreatedAt.Before(as[j].CreatedAt)
	})
}
