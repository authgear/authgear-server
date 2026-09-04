package oauthclient_test

import (
	"testing"
	"time"

	. "github.com/smartystreets/goconvey/convey"

	"github.com/authgear/authgear-server/pkg/lib/oauthclient"
)

func baseClient() *oauthclient.Client {
	return &oauthclient.Client{
		ID:              "client-row-id",
		ClientID:        "https://mcp-client.example.com/oauth/client-metadata.json",
		Source:          oauthclient.Source("CIMD"),
		CreatedAt:       time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC),
		UpdatedAt:       time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC),
		Kind:            oauthclient.Kind("THIRD_PARTY"),
		ApplicationType: "web",
		ClientName:      new("Example Client"),
		ClientURI:       new("https://mcp-client.example.com"),
		LogoURI:         new("https://mcp-client.example.com/logo.png"),
		TOSURI:          new("https://mcp-client.example.com/tos"),
		PolicyURI:       new("https://mcp-client.example.com/policy"),
		RedirectURIs:    []string{"http://127.0.0.1:3000/callback"},
		GrantTypes:      []string{"authorization_code", "refresh_token"},
		ResponseTypes:   []string{"code"},
	}
}

func baseOptions() *oauthclient.UpsertCIMDClientOptions {
	c := baseClient()
	return &oauthclient.UpsertCIMDClientOptions{
		ClientID:        c.ClientID,
		ApplicationType: c.ApplicationType,
		ClientName:      c.ClientName,
		ClientURI:       c.ClientURI,
		LogoURI:         c.LogoURI,
		TOSURI:          c.TOSURI,
		PolicyURI:       c.PolicyURI,
		RedirectURIs:    c.RedirectURIs,
		GrantTypes:      c.GrantTypes,
		ResponseTypes:   c.ResponseTypes,
	}
}

func TestClientMetadataChangedFrom(t *testing.T) {
	Convey("Client.MetadataChangedFrom", t, func() {
		Convey("identical metadata: false", func() {
			c := baseClient()
			So(c.MetadataChangedFrom(baseOptions()), ShouldBeFalse)
		})

		Convey("client_name changed: true", func() {
			c := baseClient()
			options := baseOptions()
			options.ClientName = new("New Name")
			So(c.MetadataChangedFrom(options), ShouldBeTrue)
		})

		Convey("client_uri changed: true", func() {
			c := baseClient()
			options := baseOptions()
			options.ClientURI = new("https://new.example.com")
			So(c.MetadataChangedFrom(options), ShouldBeTrue)
		})

		Convey("logo_uri changed: true", func() {
			c := baseClient()
			options := baseOptions()
			options.LogoURI = new("https://new.example.com/logo.png")
			So(c.MetadataChangedFrom(options), ShouldBeTrue)
		})

		Convey("tos_uri changed: true", func() {
			c := baseClient()
			options := baseOptions()
			options.TOSURI = new("https://new.example.com/tos")
			So(c.MetadataChangedFrom(options), ShouldBeTrue)
		})

		Convey("policy_uri changed: true", func() {
			c := baseClient()
			options := baseOptions()
			options.PolicyURI = new("https://new.example.com/policy")
			So(c.MetadataChangedFrom(options), ShouldBeTrue)
		})

		Convey("application_type changed: true", func() {
			c := baseClient()
			options := baseOptions()
			options.ApplicationType = "native"
			So(c.MetadataChangedFrom(options), ShouldBeTrue)
		})

		Convey("redirect_uris changed: true", func() {
			c := baseClient()
			options := baseOptions()
			options.RedirectURIs = []string{"http://127.0.0.1:3000/new-callback"}
			So(c.MetadataChangedFrom(options), ShouldBeTrue)
		})

		Convey("redirect_uris reordered only: false -- set comparison", func() {
			c := baseClient()
			c.RedirectURIs = []string{"http://127.0.0.1:3000/a", "http://127.0.0.1:3000/b"}
			options := baseOptions()
			options.RedirectURIs = []string{"http://127.0.0.1:3000/b", "http://127.0.0.1:3000/a"}
			So(c.MetadataChangedFrom(options), ShouldBeFalse)
		})

		Convey("grant_types changed: true", func() {
			c := baseClient()
			options := baseOptions()
			options.GrantTypes = []string{"authorization_code"}
			So(c.MetadataChangedFrom(options), ShouldBeTrue)
		})

		Convey("response_types changed: true", func() {
			c := baseClient()
			options := baseOptions()
			options.ResponseTypes = []string{}
			So(c.MetadataChangedFrom(options), ShouldBeTrue)
		})

		Convey("client_name nil -> \"\": false -- normalization", func() {
			c := baseClient()
			c.ClientName = nil
			options := baseOptions()
			options.ClientName = new("")
			So(c.MetadataChangedFrom(options), ShouldBeFalse)
		})

		Convey("client_name \"\" -> nil: false -- normalization, symmetric", func() {
			c := baseClient()
			c.ClientName = new("")
			options := baseOptions()
			options.ClientName = nil
			So(c.MetadataChangedFrom(options), ShouldBeFalse)
		})

		Convey("multiple fields changed: true", func() {
			c := baseClient()
			options := baseOptions()
			options.ClientName = new("New Name")
			options.LogoURI = new("https://new.example.com/logo.png")
			options.RedirectURIs = []string{"http://127.0.0.1:3000/new-callback"}
			So(c.MetadataChangedFrom(options), ShouldBeTrue)
		})

		Convey("only non-document fields differ (ID, timestamps): false -- those are never compared", func() {
			c := baseClient()
			c.ID = "different-row-id"
			c.CreatedAt = c.CreatedAt.Add(24 * time.Hour)
			c.UpdatedAt = c.UpdatedAt.Add(24 * time.Hour)
			later := c.UpdatedAt.Add(time.Hour)
			c.LastFetchedAt = &later
			So(c.MetadataChangedFrom(baseOptions()), ShouldBeFalse)
		})
	})
}
