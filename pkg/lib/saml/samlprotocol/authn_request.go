package samlprotocol

import (
	"bytes"
	"encoding/xml"

	xrv "github.com/mattermost/xml-roundtrip-validator"
)

func (a *AuthnRequest) GetProtocolBinding() SAMLBinding {
	return SAMLBinding(a.ProtocolBinding)
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

func (a *AuthnRequest) GetNameIDFormat() (SAMLNameIDFormat, bool) {
	if a.NameIDPolicy != nil && a.NameIDPolicy.Format != nil {
		return SAMLNameIDFormat(*a.NameIDPolicy.Format), true
	}
	return "", false
}

func (a *AuthnRequest) CollectAudiences() (audiences []string) {
	audiences = []string{}
	if a.Conditions == nil {
		return
	}
	if len(a.Conditions.AudienceRestrictions) > 0 {
		for _, r := range a.Conditions.AudienceRestrictions {
			for _, aud := range r.Audience {
				audiences = append(audiences, aud.Value)
			}
		}
	}
	return
}

func (a *AuthnRequest) ToXMLBytes() []byte {
	buf, err := xml.Marshal(a)
	if err != nil {
		panic(err)
	}
	return buf
}

func ParseAuthnRequest(input []byte) (*AuthnRequest, error) {
	var req AuthnRequest
	if err := xrv.Validate(bytes.NewReader(input)); err != nil {
		return nil, err
	}

	if err := xml.Unmarshal(input, &req); err != nil {
		return nil, err
	}

	return &req, nil
}
