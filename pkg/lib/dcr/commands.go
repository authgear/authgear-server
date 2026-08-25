package dcr

import (
	"context"

	"github.com/authgear/authgear-server/pkg/api/model"
	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/lib/oauthclient"
)

type Commands struct {
	Store              *Store
	OAuthClient        *oauthclient.Commands
	OAuthClientQueries *oauthclient.Queries
	OAuthConfig        *config.OAuthConfig
}

// CountClientsBySource is re-exported from oauthclient.Queries so that
// RegistrationHandlerDCRService depends on one collaborator rather than
// reaching into pkg/lib/oauthclient directly. Present now for symmetry;
// wired up by the client usage-limit check (docs/plans/dcr/2026-08-17-05-client-usage-limit.md).
func (c *Commands) CountClientsBySource(ctx context.Context, source model.OAuthClientSource) (uint64, error) {
	return c.OAuthClientQueries.CountClientsBySource(ctx, source)
}

// LockForClientCount is re-exported from oauthclient.Commands, mirroring
// CountClientsBySource above. See oauthclient.Store.LockForClientCount.
func (c *Commands) LockForClientCount(ctx context.Context, source model.OAuthClientSource) error {
	return c.OAuthClient.LockForClientCount(ctx, source)
}

func (c *Commands) CreateInitialAccessToken(ctx context.Context, options *NewInitialAccessTokenOptions) (plaintext string, iat *model.OAuthInitialAccessToken, err error) {
	plaintext, hash := GenerateInitialAccessToken(options.Type)
	t := c.Store.NewInitialAccessToken(options, hash)
	if err := c.Store.CreateInitialAccessToken(ctx, t); err != nil {
		return "", nil, err
	}
	return plaintext, t.ToModel(), nil
}

// RevokeInitialAccessToken reads the token row before deleting it so the
// caller can audit-log what was revoked; the row is gone afterwards and the
// audit entry is the only remaining record of its type.
func (c *Commands) RevokeInitialAccessToken(ctx context.Context, id string) (*model.OAuthInitialAccessToken, error) {
	t, err := c.Store.GetInitialAccessTokenByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if err := c.Store.DeleteInitialAccessToken(ctx, id); err != nil {
		return nil, err
	}
	return t.ToModel(), nil
}

// DeleteClient is re-exported from oauthclient.Commands so that
// pkg/admin/facade.DCRCommands depends on one collaborator rather than
// reaching into pkg/lib/oauthclient directly.
func (c *Commands) DeleteClient(ctx context.Context, clientID string) error {
	return c.OAuthClient.DeleteClient(ctx, clientID)
}
