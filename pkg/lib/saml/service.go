package saml

import (
	"bytes"
	"crypto/rsa"
	"crypto/tls"
	"net/url"
	"text/template"
	"time"

	crewjamsaml "github.com/crewjam/saml"

	dsig "github.com/russellhaering/goxmldsig"

	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/lib/saml/samlerror"
	"github.com/authgear/authgear-server/pkg/lib/saml/samlprotocol"
	"github.com/authgear/authgear-server/pkg/util/clock"
	"github.com/authgear/authgear-server/pkg/util/duration"
	"github.com/authgear/authgear-server/pkg/util/setutil"
	"github.com/authgear/authgear-server/pkg/util/slice"
)

const MetadataValidDuration = time.Hour * 24
const MaxAuthnRequestValidDuration = duration.Short

//go:generate mockgen -source=service.go -destination=service_mock_test.go -package saml_test

type SAMLEndpoints interface {
	SAMLLoginURL(serviceProviderId string) *url.URL
}

type Service struct {
	Clock                   clock.Clock
	AppID                   config.AppID
	SAMLEnvironmentConfig   config.SAMLEnvironmentConfig
	SAMLConfig              *config.SAMLConfig
	SAMLIdpSigningMaterials *config.SAMLIdpSigningMaterials
	Endpoints               SAMLEndpoints
}

func (s *Service) IdpEntityID() string {
	idpEntityIdTemplate, err := template.New("").Parse(s.SAMLEnvironmentConfig.IdPEntityIDTemplate)
	if err != nil {
		panic(err)
	}
	var idpEntityIDBytes bytes.Buffer
	err = idpEntityIdTemplate.Execute(&idpEntityIDBytes, map[string]interface{}{
		"app_id": s.AppID,
	})
	if err != nil {
		panic(err)
	}

	return idpEntityIDBytes.String()
}

func (s *Service) IdpMetadata(serviceProviderId string) (*samlprotocol.Metadata, error) {
	sp, ok := s.SAMLConfig.ResolveProvider(serviceProviderId)
	if !ok {
		return nil, samlerror.ErrServiceProviderNotFound
	}

	keyDescriptors := []crewjamsaml.KeyDescriptor{}
	if cert, ok := s.SAMLIdpSigningMaterials.FindSigningCert(s.SAMLConfig.Signing.KeyID); ok {
		keyDescriptors = append(keyDescriptors,
			crewjamsaml.KeyDescriptor{
				Use: "signing",
				KeyInfo: crewjamsaml.KeyInfo{
					X509Data: crewjamsaml.X509Data{
						X509Certificates: []crewjamsaml.X509Certificate{
							{Data: cert.Certificate.Base64Data()},
						},
					},
				},
			})
	}

	descriptor := samlprotocol.EntityDescriptor{
		EntityID: s.IdpEntityID(),
		IDPSSODescriptors: []crewjamsaml.IDPSSODescriptor{
			{
				SSODescriptor: crewjamsaml.SSODescriptor{
					RoleDescriptor: crewjamsaml.RoleDescriptor{
						ProtocolSupportEnumeration: "urn:oasis:names:tc:SAML:2.0:protocol",
						KeyDescriptors:             keyDescriptors,
					},
					NameIDFormats: []crewjamsaml.NameIDFormat{
						crewjamsaml.NameIDFormat(sp.NameIDFormat),
					},
				},
				SingleSignOnServices: []crewjamsaml.Endpoint{
					{
						Binding:  crewjamsaml.HTTPRedirectBinding,
						Location: s.Endpoints.SAMLLoginURL(sp.ID).String(),
					},
					{
						Binding:  crewjamsaml.HTTPPostBinding,
						Location: s.Endpoints.SAMLLoginURL(sp.ID).String(),
					},
				},
			},
		},
	}

	return &samlprotocol.Metadata{
		EntityDescriptor: descriptor,
	}, nil
}

