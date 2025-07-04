package challenge

import (
	"context"
	"fmt"

	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/util/clock"
)

type Provider struct {
	Store *Store
	AppID config.AppID
	Clock clock.Clock
}

func (p *Provider) Create(ctx context.Context, purpose Purpose) (*Challenge, error) {
	now := p.Clock.NowUTC()
	ttl := purpose.ValidityPeriod()
	c := &Challenge{
		Token:     GenerateChallengeToken(),
		Purpose:   purpose,
		CreatedAt: now,
		ExpireAt:  now.Add(ttl),
	}

	err := p.Store.Save(ctx, c, ttl)
	if err != nil {
		return nil, err
	}

	return c, nil
}

func (p *Provider) Get(ctx context.Context, token string) (*Challenge, error) {
	return p.Store.Get(ctx, token, false)
}

func (p *Provider) Consume(ctx context.Context, token string) (*Purpose, error) {
	c, err := p.Store.Get(ctx, token, true)
	if err != nil {
		return nil, err
	}
	return &c.Purpose, nil
}

func challengeKey(appID config.AppID, token string) string {
	return fmt.Sprintf("app:%s:challenge:%s", appID, token)
}
