package dcr

import (
	"context"

	"github.com/authgear/authgear-server/pkg/api/model"
	"github.com/authgear/authgear-server/pkg/util/clock"
)

type Queries struct {
	Store *Store
	Clock clock.Clock
}

func (q *Queries) GetInitialAccessTokenByID(ctx context.Context, id string) (*model.OAuthInitialAccessToken, error) {
	t, err := q.Store.GetInitialAccessTokenByID(ctx, id)
	if err != nil {
		return nil, err
	}
	return t.ToModel(), nil
}

func (q *Queries) ListInitialAccessTokens(ctx context.Context) ([]*model.OAuthInitialAccessToken, error) {
	ts, err := q.Store.ListActiveInitialAccessTokens(ctx)
	if err != nil {
		return nil, err
	}
	models := make([]*model.OAuthInitialAccessToken, len(ts))
	for i, t := range ts {
		models[i] = t.ToModel()
	}
	return models, nil
}

// ValidateAndGetByToken hashes the given plaintext bearer token, looks it up,
// and returns ErrInitialAccessTokenNotFound if it does not exist OR has
// expired (both cases must be indistinguishable to the caller — Part 2's
// registration handler maps both to `invalid_initial_access_token`).
func (q *Queries) ValidateAndGetByToken(ctx context.Context, plaintext string) (*model.OAuthInitialAccessToken, error) {
	hash := HashInitialAccessToken(plaintext)
	t, err := q.Store.GetInitialAccessTokenByHash(ctx, hash)
	if err != nil {
		return nil, err
	}
	if !t.ExpiresAt.After(q.Clock.NowUTC()) {
		return nil, ErrInitialAccessTokenNotFound
	}
	return t.ToModel(), nil
}