// Validate the AuthnRequest
// This method does not verify the signature
func (s *Service) ValidateAuthnRequest(serviceProviderId string, authnRequest *samlprotocol.AuthnRequest) error {
	now := s.Clock.NowUTC()
	sp, ok := s.SAMLConfig.ResolveProvider(serviceProviderId)
	if !ok {
		return samlerror.ErrServiceProviderNotFound
	}

	if authnRequest.Destination != "" {
		if authnRequest.Destination != s.Endpoints.SAMLLoginURL(sp.ID).String() {
			return &samlerror.InvalidRequestError{
				Field:    "Destination",
				Actual:   authnRequest.Destination,
				Expected: []string{s.Endpoints.SAMLLoginURL(sp.ID).String()},
			}
		}
	}

	if !authnRequest.GetProtocolBinding().IsSupported() {
		return &samlerror.InvalidRequestError{
			Field:    "ProtocolBinding",
			Actual:   authnRequest.ProtocolBinding,
			Expected: slice.Map(samlprotocol.SupportedBindings, func(b samlprotocol.SAMLBinding) string { return string(b) }),
		}
	}

	if authnRequest.IssueInstant.Add(MaxAuthnRequestValidDuration).Before(now) {
		return &samlerror.InvalidRequestError{
			Field:  "IssueInstant",
			Actual: authnRequest.IssueInstant.Format(time.RFC3339),
			Reason: "request expired",
		}
	}

	if authnRequest.Version != samlprotocol.SAMLVersion2 {
		return &samlerror.InvalidRequestError{
			Field:    "Version",
			Actual:   authnRequest.Version,
			Expected: []string{samlprotocol.SAMLVersion2},
		}
	}

	// unspecified is always allowed
	allowedNameFormats := setutil.Set[string]{
		string(config.SAMLNameIDFormatUnspecified): {},
	}
	allowedNameFormats.Add(string(sp.NameIDFormat))

	if authnRequest.NameIDPolicy != nil && authnRequest.NameIDPolicy.Format != nil {
		reqNameIDFormat := *authnRequest.NameIDPolicy.Format
		if _, ok := allowedNameFormats[reqNameIDFormat]; !ok {
			return &samlerror.InvalidRequestError{
				Field:    "NameIDPolicy/Format",
				Actual:   reqNameIDFormat,
				Expected: allowedNameFormats.Keys(),
			}
		}
	}

	if authnRequest.AssertionConsumerServiceURL != "" {
		allowed := false
		for _, allowedURL := range sp.AcsURLs {
			if allowedURL == authnRequest.AssertionConsumerServiceURL {
				allowed = true
			}
		}
		if allowed == false {
			return &samlerror.InvalidRequestError{
				Field:  "AssertionConsumerServiceURL",
				Actual: authnRequest.AssertionConsumerServiceURL,
				Reason: "AssertionConsumerServiceURL not allowed",
			}
		}
	}

	// Block unsupported combinations of IsPassive and ForceAuthn
	switch {
	case authnRequest.GetIsPassive() == false && authnRequest.GetForceAuthn() == false:
		// allow as prompt=select_account
	case authnRequest.GetIsPassive() == false && authnRequest.GetForceAuthn() == true:
		// allow as prompt=login
	case authnRequest.GetIsPassive() == true && authnRequest.GetForceAuthn() == false:
		// allow as prompt=none
	case authnRequest.GetIsPassive() == true && authnRequest.GetForceAuthn() == true:
		return &samlerror.InvalidRequestError{
			Reason: "IsPassive=true with ForceAuthn=true is not allowed",
		}
	}

	return nil
}

