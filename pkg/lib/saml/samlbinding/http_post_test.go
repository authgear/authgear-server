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

	Convey("SAMLBindingHTTPPostParseRequest", t, func() {
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
			result, err := samlbinding.SAMLBindingHTTPPostParseRequest(req)
			So(err, ShouldBeNil)

			So(result.RelayState, ShouldEqual, relayState)
			authnRequest, err := samlprotocol.ParseAuthnRequest([]byte(result.SAMLRequestXML))
			So(err, ShouldBeNil)
			So(authnRequest.ProtocolBinding, ShouldEqual, "urn:oasis:names:tc:SAML:2.0:bindings:HTTP-POST")
			So(*authnRequest.ForceAuthn, ShouldBeFalse)
			So(authnRequest.AssertionConsumerServiceURL, ShouldEqual, "http://example.com/acs")
		})
	})

	Convey("SAMLBindingHTTPPostParseResponse", t, func() {
		Convey("success", func() {
			req := &http.Request{}
			req.URL = &url.URL{}
			q := url.Values{}
			relayState := "testrelaystate"
			samlResponseXML := `
				<samlp:LogoutResponse xmlns:saml="urn:oasis:names:tc:SAML:2.0:assertion"
					xmlns:samlp="urn:oasis:names:tc:SAML:2.0:protocol"
					ID="samlresponse_MbYfFkBWyYecEYOUutsVPlZAHYu1LAQz"
					InResponseTo="ONELOGIN_a0ed6cdf159a975f74e59a6659bc7df949512be1" Version="2.0"
					IssueInstant="2024-09-23T08:35:55.167Z"
					Destination="https://authgear-demo-saml-sp.pandawork.com/?sls">
					<saml:Issuer Format="urn:oasis:names:tc:SAML:2.0:nameid-format:entity">urn:my-app.localhost</saml:Issuer>
					<samlp:Status>
						<samlp:StatusCode Value="urn:oasis:names:tc:SAML:2.0:status:Success" />
					</samlp:Status>
				</samlp:LogoutResponse>

			`
			base64EncodedResponse := base64.StdEncoding.EncodeToString([]byte(samlResponseXML))

			samlResponse := base64EncodedResponse
			q.Add("RelayState", relayState)
			q.Add("SAMLResponse", samlResponse)
			bodyStr := q.Encode()
			header := http.Header{}
			header.Set("Content-Type", "application/x-www-form-urlencoded")
			req.Method = "POST"
			req.Header = header
			req.Body = io.NopCloser(bytes.NewReader([]byte(bodyStr)))
			result, err := samlbinding.SAMLBindingHTTPPostParseResponse(req)
			So(err, ShouldBeNil)

			So(result.RelayState, ShouldEqual, relayState)
			logoutResponse, err := samlprotocol.ParseLogoutResponse([]byte(result.SAMLResponseXML))
			So(err, ShouldBeNil)
			So(logoutResponse.ID, ShouldEqual, "samlresponse_MbYfFkBWyYecEYOUutsVPlZAHYu1LAQz")
			So(logoutResponse.Status.StatusCode.Value, ShouldEqual, "urn:oasis:names:tc:SAML:2.0:status:Success")
		})
	})
}
