package oauth

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"

	"github.com/authgear/authgear-server/pkg/lib/config"
)

func TestConsentViewModelForClient(t *testing.T) {
	Convey("consentViewModelForClient", t, func() {
		Convey("static client with client_name: ClientName is the client_name", func() {
			client := &config.OAuthClientConfig{
				ClientID: "static-client",
				Name:     "Foo",
			}
			vm := consentViewModelForClient(client)
			So(vm.ClientName, ShouldEqual, "Foo")
		})

		Convey("static client with no client_name: ClientName is the config's Name, not empty", func() {
			// Name is always populated by config parsing/defaulting for a
			// static client (it falls back to ClientID if client_name is
			// unset) -- the bug this fixes was reading ClientName directly,
			// which IS empty in that case.
			client := &config.OAuthClientConfig{
				ClientID: "static-client",
				Name:     "static-client",
			}
			vm := consentViewModelForClient(client)
			So(vm.ClientName, ShouldEqual, "static-client")
			So(vm.ClientName, ShouldNotBeEmpty)
		})

		Convey("a client with an empty ClientName field but a non-empty Name: ClientName reads from Name, never renders empty", func() {
			// This is the literal regression case: oauthclient.Client's
			// ToClientConfig leaves ClientName ("") distinct from Name
			// (the DisplayName() fallback) whenever client_name was
			// omitted from a dynamic client's registration/document.
			client := &config.OAuthClientConfig{
				ClientID:   "https://mcp-client.example.com/oauth/client-metadata.json",
				ClientName: "",
				Name:       "Client https://mcp-client.example.com/oauth/client-metadata.json",
			}
			vm := consentViewModelForClient(client)
			So(vm.ClientName, ShouldEqual, "Client https://mcp-client.example.com/oauth/client-metadata.json")
		})

		Convey("PolicyURI and TOSURI pass through unchanged", func() {
			client := &config.OAuthClientConfig{
				ClientID:  "static-client",
				Name:      "Foo",
				PolicyURI: "https://example.com/policy",
				TOSURI:    "https://example.com/tos",
			}
			vm := consentViewModelForClient(client)
			So(vm.ClientPolicyURI, ShouldEqual, "https://example.com/policy")
			So(vm.ClientTOSURI, ShouldEqual, "https://example.com/tos")
		})
	})
}
