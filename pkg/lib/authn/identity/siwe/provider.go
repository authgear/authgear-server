package siwe

import (
	"context"
	"crypto/ecdsa"
	"sort"

	"github.com/authgear/authgear-server/pkg/api/model"
	"github.com/authgear/authgear-server/pkg/lib/authn/identity"
	"github.com/authgear/authgear-server/pkg/util/clock"
	"github.com/authgear/authgear-server/pkg/util/uuid"
)

// nolint: golint
type SIWEService interface {
	VerifyMessage(ctx context.Context, msg string, signature string) (*model.SIWEWallet, *ecdsa.PublicKey, error)
}

type Provider struct {
	Store *Store
	Clock clock.Clock
	SIWE  SIWEService
}

func (p *Provider) List(ctx context.Context, userID string) ([]*identity.SIWE, error) {
	ss, err := p.Store.List(ctx, userID)
	if err != nil {
		return nil, err
	}

	sortIdentities(ss)
	return ss, nil
}

func (p *Provider) Get(ctx context.Context, userID, id string) (*identity.SIWE, error) {
	return p.Store.Get(ctx, userID, id)
}

func (p *Provider) GetByMessage(ctx context.Context, msg string, signature string) (*identity.SIWE, error) {
	wallet, _, err := p.SIWE.VerifyMessage(ctx, msg, signature)
	if err != nil {
		return nil, err
	}

	return p.Store.GetByAddress(ctx, wallet.ChainID, wallet.Address)
}

func (p *Provider) GetMany(ctx context.Context, ids []string) ([]*identity.SIWE, error) {
	return p.Store.GetMany(ctx, ids)
}

func (p *Provider) New(
	ctx context.Context,
	userID string,
	msg string,
	signature string,
) (*identity.SIWE, error) {
	wallet, pubKey, err := p.SIWE.VerifyMessage(ctx, msg, signature)
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
		Address: wallet.Address,
		ChainID: wallet.ChainID,

		Data: &model.SIWEVerifiedData{
			Message:          msg,
			Signature:        signature,
			EncodedPublicKey: encodedPublicKey,
		},
	}
	return i, nil
}

func (p *Provider) Create(ctx context.Context, i *identity.SIWE) error {
	now := p.Clock.NowUTC()
	i.CreatedAt = now
	i.UpdatedAt = now
	return p.Store.Create(ctx, i)
}

func (p *Provider) Delete(ctx context.Context, i *identity.SIWE) error {
	return p.Store.Delete(ctx, i)
}

func sortIdentities(is []*identity.SIWE) {
	sort.Slice(is, func(i, j int) bool {
		return is[i].CreatedAt.Before(is[j].CreatedAt)
	})
}
