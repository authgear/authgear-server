package saml

import (
	"bytes"
	"encoding/xml"

	"github.com/authgear/authgear-server/pkg/lib/saml/samlbinding"
	crewjamsaml "github.com/crewjam/saml"
	xrv "github.com/mattermost/xml-roundtrip-validator"
)

type AuthnRequest struct {
	crewjamsaml.AuthnRequest
}

func (a *AuthnRequest) GetProtocolBinding() samlbinding.SAMLBinding {
	return samlbinding.SAMLBinding(a.AuthnRequest.ProtocolBinding)
}

func (a *AuthnRequest) GetIsPassive() bool {
	if a.IsPassive == nil {
		// Default false, See 3.4.1 of https://docs.oasis-open.org/security/saml/v2.0/saml-core-2.0-os.pdf
		return false
	}
	return *a.IsPassive
}

func (a *AuthnRequest) GetForceAuthn() bool {
	if a.ForceAuthn == nil {
		// Default false, See 3.4.1 of https://docs.oasis-open.org/security/saml/v2.0/saml-core-2.0-os.pdf
		return false
	}
	return *a.ForceAuthn
}

func (a *AuthnRequest) ToXMLBytes() []byte {
	buf, err := xml.Marshal(a)
	if err != nil {
		panic(err)
	}
	return buf
}

func ParseAuthnRequest(input []byte) (*AuthnRequest, error) {
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

	return authnRequest, nil
}