func (s *Service) IssueSuccessResponse(
	callbackURL string,
	serviceProviderId string,
	authenticatedUserId string,
	inResponseToAuthnRequest *samlprotocol.AuthnRequest,
) (*samlprotocol.Response, error) {
	sp, ok := s.SAMLConfig.ResolveProvider(serviceProviderId)
	if !ok {
		return nil, samlerror.ErrServiceProviderNotFound
	}
	now := s.Clock.NowUTC()
	issuerID := s.IdpEntityID()
	// TODO(saml): Write required fields of the response
	response := samlprotocol.NewSuccessResponse(now, issuerID, inResponseToAuthnRequest.ID)

	// TODO(saml): Use configured destination if configured
	destination := callbackURL
	response.Destination = destination

	// TODO(saml): Use configured recipient if configured
	recipient := callbackURL

	// TODO(saml): Use configured audience if configured
	audience := callbackURL

	nameIDFormat := sp.NameIDFormat
	if nameIDFormatInRequest, ok := inResponseToAuthnRequest.GetNameIDFormat(); ok {
		nameIDFormat = nameIDFormatInRequest
	}

	// allow for some clock skew
	notBefore := now.Add(-1 * duration.ClockSkew)
	// TODO(saml): Allow configurating the valid period
	notOnOrAfter := now.Add(duration.UserInteraction)
	if notBefore.Before(inResponseToAuthnRequest.IssueInstant) {
		notBefore = inResponseToAuthnRequest.IssueInstant
		notOnOrAfter = notBefore.Add(duration.UserInteraction)
	}

	assertion := &crewjamsaml.Assertion{
		ID:           samlprotocol.GenerateAssertionID(),
		IssueInstant: now,
		Version:      samlprotocol.SAMLVersion2,
		Issuer: crewjamsaml.Issuer{
			Format: samlprotocol.SAMLIssertFormatEntity,
			Value:  issuerID,
		},
		Subject: &crewjamsaml.Subject{
			NameID: &crewjamsaml.NameID{
				Format: string(nameIDFormat),
				// TODO(saml): Support different nameid
				Value: authenticatedUserId,
			},
			SubjectConfirmations: []crewjamsaml.SubjectConfirmation{
				{
					Method: "urn:oasis:names:tc:SAML:2.0:cm:bearer",
					SubjectConfirmationData: &crewjamsaml.SubjectConfirmationData{
						InResponseTo: inResponseToAuthnRequest.ID,
						// TODO(saml): Allow configurating the valid period
						NotOnOrAfter: notOnOrAfter,
						Recipient:    recipient,
					},
				},
			},
		},
		Conditions: &crewjamsaml.Conditions{
			NotBefore:    notBefore,
			NotOnOrAfter: notOnOrAfter,
			AudienceRestrictions: []crewjamsaml.AudienceRestriction{
				{
					Audience: crewjamsaml.Audience{Value: audience},
				},
			},
		},
		AuthnStatements: []crewjamsaml.AuthnStatement{
			{
				AuthnInstant: notBefore,
				// TODO(saml): Put the idp session id here
				SessionIndex: "",
				AuthnContext: crewjamsaml.AuthnContext{
					AuthnContextClassRef: &crewjamsaml.AuthnContextClassRef{
						// TODO(saml): Return a correct context by used authenticators
						Value: "urn:oasis:names:tc:SAML:2.0:ac:classes:unspecified",
					},
				},
			},
		},
		AttributeStatements: []crewjamsaml.AttributeStatement{
			{
				// TODO(saml): Return more attributes
				Attributes: []crewjamsaml.Attribute{
					{
						FriendlyName: "User ID",
						Name:         "sub",
						NameFormat:   samlprotocol.SAMLAttrnameFormatBasic,
						Values: []crewjamsaml.AttributeValue{{
							Type:  samlprotocol.SAMLAttrTypeString,
							Value: authenticatedUserId,
						}},
					},
				},
			},
		},
	}

	response.Assertion = assertion

	// Sign the assertion
	// Reference: https://github.com/crewjam/saml/blob/193e551d9a8420216fae88c2b8f4b46696b7bb63/identity_provider.go#L833
	signingContext, err := s.idpSigningContext()
	if err != nil {
		return nil, err
	}

	assertionEl := response.Assertion.Element()

	signedAssertionEl, err := signingContext.SignEnveloped(assertionEl)
	if err != nil {
		return nil, err
	}

	assertionSigEl := signedAssertionEl.ChildElements()[len(signedAssertionEl.ChildElements())-1]
	response.Assertion.Signature = assertionSigEl

	// Sign the response
	responseEl := response.Element()
	signedResponseEl, err := signingContext.SignEnveloped(responseEl)
	if err != nil {
		return nil, err
	}

	responseSigEl := signedResponseEl.ChildElements()[len(signedResponseEl.ChildElements())-1]
	response.Signature = responseSigEl

	return response, nil
}

func (s *Service) idpSigningContext() (*dsig.SigningContext, error) {
	// Create a cert chain based off of the IDP cert and its intermediates.
	activeCert, ok := s.SAMLIdpSigningMaterials.FindSigningCert(s.SAMLConfig.Signing.KeyID)
	if !ok {
		panic("unexpected: cannot find the corresponding idp key by id")
	}

	var signingContext *dsig.SigningContext
	var rsaPrivateKey rsa.PrivateKey
	err := activeCert.Key.Raw(&rsaPrivateKey)
	if err != nil {
		panic(err)
	}

	keyPair := tls.Certificate{
		Certificate: [][]byte{activeCert.Certificate.Data()},
		PrivateKey:  &rsaPrivateKey,
		Leaf:        activeCert.Certificate.X509Certificate(),
	}
	keyStore := dsig.TLSCertKeyStore(keyPair)

	signingContext = dsig.NewDefaultSigningContext(keyStore)

	signatureMethod := s.SAMLConfig.Signing.SignatureMethod.ToDsigSignatureMethod()

	signingContext.Canonicalizer = dsig.MakeC14N10ExclusiveCanonicalizerWithPrefixList("")
	if err := signingContext.SetSignatureMethod(signatureMethod); err != nil {
		return nil, err
	}

	return signingContext, nil
}
