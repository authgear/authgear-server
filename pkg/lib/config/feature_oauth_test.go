package config_test

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"

	"github.com/authgear/authgear-server/pkg/lib/config"
)

func TestOAuthClientIDMetadataDocumentFeatureConfigNilSafety(t *testing.T) {
	Convey("(*OAuthFeatureConfig)(nil) chain never panics", t, func() {
		var c *config.OAuthFeatureConfig
		So(func() {
			So(c.GetClientIDMetadataDocument().IsInsecureHTTPAllowed(), ShouldBeFalse)
			So(c.GetClientIDMetadataDocument().IsInsecureFetchAddressAllowed(), ShouldBeFalse)
		}, ShouldNotPanic)
	})

	Convey("(*OAuthClientIDMetadataDocumentFeatureConfig)(nil) is safe", t, func() {
		var c *config.OAuthClientIDMetadataDocumentFeatureConfig
		So(c.IsInsecureHTTPAllowed(), ShouldBeFalse)
		So(c.IsInsecureFetchAddressAllowed(), ShouldBeFalse)
	})
}
