package dcr_test

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"

	"github.com/authgear/authgear-server/pkg/lib/dcr"
)

func strptr(s string) *string { return &s }

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
			req.ApplicationType = strptr("native")
			req.RedirectURIs = []string{"com.example.app://callback"}
			r, err := dcr.ValidateAndNormalize(req)
			So(err, ShouldBeNil)
			So(r.ApplicationType, ShouldEqual, "native")
		})

		Convey("native application_type accepts http://localhost", func() {
			req := validReq()
			req.ApplicationType = strptr("native")
			req.RedirectURIs = []string{"http://localhost/callback"}
			_, err := dcr.ValidateAndNormalize(req)
			So(err, ShouldBeNil)
		})

		Convey("native application_type rejects https", func() {
			req := validReq()
			req.ApplicationType = strptr("native")
			req.RedirectURIs = []string{"https://example.com/callback"}
			_, err := dcr.ValidateAndNormalize(req)
			So(err, ShouldEqual, dcr.ErrDCRRedirectURIInvalid)
		})

		Convey("native application_type rejects non-localhost http", func() {
			req := validReq()
			req.ApplicationType = strptr("native")
			req.RedirectURIs = []string{"http://example.com/callback"}
			_, err := dcr.ValidateAndNormalize(req)
			So(err, ShouldEqual, dcr.ErrDCRRedirectURIInvalid)
		})

		Convey("token_endpoint_auth_method other than none is rejected", func() {
			req := validReq()
			req.TokenEndpointAuthMethod = strptr("client_secret_post")
			_, err := dcr.ValidateAndNormalize(req)
			So(err, ShouldEqual, dcr.ErrDCRTokenEndpointAuthMethodNotAccepted)
		})

		Convey("token_endpoint_auth_method=none is accepted", func() {
			req := validReq()
			req.TokenEndpointAuthMethod = new("none")
			_, err := dcr.ValidateAndNormalize(req)
			So(err, ShouldBeNil)
		})

		Convey("unsupported grant_type", func() {
			req := validReq()
			req.GrantTypes = []string{"implicit"}
			_, err := dcr.ValidateAndNormalize(req)
			So(err, ShouldEqual, dcr.ErrDCRGrantTypeUnsupported)
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
			req.ApplicationType = strptr("m2m")
			_, err := dcr.ValidateAndNormalize(req)
			So(err, ShouldEqual, dcr.ErrDCRApplicationTypeUnsupported)
		})

		Convey("non-https logo_uri", func() {
			req := validReq()
			req.LogoURI = strptr("http://example.com/logo.png")
			_, err := dcr.ValidateAndNormalize(req)
			So(err, ShouldEqual, dcr.ErrDCRURIFieldNotHTTPS)
		})

		Convey("non-https client_uri", func() {
			req := validReq()
			req.ClientURI = strptr("http://example.com")
			_, err := dcr.ValidateAndNormalize(req)
			So(err, ShouldEqual, dcr.ErrDCRURIFieldNotHTTPS)
		})

		Convey("non-https tos_uri", func() {
			req := validReq()
			req.TOSURI = strptr("http://example.com/tos")
			_, err := dcr.ValidateAndNormalize(req)
			So(err, ShouldEqual, dcr.ErrDCRURIFieldNotHTTPS)
		})

		Convey("non-https policy_uri", func() {
			req := validReq()
			req.PolicyURI = strptr("http://example.com/policy")
			_, err := dcr.ValidateAndNormalize(req)
			So(err, ShouldEqual, dcr.ErrDCRURIFieldNotHTTPS)
		})

		Convey("https uri fields are accepted", func() {
			req := validReq()
			req.LogoURI = strptr("https://example.com/logo.png")
			req.ClientURI = strptr("https://example.com")
			req.TOSURI = strptr("https://example.com/tos")
			req.PolicyURI = strptr("https://example.com/policy")
			r, err := dcr.ValidateAndNormalize(req)
			So(err, ShouldBeNil)
			So(*r.LogoURI, ShouldEqual, "https://example.com/logo.png")
		})
	})
}
