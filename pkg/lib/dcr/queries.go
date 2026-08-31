package dcr

import (
	"context"

	"github.com/authgear/authgear-server/pkg/api/model"
	"github.com/authgear/authgear-server/pkg/lib/oauthclient"
	"github.com/authgear/authgear-server/pkg/util/clock"
	"github.com/authgear/authgear-server/pkg/util/graphqlutil"
)

type Queries struct {
	Store              *Store
	Clock              clock.Clock
	OAuthClientQueries *oauthclient.Queries
}

// ListClients is re-exported from oauthclient.Queries so that
// pkg/admin/facade.DCRQueries depends on one collaborator rather than
// reaching into pkg/lib/oauthclient directly.
func (q *Queries) ListClients(ctx context.Context, pageArgs graphqlutil.PageArgs) (*oauthclient.ListClientResult, error) {
	return q.OAuthClientQueries.ListClients(ctx, pageArgs)
}

func (q *Queries) GetInitialAccessTokenByID(ctx context.Context, id string) (*model.OAuthInitialAccessToken, error) {
	t, err := q.Store.GetInitialAccessTokenByID(ctx, id)
	if err != nil {
		return nil, err
	}
	return t.ToModel(), nil
}

func (q *Queries) GetManyInitialAccessTokens(ctx context.Context, ids []string) ([]*model.OAuthInitialAccessToken, error) {
	ts, err := q.Store.GetManyInitialAccessTokensByID(ctx, ids)
	if err != nil {
		return nil, err
	}
	models := make([]*model.OAuthInitialAccessToken, len(ts))
	for i, t := range ts {
		models[i] = t.ToModel()
	}
	return models, nil
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

// ValidateAndGetByToken hashes the given plaintext bearer token and looks
// it up. It returns ErrInitialAccessTokenNotFound if no such token exists,
// and ErrInitialAccessTokenExpired -- together with the token's own model,
// non-nil -- if it exists but has expired. The two are indistinguishable
// in the HTTP response Part 2's registration handler produces (both map to
// `invalid_initial_access_token`), but distinguishable in the audit log
// (Part 8) -- see ErrInitialAccessTokenExpired's own doc comment for why.
//
// Returning a non-nil token together with a non-nil error is unusual in
// this codebase and is deliberate here, not a slip: the caller must still
// reject the registration (check the error first), but the audit event
// needs the row to describe. The token is for reporting only, never for
// authorizing -- do not use it to proceed with registration.
func (q *Queries) ValidateAndGetByToken(ctx context.Context, plaintext string) (*model.OAuthInitialAccessToken, error) {
	hash := HashInitialAccessToken(plaintext)
	t, err := q.Store.GetInitialAccessTokenByHash(ctx, hash)
	if err != nil {
		return nil, err
	}
	if !t.ExpiresAt.After(q.Clock.NowUTC()) {
		return t.ToModel(), ErrInitialAccessTokenExpired
	}
	return t.ToModel(), nil
}
