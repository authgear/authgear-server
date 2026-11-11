package saml

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"

	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/lib/saml/samlprotocol"
)

func TestResolveAudiences(t *testing.T) {
	Convey("resolveAudiences", t, func() {
		const acsURL = "https://sp.example.com/acs"
		const audience = "https://sp.example.com"
		boolPtr := func(b bool) *bool { return &b }
		requesting := func(auds ...string) *samlprotocol.AuthnRequest {
			r := samlprotocol.AudienceRestriction{}
			for _, aud := range auds {
				r.Audience = append(r.Audience, samlprotocol.Audience{Value: aud})
			}
			return &samlprotocol.AuthnRequest{
				Conditions: &samlprotocol.Conditions{
					AudienceRestrictions: []samlprotocol.AudienceRestriction{r},
				},
			}
		}

		Convey("uses the ACS URL when audience is not set", func() {
			sp := &config.SAMLServiceProviderConfig{}
			So(resolveAudiences(sp, acsURL, nil), ShouldResemble, []string{acsURL})
		})

		Convey("includes the ACS URL by default", func() {
			sp := &config.SAMLServiceProviderConfig{Audience: audience}
			So(resolveAudiences(sp, acsURL, nil), ShouldResemble, []string{audience, acsURL})
		})

		Convey("includes the ACS URL when audience_includes_acs_url is true", func() {
			sp := &config.SAMLServiceProviderConfig{Audience: audience, AudienceIncludesACSURL: boolPtr(true)}
			So(resolveAudiences(sp, acsURL, nil), ShouldResemble, []string{audience, acsURL})
		})

		Convey("excludes the ACS URL when audience_includes_acs_url is false", func() {
			sp := &config.SAMLServiceProviderConfig{Audience: audience, AudienceIncludesACSURL: boolPtr(false)}
			So(resolveAudiences(sp, acsURL, nil), ShouldResemble, []string{audience})
		})

		Convey("includes the ACS URL when requested even if audience_includes_acs_url is false", func() {
			sp := &config.SAMLServiceProviderConfig{Audience: audience, AudienceIncludesACSURL: boolPtr(false)}
			So(resolveAudiences(sp, acsURL, requesting(acsURL)), ShouldResemble, []string{audience, acsURL})
		})

		Convey("includes requested audiences", func() {
			sp := &config.SAMLServiceProviderConfig{}
			So(resolveAudiences(sp, acsURL, requesting("https://other.example.com")), ShouldResemble,
				[]string{"https://other.example.com", acsURL})
		})
	})
}
