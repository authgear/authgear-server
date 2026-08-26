package oauthclient_test

import (
	"context"
	"testing"
	"time"

	. "github.com/smartystreets/goconvey/convey"

	"github.com/authgear/authgear-server/pkg/api/model"
	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/lib/oauthclient"
)

// TestGetClientConfigByClientIDCacheHit covers the branches of
// getClientByClientIDCached that a cache hit reaches without ever touching
// Store or opening a database scope — Store is deliberately left nil here,
// so any accidental fallthrough to it would panic and fail the test.
func TestGetClientConfigByClientIDCacheHit(t *testing.T) {
	ctx := context.Background()

	Convey("Queries.GetClientConfigByClientID on a cache hit", t, func() {
		cache, _ := newTestClientCache(t)
		queries := &oauthclient.Queries{
			Cache:       cache,
			OAuthConfig: &config.OAuthConfig{},
			Store:       nil, // must not be touched by either sub-test below
		}

		Convey("positive cache hit returns the cached client's config, without touching Store", func() {
			clientName := "PR #123 preview"
			client := &oauthclient.Client{
				ID:              "row-id",
				ClientID:        "dcrc_test",
				Source:          model.OAuthClientSourceDCR,
				CreatedAt:       time.Now(),
				UpdatedAt:       time.Now(),
				Kind:            model.OAuthClientKindThirdParty,
				ApplicationType: "web",
				ClientName:      &clientName,
				RedirectURIs:    []string{"https://example.com/callback"},
				GrantTypes:      []string{"authorization_code", "refresh_token"},
				ResponseTypes:   []string{"code"},
			}
			So(cache.Set(ctx, client), ShouldBeNil)

			cfg, err := queries.GetClientConfigByClientID(ctx, "dcrc_test")
			So(err, ShouldBeNil)
			So(cfg.ClientID, ShouldEqual, "dcrc_test")
		})

		Convey("cached negative result returns ErrDynamicClientNotFound, without touching Store", func() {
			So(cache.SetNotFound(ctx, "dcrc_missing"), ShouldBeNil)

			_, err := queries.GetClientConfigByClientID(ctx, "dcrc_missing")
			So(err, ShouldEqual, oauthclient.ErrDynamicClientNotFound)
		})
	})
}
