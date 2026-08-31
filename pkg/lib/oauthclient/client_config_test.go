package oauthclient_test

import (
	"testing"
	"time"

	. "github.com/smartystreets/goconvey/convey"

	"github.com/authgear/authgear-server/pkg/api/model"
	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/lib/oauthclient"
)

func TestResolveTokenLifetimes(t *testing.T) {
	Convey("ResolveTokenLifetimes", t, func() {
		dcrLifetimes := &config.OAuthDynamicClientTokenLifetimesConfig{
			AccessTokenLifetime:  111,
			RefreshTokenLifetime: 222,
		}
		cimdLifetimes := &config.OAuthDynamicClientTokenLifetimesConfig{
			AccessTokenLifetime:  333,
			RefreshTokenLifetime: 444,
		}
		oauthConfig := &config.OAuthConfig{
			DynamicClientRegistration: &config.OAuthDynamicClientRegistrationConfig{
				DefaultClientConfig: dcrLifetimes,
			},
			ClientIDMetadataDocument: &config.OAuthClientIDMetadataDocumentConfig{
				ClientConfig: cimdLifetimes,
			},
		}

		Convey("DCR reads dynamic_client_registration.default_client_config", func() {
			got := oauthclient.ResolveTokenLifetimes(oauthConfig, model.OAuthClientSourceDCR)
			So(got, ShouldEqual, dcrLifetimes)
		})

		Convey("CIMD reads client_id_metadata_document.client_config, not DCR's key", func() {
			got := oauthclient.ResolveTokenLifetimes(oauthConfig, model.OAuthClientSourceCIMD)
			So(got, ShouldEqual, cimdLifetimes)
			So(got, ShouldNotEqual, dcrLifetimes)
		})

		Convey("an unrecognized source (e.g. static) resolves to nil, not a panic", func() {
			So(func() {
				got := oauthclient.ResolveTokenLifetimes(oauthConfig, model.OAuthClientSourceStatic)
				So(got, ShouldBeNil)
			}, ShouldNotPanic)
		})
	})
}

func TestClientToClientConfigCIMD(t *testing.T) {
	Convey("Client.ToClientConfig for a CIMD-sourced, third-party row", t, func() {
		now := time.Date(2026, 8, 31, 0, 0, 0, 0, time.UTC)
		c := &oauthclient.Client{
			ID:              "row-id",
			ClientID:        "https://mcp-client.example.com/oauth/client-metadata.json",
			Source:          model.OAuthClientSourceCIMD,
			CreatedAt:       now,
			UpdatedAt:       now,
			LastFetchedAt:   &now,
			Kind:            model.OAuthClientKindThirdParty,
			ApplicationType: "web",
			RedirectURIs:    []string{"http://127.0.0.1:3000/callback", "http://localhost:3000/callback"},
			GrantTypes:      []string{"authorization_code", "refresh_token"},
			ResponseTypes:   []string{"code"},
		}

		cimdLifetimes := &config.OAuthDynamicClientTokenLifetimesConfig{
			AccessTokenLifetime:            1800,
			RefreshTokenLifetime:           2592000,
			RefreshTokenIdleTimeoutEnabled: func() *bool { b := true; return &b }(),
			RefreshTokenIdleTimeout:        1209600,
		}
		dcrLifetimes := &config.OAuthDynamicClientTokenLifetimesConfig{
			AccessTokenLifetime:            999,
			RefreshTokenLifetime:           999,
			RefreshTokenIdleTimeoutEnabled: func() *bool { b := false; return &b }(),
			RefreshTokenIdleTimeout:        999,
		}
		oauthConfig := &config.OAuthConfig{
			DynamicClientRegistration: &config.OAuthDynamicClientRegistrationConfig{
				DefaultClientConfig: dcrLifetimes,
			},
			ClientIDMetadataDocument: &config.OAuthClientIDMetadataDocumentConfig{
				ClientConfig: cimdLifetimes,
			},
		}

		cfg := c.ToClientConfig(oauthclient.ResolveTokenLifetimes(oauthConfig, c.Source))

		Convey("application type is third-party and public", func() {
			So(cfg.ApplicationType, ShouldEqual, config.OAuthClientApplicationTypeDynamicThirdParty)
			So(cfg.ApplicationType.IsThirdParty(), ShouldBeTrue)
			So(cfg.ApplicationType.IsConfidential(), ShouldBeFalse)
			So(cfg.ApplicationType.IsPublic(), ShouldBeTrue)
			So(cfg.ApplicationType.HasFullAccessScope(), ShouldBeFalse)
			So(cfg.ApplicationType.IsClientCredentialsFlowAllowed(), ShouldBeFalse)
		})

		Convey("dynamic and CIMD flags", func() {
			So(cfg.IsDynamicClient(), ShouldBeTrue)
			So(cfg.IsCIMDClient(), ShouldBeTrue)
		})

		Convey("lifetimes come from client_config, not default_client_config", func() {
			So(cfg.AccessTokenLifetime, ShouldEqual, config.DurationSeconds(1800))
			So(cfg.RefreshTokenLifetime, ShouldEqual, config.DurationSeconds(2592000))
			So(*cfg.RefreshTokenIdleTimeoutEnabled, ShouldBeTrue)
			So(cfg.RefreshTokenIdleTimeout, ShouldEqual, config.DurationSeconds(1209600))
		})

		Convey("through ToModel: RegisteredAt nil, LastFetchedAt passed through, Source CIMD, PostLogoutRedirectURIs empty", func() {
			m := c.ToModel(cimdLifetimes)
			So(m.RegisteredAt, ShouldBeNil)
			So(m.LastFetchedAt, ShouldNotBeNil)
			So(m.LastFetchedAt.Equal(now), ShouldBeTrue)
			So(m.Source, ShouldEqual, model.OAuthClientSourceCIMD)
			So(m.PostLogoutRedirectURIs, ShouldResemble, []string{})
		})
	})

	Convey("a static client's OAuthClientConfig is never a CIMD client, even with an https:// client_id", t, func() {
		cfg := &config.OAuthClientConfig{ClientID: "https://example.com/client"}
		So(cfg.IsCIMDClient(), ShouldBeFalse)
		So(cfg.IsDynamicClient(), ShouldBeFalse)
	})
}
