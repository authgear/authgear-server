package passkey

import (
	"sort"

	"github.com/authgear/authgear-server/pkg/api/model"
	"github.com/authgear/authgear-server/pkg/lib/authn/authenticator"
	"github.com/authgear/authgear-server/pkg/util/clock"
	"github.com/authgear/authgear-server/pkg/util/uuid"
)

// nolint: golint
type PasskeyService interface {
	VerifyAttestationResponse(attestationResponse []byte) (credentialID string, signCount int64, err error)
	ParseAssertionResponse(assertionResponse []byte) (credentialID string, signCount int64, err error)
}

type Provider struct {
	Store   *Store
	Clock   clock.Clock
	Passkey PasskeyService
}

func (p *Provider) Get(userID string, id string) (*authenticator.Passkey, error) {
	return p.Store.Get(userID, id)
}

func (p *Provider) GetMany(ids []string) ([]*authenticator.Passkey, error) {
	return p.Store.GetMany(ids)
}

func (p *Provider) Delete(a *authenticator.Passkey) error {
	return p.Store.Delete(a.ID)
}

func (p *Provider) Create(a *authenticator.Passkey) error {
	now := p.Clock.NowUTC()
	a.CreatedAt = now
	a.UpdatedAt = now
	return p.Store.Create(a)
}

func (p *Provider) Update(a *authenticator.Passkey) error {
	now := p.Clock.NowUTC()
	a.UpdatedAt = now

	err := p.Store.UpdateSignCount(a)
	if err != nil {
		return err
	}

	return nil
}

func (p *Provider) List(userID string) ([]*authenticator.Passkey, error) {
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
	creationOptions *model.WebAuthnCreationOptions,
	attestationResponse []byte,
	isDefault bool,
	kind string,
) (*authenticator.Passkey, error) {
	credentialID, signCount, err := p.Passkey.VerifyAttestationResponse(attestationResponse)
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

func (p *Provider) Authenticate(a *authenticator.Passkey, assertionResponse []byte) (requireUpdate bool, err error) {
	_, signCount, err := p.Passkey.ParseAssertionResponse(assertionResponse)
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
