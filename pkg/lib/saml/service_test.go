package saml_test

import (
	"net/url"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	. "github.com/smartystreets/goconvey/convey"

	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/lib/saml"
	"github.com/authgear/authgear-server/pkg/lib/saml/samlerror"
	"github.com/authgear/authgear-server/pkg/lib/saml/samlprotocol"
	"github.com/authgear/authgear-server/pkg/util/clock"

	crewjamsaml "github.com/crewjam/saml"
)

func TestSAMLService(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	clk := clock.NewMockClockAt("2006-01-02T15:04:05Z")
	endpoints := NewMockSAMLEndpoints(ctrl)
	spID := "testsp"
	loginEndpoint, _ := url.Parse("http://idp.local/login")
	endpoints.EXPECT().SAMLLoginURL(spID).AnyTimes().Return(loginEndpoint)
	svc := &saml.Service{
		Clock: clk,
		AppID: config.AppID("test"),
		SAMLEnvironmentConfig: config.SAMLEnvironmentConfig{
			IdPEntityIDTemplate: "urn:{{.app_id}}.localhost",
		},
		SAMLConfig: &config.SAMLConfig{
			ServiceProviders: []*config.SAMLServiceProviderConfig{
				{
					ID:           spID,
					NameIDFormat: config.SAMLNameIDFormatEmailAddress,
					AcsURLs: []string{
						"http://localhost/saml-test",
					},
				},
			},
		},
		SAMLIdpSigningMaterials: nil,
		Endpoints:               endpoints,
	}

	Convey("ValidateAuthnRequest", t, func() {
		makeValidRequest := func() *samlprotocol.AuthnRequest {
			issueInstant, _ := time.Parse(time.RFC3339, "2006-01-02T15:00:05Z")
			nameIDFormat := "urn:oasis:names:tc:SAML:1.1:nameid-format:emailAddress"
			authnRequest := &samlprotocol.AuthnRequest{
				AuthnRequest: crewjamsaml.AuthnRequest{
					ID:              "id_test",
					Destination:     "http://idp.local/login",
					ProtocolBinding: string(samlprotocol.SAMLBindingHTTPPost),
					IssueInstant:    issueInstant,
					Version:         "2.0",
					NameIDPolicy: &crewjamsaml.NameIDPolicy{
						Format: &nameIDFormat,
					},
					AssertionConsumerServiceURL: "http://localhost/saml-test",
				},
			}
			return authnRequest
		}

		Convey("valid request", func() {
			authnRequest := makeValidRequest()
			err := svc.ValidateAuthnRequest(spID, authnRequest)
			So(err, ShouldBeNil)
		})

		Convey("invalid destination", func() {
			authnRequest := makeValidRequest()
			authnRequest.Destination = "http://idp.local/wrong"
			err := svc.ValidateAuthnRequest(spID, authnRequest)

			So(err, ShouldBeError, &samlerror.InvalidRequestError{
				Field:    "Destination",
				Actual:   "http://idp.local/wrong",
				Expected: []string{"http://idp.local/login"},
			})
		})

		Convey("unsupported binding", func() {
			authnRequest := makeValidRequest()
			authnRequest.ProtocolBinding = "urn:oasis:names:tc:SAML:2.0:bindings:SOAP"
			err := svc.ValidateAuthnRequest(spID, authnRequest)
			So(err, ShouldBeError, &samlerror.InvalidRequestError{
				Field:  "ProtocolBinding",
				Actual: authnRequest.ProtocolBinding,
				Expected: []string{
					"urn:oasis:names:tc:SAML:2.0:bindings:HTTP-Redirect",
					"urn:oasis:names:tc:SAML:2.0:bindings:HTTP-POST",
				}})
		})

		Convey("unsupported version", func() {
			authnRequest := makeValidRequest()
			authnRequest.Version = "1.0"
			err := svc.ValidateAuthnRequest(spID, authnRequest)
			So(err, ShouldBeError, &samlerror.InvalidRequestError{
				Field:    "Version",
				Actual:   authnRequest.Version,
				Expected: []string{samlprotocol.SAMLVersion2},
			})
		})

		Convey("expired request", func() {
			authnRequest := makeValidRequest()
			issueInstant, _ := time.Parse(time.RFC3339, "2006-01-02T14:00:05Z")
			authnRequest.IssueInstant = issueInstant
			err := svc.ValidateAuthnRequest(spID, authnRequest)
			So(err, ShouldBeError, &samlerror.InvalidRequestError{
				Field:  "IssueInstant",
				Actual: issueInstant.Format(time.RFC3339),
				Reason: "request expired",
			})
		})

		Convey("unsupported name id format", func() {
			authnRequest := makeValidRequest()
			format := "urn:oasis:names:tc:SAML:1.1:nameid-format:X509SubjectName"
			authnRequest.NameIDPolicy.Format = &format
			err := svc.ValidateAuthnRequest(spID, authnRequest)
			So(err, ShouldBeError, &samlerror.InvalidRequestError{
				Field:  "NameIDPolicy/Format",
				Actual: format,
				Expected: []string{
					"urn:oasis:names:tc:SAML:1.1:nameid-format:unspecified",
					"urn:oasis:names:tc:SAML:1.1:nameid-format:emailAddress",
				},
			})
		})

		Convey("acs url not allowed", func() {
			authnRequest := makeValidRequest()
			authnRequest.AssertionConsumerServiceURL = "http://localhost/wrong"
			err := svc.ValidateAuthnRequest(spID, authnRequest)
			So(err, ShouldBeError, &samlerror.InvalidRequestError{
				Field:  "AssertionConsumerServiceURL",
				Actual: "http://localhost/wrong",
				Reason: "AssertionConsumerServiceURL not allowed",
			})
		})
	})

}
