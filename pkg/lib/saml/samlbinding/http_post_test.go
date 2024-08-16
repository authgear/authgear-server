package samlbinding_test

import (
	"bytes"
	"io"
	"net/http"
	"net/url"
	"testing"

	. "github.com/smartystreets/goconvey/convey"

	"github.com/authgear/authgear-server/pkg/lib/saml/samlbinding"
)

func TestSAMLBindingHTTPPost(t *testing.T) {

	Convey("SAMLBindingHTTPPostParser", t, func() {
		parser := &samlbinding.SAMLBindingHTTPPostParser{}
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
			samlRequest := "PHNhbWxwOkF1dGhuUmVxdWVzdA0KCXhtbG5zOnNhbWxwPSJ1cm46b2FzaXM6bmFtZXM6dGM6U0FNTDoyLjA6cHJvdG9jb2wiDQoJeG1sbnM6c2FtbD0idXJuOm9hc2lzOm5hbWVzOnRjOlNBTUw6Mi4wOmFzc2VydGlvbiINCglGb3JjZUF1dGhuPSJmYWxzZSINCglJRD0iYTUwNjA0NTM0MGRmNGUyMjkwNjQwZGFmMTg5ZjRiNTE3N2Y2OTE4MjMiDQoJSXNzdWVJbnN0YW50PSIyMDI0LTA4LTE2VDA4OjI1OjU5WiINCglEZXN0aW5hdGlvbj0iaHR0cDovL2xvY2FsaG9zdDozMDAwL3NhbWwyL2xvZ2luL3NwMSINCglBc3NlcnRpb25Db25zdW1lclNlcnZpY2VVUkw9Imh0dHA6Ly9leGFtcGxlLmNvbS9hY3MiDQoJUHJvdG9jb2xCaW5kaW5nPSJ1cm46b2FzaXM6bmFtZXM6dGM6U0FNTDoyLjA6YmluZGluZ3M6SFRUUC1QT1NUIg0KCVZlcnNpb249IjIuMCINCgk+PHNhbWw6SXNzdWVyPklBTVNob3djYXNlPC9zYW1sOklzc3Vlcj48L3NhbWxwOkF1dGhuUmVxdWVzdA0KPg=="
			q.Add("RelayState", relayState)
			q.Add("SAMLRequest", samlRequest)
			bodyStr := q.Encode()
			header := http.Header{}
			header.Set("Content-Type", "application/x-www-form-urlencoded")
			req.Method = "POST"
			req.Header = header
			req.Body = io.NopCloser(bytes.NewReader([]byte(bodyStr)))
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
