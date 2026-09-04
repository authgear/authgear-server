package oauth

import (
	"net/url"
	"testing"

	. "github.com/smartystreets/goconvey/convey"

	"github.com/authgear/authgear-server/pkg/api/model"
	"github.com/authgear/authgear-server/pkg/lib/config"
)

func TestConsentViewModelForClient(t *testing.T) {
	Convey("consentViewModelForClient", t, func() {
		Convey("static client with client_name: ClientName is the client_name", func() {
			client := &config.OAuthClientConfig{
				ClientID: "static-client",
				Name:     "Foo",
			}
			vm := consentViewModelForClient(client, nil)
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
			vm := consentViewModelForClient(client, nil)
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
			vm := consentViewModelForClient(client, nil)
			So(vm.ClientName, ShouldEqual, "Client https://mcp-client.example.com/oauth/client-metadata.json")
		})

		Convey("PolicyURI and TOSURI pass through unchanged", func() {
			client := &config.OAuthClientConfig{
				ClientID:  "static-client",
				Name:      "Foo",
				PolicyURI: "https://example.com/policy",
				TOSURI:    "https://example.com/tos",
			}
			vm := consentViewModelForClient(client, nil)
			So(vm.ClientPolicyURI, ShouldEqual, "https://example.com/policy")
			So(vm.ClientTOSURI, ShouldEqual, "https://example.com/tos")
		})

		Convey("a CIMD client: ClientIDHostname is the client_id URL's hostname", func() {
			client := &config.OAuthClientConfig{
				ClientID:      "https://mcp-client.example.com/oauth/client-metadata.json",
				Name:          "Example MCP Client",
				DynamicSource: model.OAuthClientSourceCIMD,
			}
			vm := consentViewModelForClient(client, nil)
			So(vm.ClientName, ShouldEqual, "Example MCP Client")
			So(vm.ClientIDHostname, ShouldEqual, "mcp-client.example.com")
		})

		Convey("a CIMD client with no client_name: ClientName is the DisplayName() fallback, ClientIDHostname is still set", func() {
			client := &config.OAuthClientConfig{
				ClientID:      "https://mcp-client.example.com/oauth/client-metadata.json",
				Name:          "Client https://mcp-client.example.com/oauth/client-metadata.json",
				DynamicSource: model.OAuthClientSourceCIMD,
			}
			vm := consentViewModelForClient(client, nil)
			So(vm.ClientName, ShouldEqual, "Client https://mcp-client.example.com/oauth/client-metadata.json")
			So(vm.ClientIDHostname, ShouldEqual, "mcp-client.example.com")
		})

		Convey("a STATIC client whose client_id happens to be an https:// URL: ClientIDHostname is empty -- the pre-registration pattern must not pick up the CIMD banner", func() {
			client := &config.OAuthClientConfig{
				ClientID: "https://pinned.example.com/x",
				Name:     "Pinned Client",
				// DynamicSource left unset ("") -- this is a STATIC client
				// whose client_id happens to be shaped like a URL, not a
				// CIMD-resolved one. A naive strings.HasPrefix(clientID,
				// "https://") check would get this wrong.
			}
			vm := consentViewModelForClient(client, nil)
			So(vm.ClientIDHostname, ShouldBeEmpty)
		})

		Convey("a DCR client: ClientIDHostname is empty -- the hostname banner is CIMD-specific", func() {
			client := &config.OAuthClientConfig{
				ClientID:      "dcrc_abc123",
				Name:          "Some DCR Client",
				DynamicSource: model.OAuthClientSourceDCR,
			}
			vm := consentViewModelForClient(client, nil)
			So(vm.ClientIDHostname, ShouldBeEmpty)
		})

		Convey("a CIMD client with logo_uri, no endpoints supplied: ClientLogoURI is the raw logo_uri", func() {
			client := &config.OAuthClientConfig{
				ClientID:      "https://mcp-client.example.com/oauth/client-metadata.json",
				Name:          "Example MCP Client",
				DynamicSource: model.OAuthClientSourceCIMD,
				LogoURI:       "https://mcp-client.example.com/logo.png",
			}
			vm := consentViewModelForClient(client, nil)
			So(vm.ClientLogoURI, ShouldEqual, "https://mcp-client.example.com/logo.png")
		})

		Convey("a client with no logo_uri: ClientLogoURI is empty", func() {
			client := &config.OAuthClientConfig{
				ClientID: "static-client",
				Name:     "Foo",
			}
			vm := consentViewModelForClient(client, nil)
			So(vm.ClientLogoURI, ShouldBeEmpty)
		})

		Convey("a CIMD client with logo_uri, endpoints supplied: ClientLogoURI is the Authgear proxy URL -- the browser's actual connection target (the URL's host) is Authgear's origin, not the client's", func() {
			client := &config.OAuthClientConfig{
				ClientID:      "https://mcp-client.example.com/oauth/client-metadata.json",
				Name:          "Example MCP Client",
				DynamicSource: model.OAuthClientSourceCIMD,
				IsDynamic:     true,
				LogoURI:       "https://mcp-client.example.com/logo.png",
			}
			vm := consentViewModelForClient(client, &stubConsentClientLogoEndpoint{})
			So(vm.ClientLogoURI, ShouldStartWith, "https://authgear.example.com/_internals/client_logo?")
			parsed, err := url.Parse(vm.ClientLogoURI)
			So(err, ShouldBeNil)
			// The client's own hostname still appears TEXTUALLY inside the
			// client_id query parameter (it names that host) -- what
			// actually matters, and what a plain substring check would
			// miss, is that the URL's own authority (what the browser
			// connects to) is Authgear's, not the client's.
			So(parsed.Host, ShouldEqual, "authgear.example.com")
		})

		Convey("a DCR client with logo_uri, endpoints supplied: ClientLogoURI is also the proxy URL -- the proxy covers both dynamic sources", func() {
			client := &config.OAuthClientConfig{
				ClientID:      "dcrc_abc123",
				Name:          "Some DCR Client",
				DynamicSource: model.OAuthClientSourceDCR,
				IsDynamic:     true,
				LogoURI:       "https://dcr-client.example.com/logo.png",
			}
			vm := consentViewModelForClient(client, &stubConsentClientLogoEndpoint{})
			So(vm.ClientLogoURI, ShouldStartWith, "https://authgear.example.com/_internals/client_logo?")
			// The proxy URL is keyed on ClientID, not on LogoURI -- so the
			// client's own logo host (unlike a CIMD client_id, which is
			// itself a URL) never appears in it at all.
			So(vm.ClientLogoURI, ShouldNotContainSubstring, "dcr-client.example.com")
		})

		Convey("a STATIC client with logo_uri, endpoints supplied: ClientLogoURI is still the raw logo_uri -- the proxy is dynamic-clients only", func() {
			client := &config.OAuthClientConfig{
				ClientID: "static-client",
				Name:     "Foo",
				LogoURI:  "https://static-client.example.com/logo.png",
				// IsDynamic left false -- a static client.
			}
			vm := consentViewModelForClient(client, &stubConsentClientLogoEndpoint{})
			So(vm.ClientLogoURI, ShouldEqual, "https://static-client.example.com/logo.png")
		})

		Convey("the client_id in the proxy URL round-trips exactly, even when it contains ':', '/' and a query string", func() {
			clientID := "https://mcp-client.example.com:8443/a/b/oauth/client-metadata.json?tenant=acme&v=2"
			client := &config.OAuthClientConfig{
				ClientID:      clientID,
				Name:          "Example MCP Client",
				DynamicSource: model.OAuthClientSourceCIMD,
				IsDynamic:     true,
				LogoURI:       "https://mcp-client.example.com/logo.png",
			}
			vm := consentViewModelForClient(client, &stubConsentClientLogoEndpoint{})
			parsed, err := url.Parse(vm.ClientLogoURI)
			So(err, ShouldBeNil)
			So(parsed.Query().Get("client_id"), ShouldEqual, clientID)
		})
	})
}

// stubConsentClientLogoEndpoint mirrors *endpoints.Endpoints.ClientLogoURL
// closely enough to test the round trip (percent-encoding a client_id
// containing reserved URL characters) without depending on that package.
type stubConsentClientLogoEndpoint struct{}

func (stubConsentClientLogoEndpoint) ClientLogoURL(clientID string) *url.URL {
	u, _ := url.Parse("https://authgear.example.com/_internals/client_logo")
	q := url.Values{}
	q.Set("client_id", clientID)
	u.RawQuery = q.Encode()
	return u
}
