package samlbinding_test

import (
	"net/http"
	"net/url"
	"testing"

	. "github.com/smartystreets/goconvey/convey"

	"github.com/authgear/authgear-server/pkg/lib/saml/samlbinding"
)

func TestSAMLBindingHTTPRedirect(t *testing.T) {

	Convey("SAMLBindingHTTPRedirectParser", t, func() {
		parser := &samlbinding.SAMLBindingHTTPRedirectParser{}
		Convey("success", func() {
			req := &http.Request{}
			req.URL = &url.URL{}
			q := url.Values{}
			relayState := "testrelaystate"
			/*
				<samlp:AuthnRequest
					xmlns:samlp="urn:oasis:names:tc:SAML:2.0:protocol"
					xmlns:saml="urn:oasis:names:tc:SAML:2.0:assertion"
					ForceAuthn="false"
					ID="a506045340df4e2290640daf189f4b5177f691823"
					IssueInstant="2024-08-16T08:25:59Z"
					Destination="http://localhost:3000/saml2/login/sp1"
					AssertionConsumerServiceURL="http://example.com/acs"
					ProtocolBinding="urn:oasis:names:tc:SAML:2.0:bindings:HTTP-POST"
					Version="2.0"
					><saml:Issuer>IAMShowcase</saml:Issuer></samlp:AuthnRequest
				>
			*/
			samlRequest := "fZFRT8IwFIXfTfwPZO+wrmxjaxjJlBhJIBI2ffCtlDu2pGtnb6f8fEdRoz7w1t57vttzbufIW9mxvLe12sFbD2hvb0ajUysVMtfLvN4opjk2yBRvAZkVrMg3a0YnhHVGWy209P5S1yGOCMY2WjnqQRsBzkDmVVwiuOpqmXk8IjEJo2lIDlUIlKYkHo68CpK0CvdRMJtVcRokdHohEHtYKbRc2cyjhIZjkoyDuCQJoxGL0lcnWw4RG8XPz2debW3HfF9qwWWt0bIpIcQ/J6BD8dgoH7vAYfm353utsG/BFGDeGwHPu/XPGDjxtpMwEbr1uUDHbb8WdNeoQ6OO1xezv4iQPZbldrx9Kko34wUMOruDxhUW87NF5hKbxSrfFLX+EBxh7v9uXG7/f3fxCQ=="
			q.Add("RelayState", relayState)
			q.Add("SAMLRequest", samlRequest)
			req.URL.RawQuery = q.Encode()
			result, err := parser.Parse(req)
			So(err, ShouldBeNil)

			So(result.RelayState, ShouldEqual, relayState)
			So(result.AuthnRequest, ShouldNotBeNil)
			So(result.AuthnRequest.ProtocolBinding, ShouldEqual, "urn:oasis:names:tc:SAML:2.0:bindings:HTTP-POST")
			So(*result.AuthnRequest.ForceAuthn, ShouldBeFalse)
			So(result.AuthnRequest.AssertionConsumerServiceURL, ShouldEqual, "http://example.com/acs")
		})
	})
}
