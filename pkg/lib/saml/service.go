package saml

import (
	"bytes"
	"fmt"
	"net/url"
	"text/template"
	"time"

	crewjamsaml "github.com/crewjam/saml"

	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/lib/saml/samlprotocol"
	"github.com/authgear/authgear-server/pkg/util/clock"
	"github.com/authgear/authgear-server/pkg/util/duration"
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

func (s *Service) idpEntityID() string {
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
		return nil, ErrServiceProviderNotFound
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
		EntityID: s.idpEntityID(),
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
		descriptor,
	}, nil
}

// Validate the AuthnRequest
// This method does not verify the signature
func (s *Service) ValidateAuthnRequest(serviceProviderId string, authnRequest *samlprotocol.AuthnRequest) error {
	now := s.Clock.NowUTC()
	sp, ok := s.SAMLConfig.ResolveProvider(serviceProviderId)
	if !ok {
		return ErrServiceProviderNotFound
	}

	if authnRequest.Destination != "" {
		if authnRequest.Destination != s.Endpoints.SAMLLoginURL(sp.ID).String() {
			return fmt.Errorf("unexpected destination")
		}
	}

	if !authnRequest.GetProtocolBinding().IsSupported() {
		return fmt.Errorf("unsupported binding")
	}

	if authnRequest.IssueInstant.Add(MaxAuthnRequestValidDuration).Before(now) {
		return fmt.Errorf("request expired")
	}

	if authnRequest.Version != samlprotocol.SAMLVersion2 {
		return fmt.Errorf("Request Version must be 2.0")
	}

	if authnRequest.NameIDPolicy != nil && authnRequest.NameIDPolicy.Format != nil {
		reqNameIDFormat := *authnRequest.NameIDPolicy.Format
		if reqNameIDFormat != string(sp.NameIDFormat) &&
			// unspecified is always allowed
			reqNameIDFormat != string(config.NameIDFormatUnspecified) {
			return fmt.Errorf("unsupported Name Identifier Format")
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
			return fmt.Errorf("AssertionConsumerServiceURL not allowed")
		}
	}

	return nil
}
