package siwe

import (
	"context"
	"fmt"
	"sort"

	"github.com/authgear/authgear-server/pkg/lib/authn/identity"
	"github.com/authgear/authgear-server/pkg/util/clock"
)

type Provider struct {
	Store *Store
	Clock clock.Clock
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
	panic(fmt.Errorf("siwe: SIWE is deprecated"))
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
	panic(fmt.Errorf("siwe: SIWE is deprecated"))
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
