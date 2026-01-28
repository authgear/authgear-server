package saml_test

import (
	"context"
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"math/big"
	"net/url"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/lestrrat-go/jwx/v2/jwk"
	. "github.com/smartystreets/goconvey/convey"

	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/lib/saml"
	"github.com/authgear/authgear-server/pkg/lib/saml/samlprotocol"
	"github.com/authgear/authgear-server/pkg/util/clock"
	"github.com/authgear/authgear-server/pkg/util/template"
)

const pemCert = `
-----BEGIN CERTIFICATE-----
MIICvDCCAaSgAwIBAgIQdYSL2dOaN9QHxzugY+xbjjANBgkqhkiG9w0BAQsFADAP
MQ0wCwYDVQQDEwR0ZXN0MCAXDTI0MDkwNTA3MzcxMVoYDzIwNzQwODI0MDczNzEx
WjAPMQ0wCwYDVQQDEwR0ZXN0MIIBIjANBgkqhkiG9w0BAQEFAAOCAQ8AMIIBCgKC
AQEArrtotTiwy0GSjr+a4i5KXEwZYIajhVazoCyIbC1ogchkvOWMU9bKA3vR2to/
QNAOLF+ysYS/jjnctAQTz8jVCuneV1fKrIWfUyQ0gIsHCgnItXuaNiH6XCRYEUxc
g0d6owh6GtH9XFPmcGdhshl2qm59DWRkfTZ77AVnccmawdU0oyIgIJiYuRyHnUhZ
thhSX9GL7JUFjIV2cN7GwVMtrF6eCc4vOnZ6g8Q9KOU5i9cBnP85aoh17yKCZPpg
mtInA5FN+3JvKeqdFG7fw427a9JiVlT6p4WYAgCeVWwPtjvKXU9Kb+ph2urfBJoE
RVMXvG2TezY2Vzj7sNUhyKNM6wIDAQABoxIwEDAOBgNVHQ8BAf8EBAMCB4AwDQYJ
KoZIhvcNAQELBQADggEBAJNju5+RqjUrI0jS+9iwz/CoNESN0aI9zBJX/IELwCQ3
XhZ9ZPPzqH8rcl0FMR/Rh25XGfDpWO1eDLY7dPCz0AYXT+qfvhRccP32bnD2L+O8
PVHEdBEBFBMk2hlK/kozOOI8QRODvkPxmuopEAT7S+V/BK/3XOkkn8dGxoe+3sVt
og96FvZ3r3495xebFZWHxNECv5Slj8iaHzfqWOCI1p5MrRS+NeJimHMqpo7KhnlB
RnUXcFkdRIKGMztcONpsxoGMo8+QLdjSHDoRXOuHHmBK1g3woNeuZZAX944Dylzu
T2zRqm3yyu2XEfF8k/Z7+b1L1td7tZNa6EbaNi/+y4c=
-----END CERTIFICATE-----
`

const anotherCert = `
-----BEGIN CERTIFICATE-----
MIIC1TCCAb2gAwIBAgIRAJpxx1DW2ObGLT5lUpXARWkwDQYJKoZIhvcNAQELBQAw
GzEZMBcGA1UEAxMQbXktYXBwLmxvY2FsaG9zdDAgFw0yNDA4MDkwODA3MzlaGA8y
MDc0MDcyODA4MDczOVowGzEZMBcGA1UEAxMQbXktYXBwLmxvY2FsaG9zdDCCASIw
DQYJKoZIhvcNAQEBBQADggEPADCCAQoCggEBAN83SCP6m3ayNriEX6VLiwCqoIHu
E1d2vFwULyUWOjinI3olWWkA1txAZu2e0Rm+Zslq2sWx/HZ5e83NCzyLQ8aaG1JQ
OtpbxV2IOybOonveZr1qszvs+1ofGw9sW6AZa7vhH9HhuDqZnM6ArsC7E/D03D4x
J/2hb6uVj9zHb+Cx4vh1nAnBXXwOSIuo1Jm4a0vZHFs8HT2gmX31K/5hhJuchqiH
ptqerf0OHq/Zyx+v40oj3/cFwGAJ291z6kv318bfjBhZTdQ2ovbnFnU9NfQ02IgW
tSj1Grr8dAp5aIDZvgvvYg/m+FnyMqrSU5s0NIyn13tqipZgN4YUk8CUkCECAwEA
AaMSMBAwDgYDVR0PAQH/BAQDAgeAMA0GCSqGSIb3DQEBCwUAA4IBAQAVuZEbgLi0
gzKy5x+L1j+uQMFdY4taFWGdTF7gZx/hw2YpKakPSCl/Sb+624u3+XhQSzByjt7m
0yGhAml5aLQ+y7jOAwagL0pWhK/AW6kZKU2lz36J+T8LTzq3YOFBHrLTJ58ZcWKe
kgwAWDr8Uj9BgxnQWF4Rwu8yAP8POV4E6aIajalFK3tNdyGaXIS5rSHGd/QKuJNW
eCHF7sKGUSTw3p3MADXGkDykUCuXevyNACH6opOLrDCHr/uEEFmSTVf5zlIeSk+Y
EMgvAyAtQw4fi3WItQNOSLm+01kxkCC1SF+LXTSUPMsLOnX++WJ4u4VJTMfqrh6d
UgPkRnolBQXT
-----END CERTIFICATE-----`

