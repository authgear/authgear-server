package samlbinding_test

import (
	"bytes"
	"encoding/base64"
	"io"
	"net/http"
	"net/url"
	"testing"

	. "github.com/smartystreets/goconvey/convey"

	"github.com/authgear/authgear-server/pkg/lib/saml/samlbinding"
	"github.com/authgear/authgear-server/pkg/lib/saml/samlprotocol"
)

func TestSAMLBindingHTTPPost(t *testing.T) {

	Convey("SAMLBindingHTTPPostParser", t, func() {
		Convey("success", func() {
			req := &http.Request{}
			req.URL = &url.URL{}
			q := url.Values{}
			relayState := "testrelaystate"
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
			base64EncodedRequest := base64.StdEncoding.EncodeToString([]byte(samlRequestXML))

			samlRequest := base64EncodedRequest
			q.Add("RelayState", relayState)
			q.Add("SAMLRequest", samlRequest)
			bodyStr := q.Encode()
			header := http.Header{}
			header.Set("Content-Type", "application/x-www-form-urlencoded")
			req.Method = "POST"
			req.Header = header
			req.Body = io.NopCloser(bytes.NewReader([]byte(bodyStr)))
			result, err := samlbinding.SAMLBindingHTTPPostParse(req)
			So(err, ShouldBeNil)

			So(result.RelayState, ShouldEqual, relayState)
			authnRequest, err := samlprotocol.ParseAuthnRequest([]byte(result.SAMLRequestXML))
			So(err, ShouldBeNil)
			So(authnRequest.ProtocolBinding, ShouldEqual, "urn:oasis:names:tc:SAML:2.0:bindings:HTTP-POST")
			So(*authnRequest.ForceAuthn, ShouldBeFalse)
			So(authnRequest.AssertionConsumerServiceURL, ShouldEqual, "http://example.com/acs")
		})
	})
}
