package cimd_test

import (
	"encoding/json"
	"errors"
	"testing"

	. "github.com/smartystreets/goconvey/convey"

	"github.com/authgear/authgear-server/pkg/api/apierrors"
	"github.com/authgear/authgear-server/pkg/lib/cimd"
)

const mcpExampleClientID = "https://mcp-client.example.com/oauth/client-metadata.json"

// mcpExampleDocument is the MCP Authorization spec's own example CIMD
// document (docs/specs/cimd.md § UC1 Step 2), verbatim.
var mcpExampleDocument = []byte(`{
  "client_id": "https://mcp-client.example.com/oauth/client-metadata.json",
  "client_name": "Example MCP Client",
  "redirect_uris": [
    "http://127.0.0.1:3000/callback",
    "http://localhost:3000/callback"
  ],
  "grant_types": ["authorization_code", "refresh_token"],
  "response_types": ["code"],
  "token_endpoint_auth_method": "none"
}`)

func TestParseAndValidate(t *testing.T) {
	Convey("ParseAndValidate", t, func() {
		Convey("the MCP spec's own example document is valid", func() {
			doc, err := cimd.ParseAndValidate(mcpExampleClientID, mcpExampleDocument, false)
			So(err, ShouldBeNil)
			So(doc.RedirectURIs, ShouldResemble, []string{"http://127.0.0.1:3000/callback", "http://localhost:3000/callback"})
			So(doc.ApplicationType, ShouldEqual, "web") // absent -> default, despite loopback redirect_uris
			So(doc.GrantTypes, ShouldResemble, []string{"authorization_code", "refresh_token"})
			So(doc.ResponseTypes, ShouldResemble, []string{"code"})
			So(*doc.ClientName, ShouldEqual, "Example MCP Client")
		})

		Convey("client_id", func() {
			Convey("missing is rejected", func() {
				doc := map[string]any{
					"redirect_uris": []string{"https://example.com/cb"},
				}
				_, err := parse(t, mcpExampleClientID, doc)
				So(err, ShouldEqual, cimd.ErrDocumentClientIDMismatch)
			})

			Convey("differing only by a trailing slash is rejected", func() {
				doc := validDoc()
				doc["client_id"] = mcpExampleClientID + "/"
				_, err := parse(t, mcpExampleClientID, doc)
				So(err, ShouldEqual, cimd.ErrDocumentClientIDMismatch)
			})

			Convey("differing only by scheme case is rejected", func() {
				doc := validDoc()
				doc["client_id"] = "HTTPS://mcp-client.example.com/oauth/client-metadata.json"
				_, err := parse(t, mcpExampleClientID, doc)
				So(err, ShouldEqual, cimd.ErrDocumentClientIDMismatch)
			})

			Convey("differing only by host case is rejected", func() {
				doc := validDoc()
				doc["client_id"] = "https://MCP-client.example.com/oauth/client-metadata.json"
				_, err := parse(t, mcpExampleClientID, doc)
				So(err, ShouldEqual, cimd.ErrDocumentClientIDMismatch)
			})
		})

		Convey("token_endpoint_auth_method", func() {
			Convey("none is ok", func() {
				doc := validDoc()
				doc["token_endpoint_auth_method"] = "none"
				_, err := parse(t, mcpExampleClientID, doc)
				So(err, ShouldBeNil)
			})

			Convey("absent is ok", func() {
				doc := validDoc()
				delete(doc, "token_endpoint_auth_method")
				_, err := parse(t, mcpExampleClientID, doc)
				So(err, ShouldBeNil)
			})

			Convey("client_secret_post is rejected", func() {
				doc := validDoc()
				doc["token_endpoint_auth_method"] = "client_secret_post"
				_, err := parse(t, mcpExampleClientID, doc)
				So(err, ShouldEqual, cimd.ErrDocumentTokenEndpointAuthMethodNotAccepted)
			})

			Convey("private_key_jwt is rejected", func() {
				doc := validDoc()
				doc["token_endpoint_auth_method"] = "private_key_jwt"
				_, err := parse(t, mcpExampleClientID, doc)
				So(err, ShouldEqual, cimd.ErrDocumentTokenEndpointAuthMethodNotAccepted)
			})
		})

		Convey("redirect_uris", func() {
			Convey("missing is rejected", func() {
				doc := validDoc()
				delete(doc, "redirect_uris")
				_, err := parse(t, mcpExampleClientID, doc)
				So(err, ShouldEqual, cimd.ErrDocumentRedirectURIsMissing)
			})

			Convey("empty is rejected", func() {
				doc := validDoc()
				doc["redirect_uris"] = []string{}
				_, err := parse(t, mcpExampleClientID, doc)
				So(err, ShouldEqual, cimd.ErrDocumentRedirectURIsMissing)
			})

			Convey("a non-loopback http redirect_uri is rejected", func() {
				doc := validDoc()
				doc["redirect_uris"] = []string{"http://evil.com/cb"}
				_, err := parse(t, mcpExampleClientID, doc)
				So(err, ShouldEqual, cimd.ErrDocumentRedirectURIInvalid)
			})

			acceptedURIs := []string{
				"http://localhost:1/cb",
				"http://127.0.0.1:65535/cb",
				"http://[::1]:3000/cb",
				"com.example.app:/cb",
				"myapp://cb",
				"https://example.com/cb",
			}
			for _, ru := range acceptedURIs {
				ru := ru
				Convey("accepted: "+ru, func() {
					doc := validDoc()
					doc["redirect_uris"] = []string{ru}
					_, err := parse(t, mcpExampleClientID, doc)
					So(err, ShouldBeNil)
				})
			}

			Convey("a redirect_uri with a fragment is rejected", func() {
				doc := validDoc()
				doc["redirect_uris"] = []string{"https://x/cb#frag"}
				_, err := parse(t, mcpExampleClientID, doc)
				So(err, ShouldEqual, cimd.ErrDocumentRedirectURIInvalid)
			})
		})

		Convey("application_type", func() {
			Convey("spa is rejected", func() {
				doc := validDoc()
				doc["application_type"] = "spa"
				_, err := parse(t, mcpExampleClientID, doc)
				So(err, ShouldEqual, cimd.ErrDocumentApplicationTypeUnsupported)
			})

			Convey("native is ok", func() {
				doc := validDoc()
				doc["application_type"] = "native"
				_, err := parse(t, mcpExampleClientID, doc)
				So(err, ShouldBeNil)
			})
		})

		Convey("grant_types", func() {
			Convey("client_credentials is rejected", func() {
				doc := validDoc()
				doc["grant_types"] = []string{"client_credentials"}
				_, err := parse(t, mcpExampleClientID, doc)
				So(err, ShouldEqual, cimd.ErrDocumentGrantTypeUnsupported)
			})
		})

		Convey("response_types", func() {
			Convey("token is rejected", func() {
				doc := validDoc()
				doc["response_types"] = []string{"token"}
				_, err := parse(t, mcpExampleClientID, doc)
				So(err, ShouldEqual, cimd.ErrDocumentResponseTypeInconsistent)
			})

			Convey("refresh_token grant with code response_type is rejected as inconsistent", func() {
				doc := validDoc()
				doc["grant_types"] = []string{"refresh_token"}
				doc["response_types"] = []string{"code"}
				_, err := parse(t, mcpExampleClientID, doc)
				So(err, ShouldEqual, cimd.ErrDocumentResponseTypeInconsistent)
			})
		})

		Convey("logo_uri", func() {
			Convey("http is rejected by default", func() {
				doc := validDoc()
				doc["logo_uri"] = "http://x/l.png"
				_, err := parse(t, mcpExampleClientID, doc)
				So(err, ShouldEqual, cimd.ErrDocumentURIFieldNotHTTPS)
			})

			Convey("http is accepted when allowInsecureHTTP is set", func() {
				doc := validDoc()
				doc["logo_uri"] = "http://x/l.png"
				data, err := json.Marshal(doc)
				So(err, ShouldBeNil)
				parsed, err := cimd.ParseAndValidate(mcpExampleClientID, data, true)
				So(err, ShouldBeNil)
				So(*parsed.LogoURI, ShouldEqual, "http://x/l.png")
			})

			Convey("https is accepted", func() {
				doc := validDoc()
				doc["logo_uri"] = "https://x/l.png"
				_, err := parse(t, mcpExampleClientID, doc)
				So(err, ShouldBeNil)
			})
		})

		Convey("rejected/ignored fields are simply dropped, not errors", func() {
			doc := validDoc()
			doc["client_secret"] = "shh"
			doc["client_secret_expires_at"] = 0
			doc["jwks_uri"] = "https://example.com/jwks.json"
			doc["software_statement"] = "eyJhbGciOiJub25lIn0.e30."
			doc["x_whatever"] = "ignored"

			parsed, err := parse(t, mcpExampleClientID, doc)
			So(err, ShouldBeNil)
			So(parsed, ShouldNotBeNil)
			// None of the ignored fields exist on Document at all, so there
			// is nothing further to assert -- their absence from the type
			// itself is the guarantee.
		})

		Convey("malformed bodies all produce an error", func() {
			cases := map[string][]byte{
				"json array":     []byte(`[]`),
				"json string":    []byte(`"str"`),
				"json null":      []byte(`null`),
				"json number":    []byte(`123`),
				"truncated json": []byte(`{"client_id": "https://x`),
				"empty body":     []byte(``),
			}
			for name, body := range cases {
				name, body := name, body
				Convey(name, func() {
					_, err := cimd.ParseAndValidate(mcpExampleClientID, body, false)
					So(err, ShouldBeError)
				})
			}
		})

		Convey("every ErrDocument* sentinel is a catchable apierrors.APIError sharing one Kind", func() {
			doc := validDoc()
			doc["client_id"] = "https://someone-else.example.com/oauth/client-metadata.json"
			_, err := parse(t, mcpExampleClientID, doc)
			So(err, ShouldEqual, cimd.ErrDocumentClientIDMismatch)
			So(apierrors.IsAPIError(err), ShouldBeTrue)
			So(apierrors.IsKind(err, cimd.CIMDDocumentInvalid), ShouldBeTrue)
			So(apierrors.IsAPIErrorWithCondition(err, func(e *apierrors.APIError) bool {
				return e.HasCause("ClientIDMismatch")
			}), ShouldBeTrue)
		})

		Convey("an errors.Join-wrapped ErrDocumentNotJSONObject is still catchable via errors.Is and IsAPIError", func() {
			// Not IsKind here: asAPIError special-cases a wrapped
			// *json.SyntaxError ahead of a wrapped *APIError, so this
			// specific combination resolves to the generic "invalid JSON"
			// Kind rather than CIMDDocumentInvalid -- errors.Is and
			// IsAPIError, which is what this sentinel is actually caught
			// with in practice, still see straight through the join.
			_, err := cimd.ParseAndValidate(mcpExampleClientID, []byte(`not json`), false)
			So(errors.Is(err, cimd.ErrDocumentNotJSONObject), ShouldBeTrue)
			So(apierrors.IsAPIError(err), ShouldBeTrue)
		})
	})
}

// validDoc returns a fresh map[string]any equal in shape to
// mcpExampleDocument, so individual fields can be mutated per test case
// without JSON string surgery.
func validDoc() map[string]any {
	return map[string]any{
		"client_id":                  mcpExampleClientID,
		"client_name":                "Example MCP Client",
		"redirect_uris":              []string{"http://127.0.0.1:3000/callback", "http://localhost:3000/callback"},
		"grant_types":                []string{"authorization_code", "refresh_token"},
		"response_types":             []string{"code"},
		"token_endpoint_auth_method": "none",
	}
}

func parse(t *testing.T, requestURL string, doc map[string]any) (*cimd.Document, error) {
	t.Helper()
	data, err := json.Marshal(doc)
	if err != nil {
		t.Fatal(err)
	}
	return cimd.ParseAndValidate(requestURL, data, false)
}
