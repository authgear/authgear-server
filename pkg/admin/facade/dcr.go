package facade

import (
	"context"

	"github.com/authgear/authgear-server/pkg/api/model"
	"github.com/authgear/authgear-server/pkg/lib/dcr"
	"github.com/authgear/authgear-server/pkg/lib/infra/db"
	"github.com/authgear/authgear-server/pkg/lib/oauthclient"
	"github.com/authgear/authgear-server/pkg/util/graphqlutil"
)

type DCRCommands interface {
	CreateInitialAccessToken(ctx context.Context, options *dcr.NewInitialAccessTokenOptions) (string, *model.OAuthInitialAccessToken, error)
	RevokeInitialAccessToken(ctx context.Context, id string) (*model.OAuthInitialAccessToken, error)
	DeleteClient(ctx context.Context, clientID string) error
}

type DCRQueries interface {
	ListInitialAccessTokens(ctx context.Context) ([]*model.OAuthInitialAccessToken, error)
	ListClients(ctx context.Context, pageArgs graphqlutil.PageArgs) (*oauthclient.ListClientResult, error)
}

type DCRFacade struct {
	DCRCommands DCRCommands
	DCRQueries  DCRQueries
}

func (f *DCRFacade) CreateInitialAccessToken(ctx context.Context, options *dcr.NewInitialAccessTokenOptions) (string, *model.OAuthInitialAccessToken, error) {
	return f.DCRCommands.CreateInitialAccessToken(ctx, options)
}

func (f *DCRFacade) RevokeInitialAccessToken(ctx context.Context, id string) (*model.OAuthInitialAccessToken, error) {
	return f.DCRCommands.RevokeInitialAccessToken(ctx, id)
}

func (f *DCRFacade) ListInitialAccessTokens(ctx context.Context) ([]*model.OAuthInitialAccessToken, error) {
	return f.DCRQueries.ListInitialAccessTokens(ctx)
}

func (f *DCRFacade) DeleteClient(ctx context.Context, clientID string) error {
	return f.DCRCommands.DeleteClient(ctx, clientID)
}

func (f *DCRFacade) ListClients(ctx context.Context, pageArgs graphqlutil.PageArgs) ([]model.PageItemRef, *graphqlutil.PageResult, error) {
	result, err := f.DCRQueries.ListClients(ctx, pageArgs)
	if err != nil {
		return nil, nil, err
	}

	refs := make([]model.PageItemRef, len(result.Items))
	for i, c := range result.Items {
		i_uint64 := uint64(i) // #nosec G115
		pageKey := db.PageKey{Offset: result.Offset + i_uint64}
		cursor, err := pageKey.ToPageCursor()
		if err != nil {
			return nil, nil, err
		}
		refs[i] = model.PageItemRef{ID: c.ID, Cursor: cursor}
	}

	return refs, graphqlutil.NewPageResult(pageArgs, len(refs), graphqlutil.NewLazy(func() (any, error) {
		return result.TotalCount, nil
	})), nil
}
