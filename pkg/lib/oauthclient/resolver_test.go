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

		cimdOAuthConfig := &config.OAuthConfig{
			Clients: []config.OAuthClientConfig{staticClient},
			ClientIDMetadataDocument: &config.OAuthClientIDMetadataDocumentConfig{
				Enabled: true,
			},
		}
		const cimdClientID = "https://mcp-client.example.com/oauth/client-metadata.json"

		Convey("CIMD disabled: an unknown URL-shaped client_id reaches Queries and returns nil", func() {
			cache, _ := newTestClientCache(t)
			r := &oauthclient.Resolver{
				OAuthConfig:     oauthConfig, // CIMD not configured -> disabled
				TesterEndpoints: testTesterEndpoints{},
				Queries: &oauthclient.Queries{
					Cache:       cache,
					OAuthConfig: oauthConfig,
				},
			}
			So(cache.SetNotFound(ctx, cimdClientID), ShouldBeNil)

			client := r.ResolveClient(ctx, cimdClientID)
			So(client, ShouldBeNil)
		})

		Convey("CIMD disabled: an already-persisted CIMD client_id still resolves (enabled gates fetching, not reading)", func() {
			cache, _ := newTestClientCache(t)
			r := &oauthclient.Resolver{
				OAuthConfig:     oauthConfig, // CIMD not configured -> disabled
				TesterEndpoints: testTesterEndpoints{},
				Queries: &oauthclient.Queries{
					Cache:       cache,
					OAuthConfig: oauthConfig,
				},
			}
			cachedClient := &oauthclient.Client{
				ID:              "row-id",
				ClientID:        cimdClientID,
				Source:          model.OAuthClientSourceCIMD,
				CreatedAt:       time.Now(),
				UpdatedAt:       time.Now(),
				Kind:            model.OAuthClientKindThirdParty,
				ApplicationType: "web",
				RedirectURIs:    []string{"http://127.0.0.1:3000/callback"},
				GrantTypes:      []string{"authorization_code", "refresh_token"},
				ResponseTypes:   []string{"code"},
			}
			So(cache.Set(ctx, cachedClient), ShouldBeNil)

			client := r.ResolveClient(ctx, cimdClientID)
			So(client, ShouldNotBeNil)
			So(client.ClientID, ShouldEqual, cimdClientID)
			So(client.IsCIMDClient(), ShouldBeTrue)
		})

		Convey("CIMD enabled: a URL-shaped client_id reaches Queries.GetClientConfigByClientID with the exact string", func() {
			cache, _ := newTestClientCache(t)
			r := &oauthclient.Resolver{
				OAuthConfig:     cimdOAuthConfig,
				TesterEndpoints: testTesterEndpoints{},
				Queries: &oauthclient.Queries{
					Cache:       cache,
					OAuthConfig: cimdOAuthConfig,
				},
			}

			cachedClient := &oauthclient.Client{
				ID:              "row-id",
				ClientID:        cimdClientID,
				Source:          model.OAuthClientSourceCIMD,
				CreatedAt:       time.Now(),
				UpdatedAt:       time.Now(),
				Kind:            model.OAuthClientKindThirdParty,
				ApplicationType: "web",
				RedirectURIs:    []string{"http://127.0.0.1:3000/callback"},
				GrantTypes:      []string{"authorization_code", "refresh_token"},
				ResponseTypes:   []string{"code"},
			}
			So(cache.Set(ctx, cachedClient), ShouldBeNil)

			client := r.ResolveClient(ctx, cimdClientID)
			So(client, ShouldNotBeNil)
			So(client.ClientID, ShouldEqual, cimdClientID)
			So(client.IsCIMDClient(), ShouldBeTrue)
		})

		Convey("CIMD enabled: a resolvable row on a host outside allowed_domains still resolves (allowed_domains gates fetching, not reading)", func() {
			restrictedOAuthConfig := &config.OAuthConfig{
				Clients: []config.OAuthClientConfig{staticClient},
				ClientIDMetadataDocument: &config.OAuthClientIDMetadataDocumentConfig{
					Enabled:        true,
					AllowedDomains: []string{"only-this-domain.example.com"},
				},
			}
			cache, _ := newTestClientCache(t)
			r := &oauthclient.Resolver{
				OAuthConfig:     restrictedOAuthConfig,
				TesterEndpoints: testTesterEndpoints{},
				Queries: &oauthclient.Queries{
					Cache:       cache,
					OAuthConfig: restrictedOAuthConfig,
				},
			}
			cachedClient := &oauthclient.Client{
				ID:              "row-id",
				ClientID:        cimdClientID, // host mcp-client.example.com, NOT in allowed_domains
				Source:          model.OAuthClientSourceCIMD,
				CreatedAt:       time.Now(),
				UpdatedAt:       time.Now(),
				Kind:            model.OAuthClientKindThirdParty,
				ApplicationType: "web",
				RedirectURIs:    []string{"http://127.0.0.1:3000/callback"},
				GrantTypes:      []string{"authorization_code", "refresh_token"},
				ResponseTypes:   []string{"code"},
			}
			So(cache.Set(ctx, cachedClient), ShouldBeNil)

			client := r.ResolveClient(ctx, cimdClientID)
			So(client, ShouldNotBeNil)
			So(client.ClientID, ShouldEqual, cimdClientID)
		})

		Convey("CIMD enabled: a resolvable http:// row reaches Queries regardless of insecure_http_allowed", func() {
			const httpClientID = "http://x.example.com/y"
			cache, _ := newTestClientCache(t)
			r := &oauthclient.Resolver{
				OAuthConfig:     cimdOAuthConfig, // insecure_http_allowed not set anywhere -- irrelevant to the read path
				TesterEndpoints: testTesterEndpoints{},
				Queries: &oauthclient.Queries{
					Cache:       cache,
					OAuthConfig: cimdOAuthConfig,
				},
			}
			cachedClient := &oauthclient.Client{
				ID:              "row-id",
				ClientID:        httpClientID,
				Source:          model.OAuthClientSourceCIMD,
				CreatedAt:       time.Now(),
				UpdatedAt:       time.Now(),
				Kind:            model.OAuthClientKindThirdParty,
				ApplicationType: "web",
				RedirectURIs:    []string{"http://127.0.0.1:3000/callback"},
				GrantTypes:      []string{"authorization_code", "refresh_token"},
				ResponseTypes:   []string{"code"},
			}
			So(cache.Set(ctx, cachedClient), ShouldBeNil)

			client := r.ResolveClient(ctx, httpClientID)
			So(client, ShouldNotBeNil)
			So(client.ClientID, ShouldEqual, httpClientID)
		})

		Convey("CIMD enabled: a static client whose client_id happens to be an https:// URL is returned as static, without touching Queries", func() {
			preRegisteredStatic := config.OAuthClientConfig{ClientID: cimdClientID}
			preRegisteredOAuthConfig := &config.OAuthConfig{
				Clients: []config.OAuthClientConfig{preRegisteredStatic},
				ClientIDMetadataDocument: &config.OAuthClientIDMetadataDocumentConfig{
					Enabled: true,
				},
			}
			r := &oauthclient.Resolver{
				OAuthConfig:     preRegisteredOAuthConfig,
				TesterEndpoints: testTesterEndpoints{},
				Queries:         nil, // must not be touched: static lookup wins first
			}
			client := r.ResolveClient(ctx, cimdClientID)
			So(client, ShouldNotBeNil)
			So(client.ClientID, ShouldEqual, cimdClientID)
			So(client.IsCIMDClient(), ShouldBeFalse)
		})

		Convey("CIMD enabled: a dcrc_-prefixed client_id is still resolved (no regression)", func() {
			cache, _ := newTestClientCache(t)
			r := &oauthclient.Resolver{
				OAuthConfig:     cimdOAuthConfig,
				TesterEndpoints: testTesterEndpoints{},
				Queries: &oauthclient.Queries{
					Cache:       cache,
					OAuthConfig: cimdOAuthConfig,
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
	})
}
