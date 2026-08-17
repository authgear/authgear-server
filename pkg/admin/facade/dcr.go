package facade

import (
	"context"

	"github.com/authgear/authgear-server/pkg/api/model"
	"github.com/authgear/authgear-server/pkg/lib/dcr"
)

type DCRCommands interface {
	CreateInitialAccessToken(ctx context.Context, options *dcr.NewInitialAccessTokenOptions) (string, *model.OAuthInitialAccessToken, error)
	RevokeInitialAccessToken(ctx context.Context, id string) error
}

type DCRQueries interface {
	ListInitialAccessTokens(ctx context.Context) ([]*model.OAuthInitialAccessToken, error)
}

type DCRFacade struct {
	DCRCommands DCRCommands
	DCRQueries  DCRQueries
}

func (f *DCRFacade) CreateInitialAccessToken(ctx context.Context, options *dcr.NewInitialAccessTokenOptions) (string, *model.OAuthInitialAccessToken, error) {
	return f.DCRCommands.CreateInitialAccessToken(ctx, options)
}

func (f *DCRFacade) RevokeInitialAccessToken(ctx context.Context, id string) error {
	return f.DCRCommands.RevokeInitialAccessToken(ctx, id)
}

func (f *DCRFacade) ListInitialAccessTokens(ctx context.Context) ([]*model.OAuthInitialAccessToken, error) {
	return f.DCRQueries.ListInitialAccessTokens(ctx)
}
