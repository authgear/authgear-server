package config_test

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"

	"github.com/authgear/authgear-server/pkg/lib/config"
)

func TestOAuthClientApplicationTypeDynamicThirdParty(t *testing.T) {
	Convey("OAuthClientApplicationTypeDynamicThirdParty", t, func() {
		t := config.OAuthClientApplicationTypeDynamicThirdParty

		// third-party AND public — the one client shape static config never had.
		So(t.IsThirdParty(), ShouldBeTrue)
		So(t.IsConfidential(), ShouldBeFalse)
		So(t.IsPublic(), ShouldBeTrue)
		So(t.IsClientCredentialsFlowAllowed(), ShouldBeFalse)
		So(t.HasFullAccessScope(), ShouldBeFalse)
		So(t.PIIAllowedInIDToken(), ShouldBeTrue)
	})
}
