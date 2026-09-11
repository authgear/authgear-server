package dcr_test

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"

	"github.com/authgear/authgear-server/pkg/lib/dcr"
)

func TestValidateAndNormalize(t *testing.T) {
	Convey("ValidateAndNormalize", t, func() {
		validReq := func() *dcr.RegistrationRequest {
			return &dcr.RegistrationRequest{
				RedirectURIs: []string{"https://example.com/callback"},
			}
		}

		Convey("defaults grant_types, response_types, application_type when omitted", func() {
			r, err := dcr.ValidateAndNormalize(validReq())
			So(err, ShouldBeNil)
			So(r.GrantTypes, ShouldResemble, []string{"authorization_code", "refresh_token"})
			So(r.ResponseTypes, ShouldResemble, []string{"code"})
			So(r.ApplicationType, ShouldEqual, "web")
		})

		Convey("missing redirect_uris", func() {
			req := validReq()
			req.RedirectURIs = nil
			_, err := dcr.ValidateAndNormalize(req)
			So(err, ShouldEqual, dcr.ErrDCRRedirectURIsMissing)
		})

		Convey("redirect_uri with a fragment component", func() {
			req := validReq()
			req.RedirectURIs = []string{"https://example.com/callback#section"}
			_, err := dcr.ValidateAndNormalize(req)
			So(err, ShouldEqual, dcr.ErrDCRRedirectURIInvalid)
		})

		Convey("relative redirect_uri is not absolute", func() {
			req := validReq()
			req.RedirectURIs = []string{"/callback"}
			_, err := dcr.ValidateAndNormalize(req)
			So(err, ShouldEqual, dcr.ErrDCRRedirectURIInvalid)
		})

		Convey("web application_type rejects http scheme", func() {
			req := validReq()
			req.RedirectURIs = []string{"http://example.com/callback"}
			_, err := dcr.ValidateAndNormalize(req)
			So(err, ShouldEqual, dcr.ErrDCRRedirectURIInvalid)
		})

		Convey("web application_type rejects http://localhost", func() {
			req := validReq()
			req.RedirectURIs = []string{"http://localhost/callback"}
			_, err := dcr.ValidateAndNormalize(req)
			So(err, ShouldEqual, dcr.ErrDCRRedirectURIInvalid)
		})

		Convey("native application_type accepts a custom URI scheme", func() {
			req := validReq()
			req.ApplicationType = new("native")
			req.RedirectURIs = []string{"com.example.app://callback"}
			r, err := dcr.ValidateAndNormalize(req)
			So(err, ShouldBeNil)
			So(r.ApplicationType, ShouldEqual, "native")
		})

		Convey("native application_type accepts http://localhost", func() {
			req := validReq()
			req.ApplicationType = new("native")
			req.RedirectURIs = []string{"http://localhost/callback"}
			_, err := dcr.ValidateAndNormalize(req)
			So(err, ShouldBeNil)
		})

		Convey("native application_type accepts http://127.0.0.1 loopback, any port", func() {
			req := validReq()
			req.ApplicationType = new("native")
			req.RedirectURIs = []string{"http://127.0.0.1:60327/callback/abc"}
			_, err := dcr.ValidateAndNormalize(req)
			So(err, ShouldBeNil)
		})

		Convey("native application_type accepts http://[::1] loopback", func() {
			req := validReq()
			req.ApplicationType = new("native")
			req.RedirectURIs = []string{"http://[::1]:3000/callback"}
			_, err := dcr.ValidateAndNormalize(req)
			So(err, ShouldBeNil)
		})

		Convey("native application_type rejects non-loopback IP", func() {
			req := validReq()
			req.ApplicationType = new("native")
			req.RedirectURIs = []string{"http://127.0.0.2/callback"}
			_, err := dcr.ValidateAndNormalize(req)
			So(err, ShouldEqual, dcr.ErrDCRRedirectURIInvalid)
		})

		Convey("native application_type rejects https", func() {
			req := validReq()
			req.ApplicationType = new("native")
			req.RedirectURIs = []string{"https://example.com/callback"}
			_, err := dcr.ValidateAndNormalize(req)
			So(err, ShouldEqual, dcr.ErrDCRRedirectURIInvalid)
		})

		Convey("native application_type rejects non-localhost http", func() {
			req := validReq()
			req.ApplicationType = new("native")
			req.RedirectURIs = []string{"http://example.com/callback"}
			_, err := dcr.ValidateAndNormalize(req)
			So(err, ShouldEqual, dcr.ErrDCRRedirectURIInvalid)
		})

		Convey("grant_types with no supported entry is rejected", func() {
			req := validReq()
			req.GrantTypes = []string{"implicit"}
			_, err := dcr.ValidateAndNormalize(req)
			So(err, ShouldEqual, dcr.ErrDCRGrantTypeUnsupported)
		})

		Convey("an unimplemented grant_type is dropped, not fatal", func() {
			req := validReq()
			req.GrantTypes = []string{"authorization_code", "refresh_token", "urn:ietf:params:oauth:grant-type:jwt-bearer"}
			r, err := dcr.ValidateAndNormalize(req)
			So(err, ShouldBeNil)
			So(r.GrantTypes, ShouldResemble, []string{"authorization_code", "refresh_token"})
		})

		Convey("dropping preserves the order of what remains", func() {
			req := validReq()
			req.GrantTypes = []string{"refresh_token", "implicit", "authorization_code"}
			r, err := dcr.ValidateAndNormalize(req)
			So(err, ShouldBeNil)
			So(r.GrantTypes, ShouldResemble, []string{"refresh_token", "authorization_code"})
		})

		// Nothing to drop, so the consistency rule decides it.
		Convey("an empty grant_types is left to the consistency rule", func() {
			req := validReq()
			req.GrantTypes = []string{}
			_, err := dcr.ValidateAndNormalize(req)
			So(err, ShouldEqual, dcr.ErrDCRResponseTypeInconsistent)

			req.ResponseTypes = []string{}
			r, err := dcr.ValidateAndNormalize(req)
			So(err, ShouldBeNil)
			So(r.GrantTypes, ShouldBeEmpty)
		})

		Convey("unsupported response_type", func() {
			req := validReq()
			req.ResponseTypes = []string{"token"}
			_, err := dcr.ValidateAndNormalize(req)
			So(err, ShouldEqual, dcr.ErrDCRResponseTypeInconsistent)
		})

		Convey("response_types code without authorization_code in grant_types", func() {
			req := validReq()
			req.GrantTypes = []string{"refresh_token"}
			req.ResponseTypes = []string{"code"}
			_, err := dcr.ValidateAndNormalize(req)
			So(err, ShouldEqual, dcr.ErrDCRResponseTypeInconsistent)
		})

		Convey("grant_types authorization_code without response_types code", func() {
			req := validReq()
			req.GrantTypes = []string{"authorization_code"}
			req.ResponseTypes = []string{}
			_, err := dcr.ValidateAndNormalize(req)
			So(err, ShouldEqual, dcr.ErrDCRResponseTypeInconsistent)
		})

		Convey("unsupported application_type", func() {
			req := validReq()
			req.ApplicationType = new("m2m")
			_, err := dcr.ValidateAndNormalize(req)
			So(err, ShouldEqual, dcr.ErrDCRApplicationTypeUnsupported)
		})

		Convey("non-https logo_uri", func() {
			req := validReq()
			req.LogoURI = new("http://example.com/logo.png")
			_, err := dcr.ValidateAndNormalize(req)
			So(err, ShouldEqual, dcr.ErrDCRURIFieldNotHTTPS)
		})

		Convey("non-https client_uri", func() {
			req := validReq()
			req.ClientURI = new("http://example.com")
			_, err := dcr.ValidateAndNormalize(req)
			So(err, ShouldEqual, dcr.ErrDCRURIFieldNotHTTPS)
		})

		Convey("non-https tos_uri", func() {
			req := validReq()
			req.TOSURI = new("http://example.com/tos")
			_, err := dcr.ValidateAndNormalize(req)
			So(err, ShouldEqual, dcr.ErrDCRURIFieldNotHTTPS)
		})

		Convey("non-https policy_uri", func() {
			req := validReq()
			req.PolicyURI = new("http://example.com/policy")
			_, err := dcr.ValidateAndNormalize(req)
			So(err, ShouldEqual, dcr.ErrDCRURIFieldNotHTTPS)
		})

		Convey("https uri fields are accepted", func() {
			req := validReq()
			req.LogoURI = new("https://example.com/logo.png")
			req.ClientURI = new("https://example.com")
			req.TOSURI = new("https://example.com/tos")
			req.PolicyURI = new("https://example.com/policy")
			r, err := dcr.ValidateAndNormalize(req)
			So(err, ShouldBeNil)
			So(*r.LogoURI, ShouldEqual, "https://example.com/logo.png")
		})
	})
}
