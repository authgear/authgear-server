package siwe

import (
	"sort"

	"github.com/authgear/authgear-server/pkg/api/model"
	"github.com/authgear/authgear-server/pkg/lib/authn/identity"
	"github.com/authgear/authgear-server/pkg/util/clock"
	"github.com/authgear/authgear-server/pkg/util/uuid"
	siwego "github.com/spruceid/siwe-go"
)

// nolint: golint
type SIWEService interface {
	VerifyMessage(request model.SIWEVerificationRequest) (*siwego.Message, string, error)
}

type Provider struct {
	Store *Store
	Clock clock.Clock
	SIWE  SIWEService
}

func (p *Provider) List(userID string) ([]*identity.SIWE, error) {
	ss, err := p.Store.List(userID)
	if err != nil {
		return nil, err
	}

	sortIdentities(ss)
	return ss, nil
}

func (p *Provider) Get(userID, id string) (*identity.SIWE, error) {
	return p.Store.Get(userID, id)
}

func (p *Provider) GetByMessageRequest(messageRequest model.SIWEVerificationRequest) (*identity.SIWE, error) {
	message, _, err := p.SIWE.VerifyMessage(messageRequest)
	if err != nil {
		return nil, err
	}

	address := message.GetAddress()
	chainID := message.GetChainID()

	return p.Store.GetByAddress(chainID, address.Hex())
}

func (p *Provider) GetMany(ids []string) ([]*identity.SIWE, error) {
	return p.Store.GetMany(ids)
}

func (p *Provider) New(
	userID string,
	messageRequest model.SIWEVerificationRequest,
) (*identity.SIWE, error) {
	message, pubKey, err := p.SIWE.VerifyMessage(messageRequest)
	if err != nil {
		return nil, err
	}

	i := &identity.SIWE{
		ID:      uuid.New(),
		UserID:  userID,
		Address: message.GetAddress().Hex(),
		ChainID: message.GetChainID(),

		Data: &model.SIWEVerifiedData{
			Message:   messageRequest.Message,
			Signature: messageRequest.Signature,

			EncodedPublicKey: pubKey,
		},
	}
	return i, nil
}

func (p *Provider) Create(i *identity.SIWE) error {
	now := p.Clock.NowUTC()
	i.CreatedAt = now
	i.UpdatedAt = now
	return p.Store.Create(i)
}

func (p *Provider) Delete(i *identity.SIWE) error {
	return p.Store.Delete(i)
}

func sortIdentities(is []*identity.SIWE) {
	sort.Slice(is, func(i, j int) bool {
		return is[i].CreatedAt.Before(is[j].CreatedAt)
	})
}
