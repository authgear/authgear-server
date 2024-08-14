package saml

import (
	"bytes"
	"encoding/xml"

	crewjamsaml "github.com/crewjam/saml"
	xrv "github.com/mattermost/xml-roundtrip-validator"

	"github.com/authgear/authgear-server/pkg/lib/saml/binding"
)

type AuthnRequest struct {
	crewjamsaml.AuthnRequest
}

func (a *AuthnRequest) GetProtocolBinding() binding.SAMLBinding {
	return binding.SAMLBinding(a.AuthnRequest.ProtocolBinding)
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
