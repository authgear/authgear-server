package accountmigration

import (
	"context"
	"fmt"
	"net/url"

	"github.com/authgear/authgear-server/pkg/lib/config"
)

type Service struct {
	Config   *config.AccountMigrationHookConfig
	DenoHook *AccountMigrationDenoHook
	WebHook  *AccountMigrationWebHook
}

func (s *Service) Run(ctx context.Context, migrationTokenString string) (*HookResponse, error) {
	if s.Config.URL == "" {
		return nil, InvalidConfiguration.New("missing account migration hook config")
	}

	u, err := url.Parse(s.Config.URL)
	if err != nil {
		return nil, err
	}

	req := &HookRequest{
		MigrationToken: migrationTokenString,
	}

	switch {
	case s.DenoHook.SupportURL(u):
		return s.DenoHook.Call(ctx, u, req)
	case s.WebHook.SupportURL(u):
		return s.WebHook.Call(ctx, u, req)
	default:
		return nil, fmt.Errorf("unsupported hook URL: %v", u)
	}
}
