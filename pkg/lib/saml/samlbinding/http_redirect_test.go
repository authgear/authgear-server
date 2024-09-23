package samlbinding_test

import (
	"bytes"
	"compress/flate"
	"encoding/base64"
	"net/http"
	"net/url"
	"testing"

	. "github.com/smartystreets/goconvey/convey"

	"github.com/authgear/authgear-server/pkg/lib/saml/samlbinding"
	"github.com/authgear/authgear-server/pkg/lib/saml/samlprotocol"
)

func TestSAMLBindingHTTPRedirect(t *testing.T) {

	Convey("SAMLBindingHTTPRedirectParser", t, func() {
		Convey("success", func() {
			req := &http.Request{}
			req.URL = &url.URL{}
			q := url.Values{}
			relayState := "testrelaystate"
			sigAlg := "http://www.w3.org/2001/04/xmldsig-more#rsa-sha256"
			samlRequestXML := `
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
			`
			compressedRequestBuffer := &bytes.Buffer{}
			writer, err := flate.NewWriter(compressedRequestBuffer, 9)
			So(err, ShouldBeNil)
			_, err = writer.Write([]byte(samlRequestXML))
			So(err, ShouldBeNil)
			err = writer.Close()
			So(err, ShouldBeNil)
			base64EncodedRequest := base64.StdEncoding.EncodeToString(compressedRequestBuffer.Bytes())
			q.Add("RelayState", relayState)
			q.Add("SigAlg", sigAlg)
			q.Add("SAMLRequest", base64EncodedRequest)
			req.URL.RawQuery = q.Encode()
			result, err := samlbinding.SAMLBindingHTTPRedirectParseRequest(req)
			So(err, ShouldBeNil)

			So(result.SAMLRequest, ShouldEqual, base64EncodedRequest)
			So(result.SigAlg, ShouldEqual, sigAlg)
			So(result.RelayState, ShouldEqual, relayState)
			authnRequest, err := samlprotocol.ParseAuthnRequest([]byte(result.SAMLRequestXML))
			So(err, ShouldBeNil)
			So(authnRequest, ShouldNotBeNil)
			So(authnRequest.ProtocolBinding, ShouldEqual, "urn:oasis:names:tc:SAML:2.0:bindings:HTTP-POST")
			So(*authnRequest.ForceAuthn, ShouldBeFalse)
			So(authnRequest.AssertionConsumerServiceURL, ShouldEqual, "http://example.com/acs")
		})
	})
}
