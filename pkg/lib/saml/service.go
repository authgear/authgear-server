package saml

import (
	"bytes"
	"encoding/xml"
	"fmt"
	"net/url"
	"text/template"
	"time"

	crewjamsaml "github.com/crewjam/saml"
	xrv "github.com/mattermost/xml-roundtrip-validator"

	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/util/clock"
	"github.com/authgear/authgear-server/pkg/util/duration"
)

const MetadataValidDuration = time.Hour * 24
const MaxAuthnRequestValidDuration = duration.Short

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

func (s *Service) IdpMetadata(serviceProviderId string) (*Metadata, error) {
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

	descriptor := EntityDescriptor{
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

	return &Metadata{
		descriptor,
	}, nil
}

func (s *Service) ParseAuthnRequest(serviceProviderId string, input []byte) (*AuthnRequest, error) {
	sp, ok := s.SAMLConfig.ResolveProvider(serviceProviderId)
	if !ok {
		return nil, ErrServiceProviderNotFound
	}

	now := s.Clock.NowUTC()
	var req crewjamsaml.AuthnRequest
	if err := xrv.Validate(bytes.NewReader(input)); err != nil {
		return nil, err
	}

	if err := xml.Unmarshal(input, &req); err != nil {
		return nil, err
	}

	authnRequest := &AuthnRequest{
		AuthnRequest: req,
	}

	// TODO(saml): Verify the signature

	if authnRequest.Destination != "" {
		if authnRequest.Destination != s.Endpoints.SAMLLoginURL(sp.ID).String() {
			return nil, fmt.Errorf("unexpected destination")
		}
	}

	if !authnRequest.GetProtocolBinding().IsSupported() {
		return nil, fmt.Errorf("unsupported binding")
	}

	if authnRequest.IssueInstant.Add(MaxAuthnRequestValidDuration).Before(now) {
		return nil, fmt.Errorf("request expired")
	}

	if authnRequest.Version != SAMLVersion2 {
		return nil, fmt.Errorf("Request Version must be 2.0")
	}

	if authnRequest.NameIDPolicy != nil && authnRequest.NameIDPolicy.Format != nil {
		reqNameIDFormat := *authnRequest.NameIDPolicy.Format
		if reqNameIDFormat != string(sp.NameIDFormat) &&
			// unspecified is always allowed
			reqNameIDFormat != string(config.NameIDFormatUnspecified) {
			return nil, fmt.Errorf("unsupported Name Identifier Format")
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
			return nil, fmt.Errorf("AssertionConsumerServiceURL not allowed")
		}
	}

	return authnRequest, nil
}
