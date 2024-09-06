package saml

import (
	"bytes"
	"crypto/rsa"
	"fmt"
	"net/url"
	"text/template"
	"time"

	"github.com/beevik/etree"
	crewjamsaml "github.com/crewjam/saml"

	dsig "github.com/russellhaering/goxmldsig"

	"github.com/authgear/authgear-server/pkg/lib/authn/authenticationinfo"
	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/lib/oauth"
	"github.com/authgear/authgear-server/pkg/lib/oauth/oidc"
	"github.com/authgear/authgear-server/pkg/lib/saml/samlerror"
	"github.com/authgear/authgear-server/pkg/lib/saml/samlprotocol"
	"github.com/authgear/authgear-server/pkg/lib/session"
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

type SAMLUserInfoProvider interface {
	GetUserInfo(userID string, clientLike *oauth.ClientLike) (map[string]interface{}, error)
}

type Service struct {
	Clock                   clock.Clock
	AppID                   config.AppID
	SAMLEnvironmentConfig   config.SAMLEnvironmentConfig
	SAMLConfig              *config.SAMLConfig
	SAMLIdpSigningMaterials *config.SAMLIdpSigningMaterials
	Endpoints               SAMLEndpoints
	UserInfoProvider        SAMLUserInfoProvider
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

	ssoServices := []crewjamsaml.Endpoint{}

	for _, binding := range samlprotocol.SSOSupportedBindings {
		ssoServices = append(ssoServices, crewjamsaml.Endpoint{
			Binding:  string(binding),
			Location: s.Endpoints.SAMLLoginURL(sp.GetID()).String(),
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
				SingleSignOnServices: ssoServices,
			},
		},
	}

	return &samlprotocol.Metadata{
		EntityDescriptor: descriptor,
	}, nil
}

func (s *Service) validateDestination(sp *config.SAMLServiceProviderConfig, destination string) error {
	allowedDestinations := []string{}
	if sp.Deprecated_ID != "" {
		allowedDestinations = append(allowedDestinations, s.Endpoints.SAMLLoginURL(sp.Deprecated_ID).String())
	}
	if sp.ClientID != "" {
		allowedDestinations = append(allowedDestinations, s.Endpoints.SAMLLoginURL(sp.ClientID).String())
	}

	for _, allowedDestination := range allowedDestinations {
		if destination == allowedDestination {
			return nil
		}
	}
	return &samlerror.InvalidRequestError{
		Field:    "Destination",
		Actual:   destination,
		Expected: allowedDestinations,
		Reason:   "unexpected Destination",
	}

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
		err := s.validateDestination(sp, authnRequest.Destination)
		if err != nil {
			return err
		}
	}

	if !authnRequest.GetProtocolBinding().IsACSSupported() {
		return &samlerror.InvalidRequestError{
			Field:    "ProtocolBinding",
			Actual:   authnRequest.ProtocolBinding,
			Expected: slice.Map(samlprotocol.ACSSupportedBindings, func(b samlprotocol.SAMLBinding) string { return string(b) }),
			Reason:   "unsupported ProtocolBinding",
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
			Reason:   "unsupported Version",
		}
	}

	allowedAudiences := setutil.Set[string]{}

	// acs urls are always allowed
	for _, acsURL := range sp.AcsURLs {
		allowedAudiences.Add(acsURL)
	}
	if sp.Audience != "" {
		allowedAudiences.Add(sp.Audience)
	}

	for _, aud := range authnRequest.CollectAudiences() {
		if !allowedAudiences.Has(aud) {
			return &samlerror.InvalidRequestError{
				Field:    "Conditions/AudienceRestrictions",
				Actual:   aud,
				Expected: allowedAudiences.Keys(),
				Reason:   "Audience not allowed",
			}
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
				Reason:   "unsupported NameIDPolicy Format",
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
	authInfo authenticationinfo.T,
	inResponseToAuthnRequest *samlprotocol.AuthnRequest,
) (*samlprotocol.Response, error) {
	sp, ok := s.SAMLConfig.ResolveProvider(serviceProviderId)
	if !ok {
		return nil, samlerror.ErrServiceProviderNotFound
	}
	authenticatedUserId := authInfo.UserID
	sid := oidc.EncodeSIDByRawValues(
		session.Type(authInfo.AuthenticatedBySessionType),
		authInfo.AuthenticatedBySessionID,
	)

	clientLike := spToClientLike(sp)
	userInfo, err := s.UserInfoProvider.GetUserInfo(authenticatedUserId, clientLike)
	if err != nil {
		return nil, err
	}

	now := s.Clock.NowUTC()
	issuerID := s.IdpEntityID()
	inResponseTo := ""
	if inResponseToAuthnRequest != nil {
		inResponseTo = inResponseToAuthnRequest.ID
	}
	response := samlprotocol.NewSuccessResponse(now, issuerID, inResponseTo)

	// Referencing other SAML Idp implementations,
	// use ACS url as default value of destination, recipient and audience
	destination := callbackURL
	if sp.Destination != "" {
		destination = sp.Destination
	}
	response.Destination = destination

	recipient := callbackURL
	if sp.Recipient != "" {
		recipient = sp.Recipient
	}

	nameIDFormat := sp.NameIDFormat
	if inResponseToAuthnRequest != nil {
		if nameIDFormatInRequest, ok := inResponseToAuthnRequest.GetNameIDFormat(); ok {
			nameIDFormat = nameIDFormatInRequest
		}
	}

	// allow for some clock skew
	notBefore := now.Add(-1 * duration.ClockSkew)
	assertionValidDuration := sp.AssertionValidDuration.Duration()
	notOnOrAfter := now.Add(assertionValidDuration)
	if inResponseToAuthnRequest != nil && notBefore.Before(inResponseToAuthnRequest.IssueInstant) {
		notBefore = inResponseToAuthnRequest.IssueInstant
		notOnOrAfter = notBefore.Add(assertionValidDuration)
	}

	conditions := &samlprotocol.Conditions{
		NotBefore:    notBefore,
		NotOnOrAfter: notOnOrAfter,
	}
	if inResponseToAuthnRequest != nil && inResponseToAuthnRequest.Conditions != nil {
		// Only allow conditions which are stricter than what we set by default
		if !inResponseToAuthnRequest.Conditions.NotBefore.IsZero() && inResponseToAuthnRequest.Conditions.NotBefore.After(notBefore) {
			conditions.NotBefore = inResponseToAuthnRequest.Conditions.NotBefore
		}
		if !inResponseToAuthnRequest.Conditions.NotOnOrAfter.IsZero() && inResponseToAuthnRequest.Conditions.NotOnOrAfter.Before(notOnOrAfter) {
			conditions.NotOnOrAfter = inResponseToAuthnRequest.Conditions.NotOnOrAfter
		}
	}
	audiences := setutil.Set[string]{}
	// Callback url is always included
	audiences.Add(callbackURL)

	// Include audience set in config
	if sp.Audience != "" {
		audiences.Add(sp.Audience)
	}

	// Include audiences requested
	if inResponseToAuthnRequest != nil {
		for _, aud := range inResponseToAuthnRequest.CollectAudiences() {
			audiences.Add(aud)
		}
	}

	audienceRestriction := samlprotocol.AudienceRestriction{
		Audience: []samlprotocol.Audience{},
	}

	for _, aud := range audiences.Keys() {
		audienceRestriction.Audience = append(audienceRestriction.Audience,
			samlprotocol.Audience{
				Value: aud,
			},
		)
	}

	conditions.AudienceRestrictions = []samlprotocol.AudienceRestriction{
		audienceRestriction,
	}

	nameID, err := s.getUserNameID(nameIDFormat, sp, userInfo)
	if err != nil {
		return nil, err
	}

	assertion := &samlprotocol.Assertion{
		ID:           samlprotocol.GenerateAssertionID(),
		IssueInstant: now,
		Version:      samlprotocol.SAMLVersion2,
		Issuer: samlprotocol.Issuer{
			Format: samlprotocol.SAMLIssertFormatEntity,
			Value:  issuerID,
		},
		Subject: &samlprotocol.Subject{
			NameID: &samlprotocol.NameID{
				Format: string(nameIDFormat),
				Value:  nameID,
			},
			SubjectConfirmations: []samlprotocol.SubjectConfirmation{
				{
					Method: "urn:oasis:names:tc:SAML:2.0:cm:bearer",
					SubjectConfirmationData: &samlprotocol.SubjectConfirmationData{
						InResponseTo: inResponseTo,
						NotOnOrAfter: notOnOrAfter,
						Recipient:    recipient,
					},
				},
			},
		},
		Conditions: conditions,
		AuthnStatements: []samlprotocol.AuthnStatement{
			{
				AuthnInstant: notBefore,
				SessionIndex: sid,
				AuthnContext: samlprotocol.AuthnContext{
					AuthnContextClassRef: &samlprotocol.AuthnContextClassRef{
						// TODO(saml): Return a correct context by used authenticators
						Value: "urn:oasis:names:tc:SAML:2.0:ac:classes:unspecified",
					},
				},
			},
		},
		AttributeStatements: []samlprotocol.AttributeStatement{
			{
				// TODO(saml): Return more attributes
				Attributes: []samlprotocol.Attribute{
					{
						FriendlyName: "User ID",
						Name:         "sub",
						NameFormat:   samlprotocol.SAMLAttrnameFormatBasic,
						Values: []samlprotocol.AttributeValue{{
							Type:  samlprotocol.SAMLAttrTypeString,
							Value: authenticatedUserId,
						}},
					},
				},
			},
		},
	}

	response.Assertion = assertion

	err = s.signResponse(response)
	if err != nil {
		return nil, err
	}

	return response, nil
}

func (s *Service) getUserNameID(
	format config.SAMLNameIDFormat,
	sp *config.SAMLServiceProviderConfig,
	userInfo map[string]interface{},
) (string, error) {
	switch format {
	case config.SAMLNameIDFormatEmailAddress:
		{
			email, ok := userInfo["email"].(string)
			if !ok {
				return "", &samlerror.MissingNameIDError{
					ExpectedNameIDFormat: string(config.SAMLNameIDFormatEmailAddress),
				}
			}
			return email, nil
		}
	case config.SAMLNameIDFormatUnspecified:
		{
			jsonPointer := sp.NameIDAttributePointer.MustGetJSONPointer()
			nameID, err := jsonPointer.Traverse(userInfo)
			if err != nil {
				return "", &samlerror.MissingNameIDError{
					ExpectedNameIDFormat:   string(config.SAMLNameIDFormatUnspecified),
					NameIDAttributePointer: jsonPointer.String(),
				}
			}
			switch nameID := nameID.(type) {
			case string:
				return nameID, nil
			case float64:
				return fmt.Sprintf("%v", nameID), nil
			case bool:
				return fmt.Sprintf("%v", nameID), nil
			default:
				return "", &samlerror.MissingNameIDError{
					ExpectedNameIDFormat:   string(config.SAMLNameIDFormatUnspecified),
					NameIDAttributePointer: jsonPointer.String(),
				}
			}
		}
	default:
		panic(fmt.Errorf("unknown nameid format %s", format))

	}
}

func (s *Service) signResponse(response *samlprotocol.Response) error {
	// Sign the assertion
	assertionEl := response.Assertion.Element()

	assertionSigEl, err := s.constructSignature(assertionEl)
	if err != nil {
		return err
	}
	response.Assertion.Signature = assertionSigEl

	// Sign the response
	responseEl := response.Element()
	responseSigEl, err := s.constructSignature(responseEl)
	if err != nil {
		return err
	}
	response.Signature = responseSigEl

	return nil
}

func (s *Service) constructSignature(el *etree.Element) (*etree.Element, error) {
	signingContext, err := s.idpSigningContext()
	if err != nil {
		return nil, err
	}
	signature, err := signingContext.ConstructSignature(el, true)
	if err != nil {
		return nil, err
	}
	return signature, nil
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

	keyStore := &x509KeyStore{
		privateKey: &rsaPrivateKey,
		cert:       activeCert.Certificate.Data(),
	}

	signingContext = dsig.NewDefaultSigningContext(keyStore)

	signatureMethod := string(s.SAMLConfig.Signing.SignatureMethod)

	signingContext.Canonicalizer = dsig.MakeC14N10ExclusiveCanonicalizerWithPrefixList("")
	if err := signingContext.SetSignatureMethod(signatureMethod); err != nil {
		return nil, err
	}

	return signingContext, nil
}

func spToClientLike(sp *config.SAMLServiceProviderConfig) *oauth.ClientLike {
	// Note(tung): Note sure if there could be third party SAML apps in the future,
	// now it is always first party app.
	return &oauth.ClientLike{
		IsFirstParty:        true,
		PIIAllowedInIDToken: false,
		Scopes:              []string{},
	}
}

type x509KeyStore struct {
	privateKey *rsa.PrivateKey
	cert       []byte
}

var _ dsig.X509KeyStore = &x509KeyStore{}

func (x *x509KeyStore) GetKeyPair() (privateKey *rsa.PrivateKey, cert []byte, err error) {
	return x.privateKey, x.cert, nil
}
