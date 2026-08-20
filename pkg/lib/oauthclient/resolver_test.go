package oauthclient_test

import (
	"context"
	"net/url"
	"testing"
	"time"

	. "github.com/smartystreets/goconvey/convey"

	"github.com/authgear/authgear-server/pkg/api/model"
	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/lib/oauthclient"
	"github.com/authgear/authgear-server/pkg/lib/tester"
)

type testTesterEndpoints struct{}

func (testTesterEndpoints) TesterURL() *url.URL {
	return &url.URL{Scheme: "https", Host: "example.com", Path: "/tester"}
}

func TestResolverResolveClient(t *testing.T) {
	ctx := context.Background()

	Convey("Resolver.ResolveClient", t, func() {
		staticClient := config.OAuthClientConfig{ClientID: "static-client"}
		oauthConfig := &config.OAuthConfig{
			Clients: []config.OAuthClientConfig{staticClient},
		}

		Convey("resolves the tester client without touching Queries", func() {
			r := &oauthclient.Resolver{
				OAuthConfig:     oauthConfig,
				TesterEndpoints: testTesterEndpoints{},
				Queries:         nil, // must not be touched
			}
			client := r.ResolveClient(ctx, tester.ClientIDTester)
			So(client, ShouldNotBeNil)
		})

		Convey("resolves a static client without touching Queries", func() {
			r := &oauthclient.Resolver{
				OAuthConfig:     oauthConfig,
				TesterEndpoints: testTesterEndpoints{},
				Queries:         nil, // must not be touched
			}
			client := r.ResolveClient(ctx, "static-client")
			So(client, ShouldNotBeNil)
			So(client.ClientID, ShouldEqual, "static-client")
		})

		Convey("a non-dcrc_-prefixed unknown client_id returns nil without touching Queries", func() {
			r := &oauthclient.Resolver{
				OAuthConfig:     oauthConfig,
				TesterEndpoints: testTesterEndpoints{},
				Queries:         nil, // must not be touched — this is the fast path
			}
			client := r.ResolveClient(ctx, "some-unknown-id")
			So(client, ShouldBeNil)
		})

		Convey("a dcrc_-prefixed client_id not in static config reaches Queries", func() {
			cache, _ := newTestClientCache(t)
			r := &oauthclient.Resolver{
				OAuthConfig:     oauthConfig,
				TesterEndpoints: testTesterEndpoints{},
				Queries: &oauthclient.Queries{
					Cache:       cache,
					OAuthConfig: oauthConfig,
					Store:       nil, // must not be touched: cache is warmed below
				},
			}

			cachedClient := &oauthclient.Client{
				ID:              "row-id",
				ClientID:        "dcrc_test",
				Source:          model.OAuthClientSourceDCR,
				CreatedAt:       time.Now(),
				UpdatedAt:       time.Now(),
				Kind:            model.OAuthClientKindThirdParty,
				ApplicationType: "web",
				RedirectURIs:    []string{"https://example.com/callback"},
				GrantTypes:      []string{"authorization_code", "refresh_token"},
				ResponseTypes:   []string{"code"},
			}
			So(cache.Set(ctx, cachedClient), ShouldBeNil)

			client := r.ResolveClient(ctx, "dcrc_test")
			So(client, ShouldNotBeNil)
			So(client.ClientID, ShouldEqual, "dcrc_test")
		})

		Convey("a dcrc_-prefixed client_id that does not exist returns nil", func() {
			cache, _ := newTestClientCache(t)
			r := &oauthclient.Resolver{
				OAuthConfig:     oauthConfig,
				TesterEndpoints: testTesterEndpoints{},
				Queries: &oauthclient.Queries{
					Cache:       cache,
					OAuthConfig: oauthConfig,
				},
			}
			So(cache.SetNotFound(ctx, "dcrc_missing"), ShouldBeNil)

			client := r.ResolveClient(ctx, "dcrc_missing")
			So(client, ShouldBeNil)
		})
	})
}
