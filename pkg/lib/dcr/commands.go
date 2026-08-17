package dcr

import (
	"context"

	"github.com/authgear/authgear-server/pkg/api/model"
)

type Commands struct {
	Store *Store
}

func (c *Commands) CreateInitialAccessToken(ctx context.Context, options *NewInitialAccessTokenOptions) (plaintext string, iat *model.OAuthInitialAccessToken, err error) {
	plaintext, hash := GenerateInitialAccessToken(options.Type)
	t := c.Store.NewInitialAccessToken(options, hash)
	if err := c.Store.CreateInitialAccessToken(ctx, t); err != nil {
		return "", nil, err
	}
	return plaintext, t.ToModel(), nil
}

func (c *Commands) RevokeInitialAccessToken(ctx context.Context, id string) error {
	return c.Store.DeleteInitialAccessToken(ctx, id)
}
