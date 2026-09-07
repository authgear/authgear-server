package oauth_test

import (
	"net/url"
	"testing"

	. "github.com/smartystreets/goconvey/convey"

	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/lib/oauth"
)

type metadataTestEndpoints struct{}

func (metadataTestEndpoints) AuthorizeEndpointURL() *url.URL {
	return &url.URL{Scheme: "https", Host: "example.com", Path: "/oauth2/authorize"}
}
func (metadataTestEndpoints) ConsentEndpointURL() *url.URL {
	return &url.URL{Scheme: "https", Host: "example.com", Path: "/oauth2/consent"}
}
func (metadataTestEndpoints) TokenEndpointURL() *url.URL {
	return &url.URL{Scheme: "https", Host: "example.com", Path: "/oauth2/token"}
}
func (metadataTestEndpoints) RevokeEndpointURL() *url.URL {
	return &url.URL{Scheme: "https", Host: "example.com", Path: "/oauth2/revoke"}
}
func (metadataTestEndpoints) RegistrationEndpointURL() *url.URL {
	return &url.URL{Scheme: "https", Host: "example.com", Path: "/oauth2/register"}
}

func TestMetadataProviderPopulateMetadata(t *testing.T) {
	Convey("MetadataProvider.PopulateMetadata", t, func() {
		Convey("CIMD disabled: client_id_metadata_document_supported is absent", func() {
			p := &oauth.MetadataProvider{
				Endpoints: metadataTestEndpoints{},
				OAuthConfig: &config.OAuthConfig{
					ClientIDMetadataDocument: &config.OAuthClientIDMetadataDocumentConfig{Enabled: false},
				},
			}
			meta := map[string]any{}
			p.PopulateMetadata(meta)
			_, ok := meta["client_id_metadata_document_supported"]
			So(ok, ShouldBeFalse)
		})

		Convey("CIMD enabled: client_id_metadata_document_supported is true", func() {
			p := &oauth.MetadataProvider{
				Endpoints: metadataTestEndpoints{},
				OAuthConfig: &config.OAuthConfig{
					ClientIDMetadataDocument: &config.OAuthClientIDMetadataDocumentConfig{Enabled: true},
				},
			}
			meta := map[string]any{}
			p.PopulateMetadata(meta)
			So(meta["client_id_metadata_document_supported"], ShouldEqual, true)
		})

		Convey("registration_endpoint behavior is unchanged regardless of CIMD", func() {
			p := &oauth.MetadataProvider{
				Endpoints: metadataTestEndpoints{},
				OAuthConfig: &config.OAuthConfig{
					DynamicClientRegistration: &config.OAuthDynamicClientRegistrationConfig{Enabled: true},
					ClientIDMetadataDocument:  &config.OAuthClientIDMetadataDocumentConfig{Enabled: true},
				},
			}
			meta := map[string]any{}
			p.PopulateMetadata(meta)
			So(meta["registration_endpoint"], ShouldEqual, "https://example.com/oauth2/register")

			p2 := &oauth.MetadataProvider{
				Endpoints: metadataTestEndpoints{},
				OAuthConfig: &config.OAuthConfig{
					DynamicClientRegistration: &config.OAuthDynamicClientRegistrationConfig{Enabled: false},
					ClientIDMetadataDocument:  &config.OAuthClientIDMetadataDocumentConfig{Enabled: true},
				},
			}
			meta2 := map[string]any{}
			p2.PopulateMetadata(meta2)
			_, ok := meta2["registration_endpoint"]
			So(ok, ShouldBeFalse)
		})
	})
}
