package siwe

import (
	"crypto/ecdsa"
	"sort"

	"github.com/authgear/authgear-server/pkg/api/model"
	"github.com/authgear/authgear-server/pkg/lib/authn/identity"
	"github.com/authgear/authgear-server/pkg/util/clock"
	"github.com/authgear/authgear-server/pkg/util/uuid"
	siwego "github.com/spruceid/siwe-go"
)

// nolint: golint
type SIWEService interface {
	VerifyMessage(msg string, signature string) (*siwego.Message, *ecdsa.PublicKey, error)
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

func (p *Provider) GetByMessage(msg string) (*identity.SIWE, error) {
	message, err := siwego.ParseMessage(msg)
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
	msg string,
	signature string,
) (*identity.SIWE, error) {
	message, pubKey, err := p.SIWE.VerifyMessage(msg, signature)
	if err != nil {
		return nil, err
	}

	encodedPublicKey, err := model.NewSIWEPublicKey(pubKey)
	if err != nil {
		return nil, err
	}

	i := &identity.SIWE{
		ID:      uuid.New(),
		UserID:  userID,
		Address: message.GetAddress().Hex(),
		ChainID: message.GetChainID(),

		Data: &model.SIWEVerifiedData{
			Message:          msg,
			Signature:        signature,
			EncodedPublicKey: encodedPublicKey,
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