func TestSAMLService(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	clk := clock.NewMockClockAt("2006-01-02T15:04:05Z")
	endpoints := NewMockSAMLEndpoints(ctrl)
	spID := "testsp"
	loginEndpoint, _ := url.Parse("http://idp.local/login")
	endpoints.EXPECT().SAMLLoginURL(spID).AnyTimes().Return(loginEndpoint)

	createSP := func() *config.SAMLServiceProviderConfig {
		return &config.SAMLServiceProviderConfig{
			ClientID:     spID,
			NameIDFormat: samlprotocol.SAMLNameIDFormatEmailAddress,
			AcsURLs: []string{
				"http://localhost/saml-test",
			},
		}
	}

	createService := func() *saml.Service {
		sp := createSP()
		return &saml.Service{
			Clock: clk,
			AppID: config.AppID("test"),
			SAMLEnvironmentConfig: config.SAMLEnvironmentConfig{
				IdPEntityIDTemplate: "urn:{{.app_id}}.localhost",
			},
			SAMLConfig: &config.SAMLConfig{
				ServiceProviders: []*config.SAMLServiceProviderConfig{
					sp,
				},
			},
			SAMLIdpSigningMaterials: nil,
			SAMLSpSigningMaterials:  nil,
			Endpoints:               endpoints,
			TemplateEngine:          &template.Engine{},
		}
	}

	Convey("ValidateAuthnRequest", t, func() {
		makeValidRequest := func() *samlprotocol.AuthnRequest {
			issueInstant, _ := time.Parse(time.RFC3339, "2006-01-02T15:00:05Z")
			nameIDFormat := "urn:oasis:names:tc:SAML:1.1:nameid-format:emailAddress"
			authnRequest := &samlprotocol.AuthnRequest{
				ID:              "id_test",
				Destination:     "http://idp.local/login",
				ProtocolBinding: string(samlprotocol.SAMLBindingHTTPPost),
				IssueInstant:    issueInstant,
				Version:         "2.0",
				NameIDPolicy: &samlprotocol.NameIDPolicy{
					Format: &nameIDFormat,
				},
				AssertionConsumerServiceURL: "http://localhost/saml-test",
			}
			return authnRequest
		}

		Convey("valid request", func() {
			authnRequest := makeValidRequest()
			err := createService().ValidateAuthnRequest(spID, authnRequest)
			So(err, ShouldBeNil)
		})

		Convey("invalid destination", func() {
			authnRequest := makeValidRequest()
			authnRequest.Destination = "http://idp.local/wrong"
			err := createService().ValidateAuthnRequest(spID, authnRequest)

			So(err, ShouldBeError, &samlprotocol.InvalidRequestError{
				Field:    "Destination",
				Actual:   "http://idp.local/wrong",
				Expected: []string{"http://idp.local/login"},
				Reason:   "unexpected Destination",
			})
		})

		Convey("unsupported binding", func() {
			authnRequest := makeValidRequest()
			authnRequest.ProtocolBinding = "urn:oasis:names:tc:SAML:2.0:bindings:SOAP"
			err := createService().ValidateAuthnRequest(spID, authnRequest)
			So(err, ShouldBeError, &samlprotocol.InvalidRequestError{
				Field:  "ProtocolBinding",
				Actual: authnRequest.ProtocolBinding,
				Expected: []string{
					"urn:oasis:names:tc:SAML:2.0:bindings:HTTP-POST",
				},
				Reason: "unsupported ProtocolBinding",
			})
		})

		Convey("unsupported version", func() {
			authnRequest := makeValidRequest()
			authnRequest.Version = "1.0"
			err := createService().ValidateAuthnRequest(spID, authnRequest)
			So(err, ShouldBeError, &samlprotocol.InvalidRequestError{
				Field:    "Version",
				Actual:   authnRequest.Version,
				Expected: []string{samlprotocol.SAMLVersion2},
				Reason:   "unsupported Version",
			})
		})

		Convey("expired request", func() {
			authnRequest := makeValidRequest()
			issueInstant, _ := time.Parse(time.RFC3339, "2006-01-02T14:00:05Z")
			authnRequest.IssueInstant = issueInstant
			err := createService().ValidateAuthnRequest(spID, authnRequest)
			So(err, ShouldBeError, &samlprotocol.InvalidRequestError{
				Field:  "IssueInstant",
				Actual: issueInstant.Format(time.RFC3339),
				Reason: "request expired",
			})
		})

		Convey("unsupported name id format", func() {
			authnRequest := makeValidRequest()
			format := "urn:oasis:names:tc:SAML:1.1:nameid-format:X509SubjectName"
			authnRequest.NameIDPolicy.Format = &format
			err := createService().ValidateAuthnRequest(spID, authnRequest)
			So(err, ShouldBeError, &samlprotocol.InvalidRequestError{
				Field:  "NameIDPolicy/Format",
				Actual: format,
				Expected: []string{
					"urn:oasis:names:tc:SAML:1.1:nameid-format:emailAddress",
					"urn:oasis:names:tc:SAML:1.1:nameid-format:unspecified",
				},
				Reason: "unsupported NameIDPolicy Format",
			})
		})

		Convey("acs url not allowed", func() {
			authnRequest := makeValidRequest()
			authnRequest.AssertionConsumerServiceURL = "http://localhost/wrong"
			err := createService().ValidateAuthnRequest(spID, authnRequest)
			So(err, ShouldBeError, &samlprotocol.InvalidRequestError{
				Field:  "AssertionConsumerServiceURL",
				Actual: "http://localhost/wrong",
				Reason: "AssertionConsumerServiceURL not allowed",
			})
		})
	})

	Convey("VerifyEmbeddedSignature", t, func() {

		// Keep the indentation as spaces in the xml, or the test will fail
		requestXml := `
<?xml version="1.0"?>
<samlp:AuthnRequest xmlns:samlp="urn:oasis:names:tc:SAML:2.0:protocol" xmlns:saml="urn:oasis:names:tc:SAML:2.0:assertion" ForceAuthn="false" ID="pfxfcc76a4e-1dad-24bb-6753-aa23909601e3" IssueInstant="2024-09-05T07:35:34Z" Destination="http://localhost:3000/saml2/login/sp1" AssertionConsumerServiceURL="https://sptest.iamshowcase.com/acs" ProtocolBinding="urn:oasis:names:tc:SAML:2.0:bindings:HTTP-POST" Version="2.0"><saml:Issuer>IAMShowcase</saml:Issuer><ds:Signature xmlns:ds="http://www.w3.org/2000/09/xmldsig#">
  <ds:SignedInfo><ds:CanonicalizationMethod Algorithm="http://www.w3.org/2001/10/xml-exc-c14n#"/>
    <ds:SignatureMethod Algorithm="http://www.w3.org/2001/04/xmldsig-more#rsa-sha256"/>
  <ds:Reference URI="#pfxfcc76a4e-1dad-24bb-6753-aa23909601e3"><ds:Transforms><ds:Transform Algorithm="http://www.w3.org/2000/09/xmldsig#enveloped-signature"/><ds:Transform Algorithm="http://www.w3.org/2001/10/xml-exc-c14n#"/></ds:Transforms><ds:DigestMethod Algorithm="http://www.w3.org/2000/09/xmldsig#sha1"/><ds:DigestValue>1jRKgEs73mif6vqWcGPukA9HzP4=</ds:DigestValue></ds:Reference></ds:SignedInfo><ds:SignatureValue>ncDAze9fD8EOw24HlXjsi8xVIHwDACHCnfs/axtybRC8VyEVuuZCO00MxSHGEv1oBoj8OQwGT5IxPupKUwoWNy6QLm6jC2+CHCu53FcYEvNz+m5Pk8xdUWHQLR7tZ8Eb1wyavFm7KD6VgRjppKByz8F6WGP5tP/x2KM3MI4Mh/Ki1NYbkXm7WykAOO2FjZE9Lmi9he/ScQO+g03Hzz91Uk9kdhsx7aCz4b+YltdOpk6rnUe8WCOba2/jXzhwr8IndlsmxPmqrlASbe/E5POhl89ap1Vsaur3hh1FP6DyhhSp8DvFarSeqNZYSbhDylcXa52ro6Kl8ErGOaMUW3getA==</ds:SignatureValue>
<ds:KeyInfo><ds:X509Data><ds:X509Certificate>MIICvDCCAaSgAwIBAgIQdYSL2dOaN9QHxzugY+xbjjANBgkqhkiG9w0BAQsFADAPMQ0wCwYDVQQDEwR0ZXN0MCAXDTI0MDkwNTA3MzcxMVoYDzIwNzQwODI0MDczNzExWjAPMQ0wCwYDVQQDEwR0ZXN0MIIBIjANBgkqhkiG9w0BAQEFAAOCAQ8AMIIBCgKCAQEArrtotTiwy0GSjr+a4i5KXEwZYIajhVazoCyIbC1ogchkvOWMU9bKA3vR2to/QNAOLF+ysYS/jjnctAQTz8jVCuneV1fKrIWfUyQ0gIsHCgnItXuaNiH6XCRYEUxcg0d6owh6GtH9XFPmcGdhshl2qm59DWRkfTZ77AVnccmawdU0oyIgIJiYuRyHnUhZthhSX9GL7JUFjIV2cN7GwVMtrF6eCc4vOnZ6g8Q9KOU5i9cBnP85aoh17yKCZPpgmtInA5FN+3JvKeqdFG7fw427a9JiVlT6p4WYAgCeVWwPtjvKXU9Kb+ph2urfBJoERVMXvG2TezY2Vzj7sNUhyKNM6wIDAQABoxIwEDAOBgNVHQ8BAf8EBAMCB4AwDQYJKoZIhvcNAQELBQADggEBAJNju5+RqjUrI0jS+9iwz/CoNESN0aI9zBJX/IELwCQ3XhZ9ZPPzqH8rcl0FMR/Rh25XGfDpWO1eDLY7dPCz0AYXT+qfvhRccP32bnD2L+O8PVHEdBEBFBMk2hlK/kozOOI8QRODvkPxmuopEAT7S+V/BK/3XOkkn8dGxoe+3sVtog96FvZ3r3495xebFZWHxNECv5Slj8iaHzfqWOCI1p5MrRS+NeJimHMqpo7KhnlBRnUXcFkdRIKGMztcONpsxoGMo8+QLdjSHDoRXOuHHmBK1g3woNeuZZAX944DylzuT2zRqm3yyu2XEfF8k/Z7+b1L1td7tZNa6EbaNi/+y4c=</ds:X509Certificate></ds:X509Data></ds:KeyInfo></ds:Signature><saml:Subject>
<saml:NameID Format="urn:oasis:names:tc:SAML:1.1:nameid-format:emailAddress">test@example.com</saml:NameID>
</saml:Subject></samlp:AuthnRequest>
		`

		Convey("success", func() {
			svc := createService()
			svc.SAMLSpSigningMaterials = &config.SAMLSpSigningMaterials{
				config.SAMLSpSigningCertificate{
					ServiceProviderID: spID,
					Certificates: []config.X509Certificate{
						{
							Pem: config.X509CertificatePem(pemCert),
						},
					},
				},
			}
			sp := svc.SAMLConfig.ServiceProviders[0]
			sp.SignatureVerificationEnabled = true

			err := svc.VerifyEmbeddedSignature(sp, requestXml)
			So(err, ShouldBeNil)
		})

		Convey("will fail when the cert is incorrect", func() {
			svc := createService()
			svc.SAMLSpSigningMaterials = &config.SAMLSpSigningMaterials{
				config.SAMLSpSigningCertificate{
					ServiceProviderID: spID,
					Certificates: []config.X509Certificate{
						{
							Pem: config.X509CertificatePem(anotherCert),
						},
					},
				},
			}
			sp := svc.SAMLConfig.ServiceProviders[0]
			sp.SignatureVerificationEnabled = true

			err := svc.VerifyEmbeddedSignature(sp, requestXml)
			expectedErr := &samlprotocol.InvalidSignatureError{}
			So(err, ShouldHaveSameTypeAs, expectedErr)
		})

	})

	Convey("VerifyExternalSignature", t, func() {
		// The request in xml is
		/*
			<samlp:AuthnRequest xmlns:samlp="urn:oasis:names:tc:SAML:2.0:protocol" xmlns:saml="urn:oasis:names:tc:SAML:2.0:assertion" ForceAuthn="false"  ID="ae230b376c88c3f4f8c7a4db12d24e38357205829" IssueInstant="2024-09-05T07:35:34Z" Destination="http://localhost:3000/saml2/login/sp1" AssertionConsumerServiceURL="https://sptest.iamshowcase.com/acs"  ProtocolBinding="urn:oasis:names:tc:SAML:2.0:bindings:HTTP-POST"   Version="2.0"><saml:Issuer >IAMShowcase</saml:Issuer><saml:Subject >
			<saml:NameID Format="urn:oasis:names:tc:SAML:1.1:nameid-format:emailAddress">test@example.com</saml:NameID>
			</saml:Subject></samlp:AuthnRequest>
		*/

		// Use this site to encode the request:
		// https://www.samltool.com/encode.php
		samlRequest := "fVLLTuMwFN0jzT9Y3rdxnZQGq4kmTDWaSDwqUmYxO9e5pUaxnfF1gM8njyKVBSx97rnnceU1StO0oujC0T7A/w4wkDfTWBTjIKOdt8JJ1CisNIAiKFEVtzeCz5lovQtOuYaerXy/IRHBB+0sJb+dVzD6ZvQgGwRKSLnJqAQes328ulRpquJDckjVSib1fsFrnkCcxssVZ8uUX1FSInZQWgzShoxyxpMZu5qx5Y6tRLwUcfKPkk3fSFs5eGb0GEIroqhxSjZHh0HEjLFoiM178EnbCNsFJcVHyl/OYmfAV+BftILHh5tJA3sRbEMvPdfS4NG9KokwV85EUmFfZHu6zLW2tbZP3x9lP5FQ/NnttrPtfbXrFchf8DiG7ik0Xw8hxVjYk7wsbquT6To6m5xoVbd/BhVI/uNiAu56z3IznNzI8HWYxXwxIrqeHUaqACN1U9S1B0SaD4V/wps0bTOWPXlP6oNZdG6fT8/Pnyt/Bw=="
		// Use this site to generate the signature:
		// https://www.samltool.com/sign_authn.php
		signature := "LAre0pDAbJPSP1swdYTIDuTltnQGyfDtmJBnXyCr6Hij/EWvAhtS7g3SuDx3GYaUc2gv/NE1JFIXMEewziF80n2GcP9Xfog8ToxEqKcjT2VUTvAZGnY66u9jRcoqVhnbG15Q11HmQiGFVD0MoPVebOD8LtDOD1l6+IzuIYk+uHsiqHNM98UM+VDIZ0YlHGoO/bu9cJIpGStr+xQEA/VJcrpD+qB6a2QB7Tn2D+CIK5cf+7uROm44loJeI7vs9bwSvNQM7xvJPewXhWtqWCqg/mFsaV/FgYoHfP8zsBAi2RNJLf454Klih47he7wps8VN4FvtW4DP4ZE8J9HXXaYO/Q=="

		relayState := "indigo"

		Convey("success", func() {
			svc := createService()
			svc.SAMLSpSigningMaterials = &config.SAMLSpSigningMaterials{
				config.SAMLSpSigningCertificate{
					ServiceProviderID: spID,
					Certificates: []config.X509Certificate{
						{
							Pem: config.X509CertificatePem(pemCert),
						},
					},
				},
			}
			sp := svc.SAMLConfig.ServiceProviders[0]
			sp.SignatureVerificationEnabled = true
			err := svc.VerifyExternalSignature(
				sp,
				&saml.SAMLElementSigned{
					SAMLRequest: samlRequest,
				},
				"http://www.w3.org/2001/04/xmldsig-more#rsa-sha256",
				relayState,
				signature)
			So(err, ShouldBeNil)
		})

		Convey("will fail when the cert is incorrect", func() {
			svc := createService()
			svc.SAMLSpSigningMaterials = &config.SAMLSpSigningMaterials{
				config.SAMLSpSigningCertificate{
					ServiceProviderID: spID,
					Certificates: []config.X509Certificate{
						{
							Pem: config.X509CertificatePem(anotherCert),
						},
					},
				},
			}
			sp := svc.SAMLConfig.ServiceProviders[0]
			sp.SignatureVerificationEnabled = true

			err := svc.VerifyExternalSignature(
				sp,
				&saml.SAMLElementSigned{
					SAMLRequest: samlRequest,
				},
				"http://www.w3.org/2001/04/xmldsig-more#rsa-sha256",
				relayState,
				signature)
			expectedErr := &samlprotocol.InvalidSignatureError{}
			So(err, ShouldHaveSameTypeAs, expectedErr)
		})
	})

	Convey("ResolveUserAttributes", t, func() {
		Convey("success", func() {
			svc := createService()
			sp := createSP()
			sp.Attributes = &config.SAMLAttributesConfig{
				Definitions: []config.SAMLAttributeDefinition{
					{Name: "strattr", NameFormat: config.SAMLAttributeNameFormatUnspecified, FriendlyName: "A string"},
					{Name: "boolattr", NameFormat: config.SAMLAttributeNameFormatUnspecified},
					{Name: "floatattr", NameFormat: config.SAMLAttributeNameFormatUnspecified},
					{Name: "sliceattr", NameFormat: config.SAMLAttributeNameFormatUnspecified},
					{Name: "nestedattr", NameFormat: config.SAMLAttributeNameFormatUnspecified},
					{Name: "missingattr", NameFormat: config.SAMLAttributeNameFormatUnspecified},
					{Name: "nullattr", NameFormat: config.SAMLAttributeNameFormatUnspecified},
					{Name: "templateattr", NameFormat: config.SAMLAttributeNameFormatUnspecified},
				},
				Mappings: []config.SAMLAttributeMapping{
					{
						From: &config.SAMLAttributeMappingFrom{
							UserProfileJSONPointer: config.UserProfileJSONPointer{
								UserProfile: &config.JSONPointer{
									Pointer: "/str",
								},
							},
						},
						To: &config.SAMLAttributeMappingTo{
							SAMLAttribute: "strattr",
						},
					},
					{
						From: &config.SAMLAttributeMappingFrom{
							UserProfileJSONPointer: config.UserProfileJSONPointer{
								UserProfile: &config.JSONPointer{
									Pointer: "/bool",
								},
							},
						},
						To: &config.SAMLAttributeMappingTo{
							SAMLAttribute: "boolattr",
						},
					},
					{
						From: &config.SAMLAttributeMappingFrom{
							UserProfileJSONPointer: config.UserProfileJSONPointer{
								UserProfile: &config.JSONPointer{
									Pointer: "/float",
								},
							},
						},
						To: &config.SAMLAttributeMappingTo{
							SAMLAttribute: "floatattr",
						},
					},
					{
						From: &config.SAMLAttributeMappingFrom{
							UserProfileJSONPointer: config.UserProfileJSONPointer{
								UserProfile: &config.JSONPointer{
									Pointer: "/slice",
								},
							},
						},
						To: &config.SAMLAttributeMappingTo{
							SAMLAttribute: "sliceattr",
						},
					},
					{
						From: &config.SAMLAttributeMappingFrom{
							UserProfileJSONPointer: config.UserProfileJSONPointer{
								UserProfile: &config.JSONPointer{
									Pointer: "/map/nested",
								},
							},
						},
						To: &config.SAMLAttributeMappingTo{
							SAMLAttribute: "nestedattr",
						},
					},
					{
						From: &config.SAMLAttributeMappingFrom{
							UserProfileJSONPointer: config.UserProfileJSONPointer{
								UserProfile: &config.JSONPointer{
									Pointer: "/missing",
								},
							},
						},
						To: &config.SAMLAttributeMappingTo{
							SAMLAttribute: "missingattr",
						},
					},
					{
						From: &config.SAMLAttributeMappingFrom{
							UserProfileJSONPointer: config.UserProfileJSONPointer{
								UserProfile: &config.JSONPointer{
									Pointer: "/null",
								},
							},
						},
						To: &config.SAMLAttributeMappingTo{
							SAMLAttribute: "nullattr",
						},
					},
					{
						From: &config.SAMLAttributeMappingFrom{
							TextTemplate: config.TextTemplate{TextTemplate: &config.TextTemplateBody{
								Template: `{{if .sub}}{{.sub}}@example.com{{end}}`,
							}},
						},
						To: &config.SAMLAttributeMappingTo{
							SAMLAttribute: "templateattr",
						},
					},
				},
			}

			var userInfo map[string]interface{}
			err := json.Unmarshal([]byte(userInfoJson), &userInfo)
			So(err, ShouldBeNil)
			attrs, err := svc.ResolveUserAttributes(context.Background(), sp, userInfo)
			So(err, ShouldBeNil)
			So(attrs, ShouldResemble, []samlprotocol.Attribute{
				{
					FriendlyName: "User ID",
					Name:         "sub",
					NameFormat:   samlprotocol.SAMLAttrnameFormatBasic,
					Values: []samlprotocol.AttributeValue{{
						Type:  samlprotocol.SAMLAttrTypeString,
						Value: "userid",
					}},
				},
				{
					Name:         "strattr",
					NameFormat:   string(config.SAMLAttributeNameFormatUnspecified),
					FriendlyName: "A string",
					Values: []samlprotocol.AttributeValue{{
						Type:  samlprotocol.SAMLAttrTypeString,
						Value: "teststr",
					}},
				},
				{
					Name:       "boolattr",
					NameFormat: string(config.SAMLAttributeNameFormatUnspecified),
					Values: []samlprotocol.AttributeValue{{
						Type:  samlprotocol.SAMLAttrTypeBoolean,
						Value: "true",
					}},
				},
				{
					Name:       "floatattr",
					NameFormat: string(config.SAMLAttributeNameFormatUnspecified),
					Values: []samlprotocol.AttributeValue{{
						Type:  samlprotocol.SAMLAttrTypeDecimal,
						Value: "12.5",
					}},
				},
				{
					Name:       "sliceattr",
					NameFormat: string(config.SAMLAttributeNameFormatUnspecified),
					Values: []samlprotocol.AttributeValue{{
						Type:  samlprotocol.SAMLAttrTypeString,
						Value: "item1",
					}, {
						Type:  samlprotocol.SAMLAttrTypeString,
						Value: "item2",
					}},
				},
				{
					Name:       "nestedattr",
					NameFormat: string(config.SAMLAttributeNameFormatUnspecified),
					Values: []samlprotocol.AttributeValue{{
						Type:  samlprotocol.SAMLAttrTypeString,
						Value: "nesteditem",
					}},
				},
				{
					Name:       "missingattr",
					NameFormat: string(config.SAMLAttributeNameFormatUnspecified),
					Values:     []samlprotocol.AttributeValue{},
				},
				{
					Name:       "nullattr",
					NameFormat: string(config.SAMLAttributeNameFormatUnspecified),
					Values:     []samlprotocol.AttributeValue{{IsNil: true}},
				},
				{
					Name:       "templateattr",
					NameFormat: string(config.SAMLAttributeNameFormatUnspecified),
					Values: []samlprotocol.AttributeValue{{
						Type:  samlprotocol.SAMLAttrTypeString,
						Value: "userid@example.com",
					}},
				},
			})
		})

		Convey("map is not supported", func() {
			svc := createService()
			sp := createSP()
			sp.Attributes = &config.SAMLAttributesConfig{
				Definitions: []config.SAMLAttributeDefinition{
					{Name: "mapattr", NameFormat: config.SAMLAttributeNameFormatUnspecified},
				},
				Mappings: []config.SAMLAttributeMapping{
					{
						From: &config.SAMLAttributeMappingFrom{
							UserProfileJSONPointer: config.UserProfileJSONPointer{
								UserProfile: &config.JSONPointer{
									Pointer: "/map",
								},
							},
						},
						To: &config.SAMLAttributeMappingTo{
							SAMLAttribute: "mapattr",
						},
					},
				},
			}

			var userInfo map[string]interface{}
			err := json.Unmarshal([]byte(userInfoJson), &userInfo)
			So(err, ShouldBeNil)
			_, err = svc.ResolveUserAttributes(context.Background(), sp, userInfo)
			So(err, ShouldBeError, &samlprotocol.UnsupportedAttributeTypeError{
				AttributeName:      "mapattr",
				UserProfilePointer: "/map",
			})
		})
	})

	Convey("ConstructSignedQueryParameters", t, func() {
		Convey("should generate valid signature for redirect binding", func() {
			svc := createService()

			// Generate a key for testing
			rsaKey, err := rsa.GenerateKey(rand.Reader, 2048)
			So(err, ShouldBeNil)
			jwkKey, err := jwk.FromRaw(rsaKey)
			So(err, ShouldBeNil)
			_ = jwkKey.Set(jwk.KeyIDKey, "test-key-id")

			svc.SAMLConfig.Signing = &config.SAMLSigningConfig{
				KeyID:           "test-key-id",
				SignatureMethod: config.SAMLSigningSignatureMethodRSASHA256,
			}

			certBytes, err := x509.CreateCertificate(rand.Reader, &x509.Certificate{
				SerialNumber: big.NewInt(1),
			}, &x509.Certificate{
				SerialNumber: big.NewInt(1),
			}, &rsaKey.PublicKey, rsaKey)
			So(err, ShouldBeNil)

			svc.SAMLIdpSigningMaterials = &config.SAMLIdpSigningMaterials{
				Certificates: []*config.SAMLIdpSigningCertificate{
					{
						Key: &config.JWK{Key: jwkKey},
						Certificate: &config.X509Certificate{
							Pem: config.X509CertificatePem(pem.EncodeToMemory(&pem.Block{
								Type:  "CERTIFICATE",
								Bytes: certBytes,
							})),
						},
					},
				},
			}

			// Also set it as SP signing material for verification
			svc.SAMLSpSigningMaterials = &config.SAMLSpSigningMaterials{
				{
					ServiceProviderID: spID,
					Certificates: []config.X509Certificate{
						{
							Pem: config.X509CertificatePem(pem.EncodeToMemory(&pem.Block{
								Type:  "CERTIFICATE",
								Bytes: certBytes,
							})),
						},
					},
				},
			}

			relayState := "test-relay-state"
			el := &saml.SAMLElementToSign{
				SAMLRequest: "test-request",
			}

			q, err := svc.ConstructSignedQueryParameters(relayState, el)
			So(err, ShouldBeNil)

			signature := q.Get("Signature")
			sigAlg := q.Get("SigAlg")

			So(signature, ShouldNotBeEmpty)
			So(sigAlg, ShouldEqual, "http://www.w3.org/2001/04/xmldsig-more#rsa-sha256")

			// Verify the signature independently
			// 1. Reconstruct the signing value according to SAML spec
			// SAMLRequest=value&RelayState=value&SigAlg=value (URL escaped)
			expectedSigningValue := fmt.Sprintf("SAMLRequest=%s&RelayState=%s&SigAlg=%s",
				url.QueryEscape(q.Get("SAMLRequest")),
				url.QueryEscape(q.Get("RelayState")),
				url.QueryEscape(q.Get("SigAlg")),
			)

			// 2. Decode the signature
			decodedSignature, err := base64.StdEncoding.DecodeString(signature)
			So(err, ShouldBeNil)

			// 3. Hash the signing value
			hash := crypto.SHA256.New()
			_, _ = hash.Write([]byte(expectedSigningValue))
			hashed := hash.Sum(nil)

			// 4. Verify using the public key directly
			err = rsa.VerifyPKCS1v15(&rsaKey.PublicKey, crypto.SHA256, hashed, decodedSignature)
			So(err, ShouldBeNil)
		})
	})
}

var userInfoJson string = `
{
		"sub": "userid",
		"str": "teststr",
		"bool": true,
		"float": 12.5,
		"slice": ["item1", "item2"],
		"map": {
			"nested": "nesteditem"
		},
		"null": null
}
`
