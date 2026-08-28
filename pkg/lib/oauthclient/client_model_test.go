package oauthclient_test

import (
	"testing"
	"time"

	. "github.com/smartystreets/goconvey/convey"

	"github.com/authgear/authgear-server/pkg/api/model"
	"github.com/authgear/authgear-server/pkg/lib/oauthclient"
)

func TestClientToModel(t *testing.T) {
	Convey("Client.ToModel", t, func() {
		now := time.Date(2026, 8, 17, 0, 0, 0, 0, time.UTC)
		clientName := "PR #123 preview"

		Convey("with client_name supplied", func() {
			c := &oauthclient.Client{
				ID:              "row-id",
				ClientID:        "dcrc_test",
				Source:          model.OAuthClientSourceDCR,
				CreatedAt:       now,
				UpdatedAt:       now,
				Kind:            model.OAuthClientKindThirdParty,
				ApplicationType: "web",
				ClientName:      &clientName,
				RedirectURIs:    []string{"https://example.com/callback"},
				GrantTypes:      []string{"authorization_code", "refresh_token"},
				ResponseTypes:   []string{"code"},
			}

			m := c.ToModel(nil)

			So(m.Name, ShouldEqual, "PR #123 preview")
			So(*m.ClientName, ShouldEqual, "PR #123 preview")
			So(m.IsConfidential, ShouldBeFalse)
			So(m.IsServiceClient, ShouldBeFalse)
			So(m.RefreshTokenRotationEnabled, ShouldBeFalse)
			So(m.IssueJWTAccessToken, ShouldBeFalse)
			So(m.MaxConcurrentSession, ShouldEqual, 0)
			So(m.CustomUIURI, ShouldBeNil)
			So(m.App2appEnabled, ShouldBeFalse)
			So(m.App2appInsecureDeviceKeyBindingEnabled, ShouldBeFalse)
			So(m.DPoPDisabled, ShouldBeFalse)
			So(m.PreAuthenticatedURLEnabled, ShouldBeFalse)
			So(m.PreAuthenticatedURLAllowedOrigins, ShouldBeEmpty)
			So(m.ReplaceProjectLogoWithLogoURI, ShouldBeFalse)
			So(m.PostLogoutRedirectURIs, ShouldResemble, []string{})
			So(m.RegisteredAt, ShouldNotBeNil)
			So(m.RegisteredAt.Equal(now), ShouldBeTrue)
			So(m.LastFetchedAt, ShouldBeNil)

			// Resolved token lifetimes with default_client_config nil — these
			// pin OAuthClientConfig.SetDefaults()'s built-in fallbacks, and
			// catch a reintroduction of hand-rolled fallback arithmetic.
			So(m.AccessTokenLifetimeSeconds, ShouldEqual, 1800)
			So(m.RefreshTokenLifetimeSeconds, ShouldEqual, 31449600)
			So(m.RefreshTokenIdleTimeoutEnabled, ShouldBeTrue)
			So(m.RefreshTokenIdleTimeoutSeconds, ShouldEqual, 2592000)
		})

		Convey("with client_name omitted falls back to generated display name", func() {
			c := &oauthclient.Client{
				ID:              "row-id",
				ClientID:        "dcrc_test",
				Source:          model.OAuthClientSourceDCR,
				CreatedAt:       now,
				UpdatedAt:       now,
				Kind:            model.OAuthClientKindFirstParty,
				ApplicationType: "native",
				RedirectURIs:    []string{"com.example.app://callback"},
				GrantTypes:      []string{"authorization_code", "refresh_token"},
				ResponseTypes:   []string{"code"},
			}

			m := c.ToModel(nil)

			So(m.Name, ShouldEqual, "Client dcrc_test")
			So(m.ClientName, ShouldBeNil)
		})

		Convey("CIMD source has nil RegisteredAt", func() {
			c := &oauthclient.Client{
				ID:              "row-id",
				ClientID:        "https://mcp.example.com/client",
				Source:          model.OAuthClientSourceCIMD,
				CreatedAt:       now,
				UpdatedAt:       now,
				Kind:            model.OAuthClientKindThirdParty,
				ApplicationType: "web",
				RedirectURIs:    []string{"https://example.com/callback"},
				GrantTypes:      []string{"authorization_code", "refresh_token"},
				ResponseTypes:   []string{"code"},
			}

			m := c.ToModel(nil)
			So(m.RegisteredAt, ShouldBeNil)
		})
	})
}
