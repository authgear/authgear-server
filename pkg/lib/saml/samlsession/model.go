package samlsession

import (
	"github.com/authgear/authgear-server/pkg/lib/saml"
	"github.com/authgear/authgear-server/pkg/util/base32"
	"github.com/authgear/authgear-server/pkg/util/rand"
)

type SAMLSession struct {
	ID              string `json:"id,omitempty"`
	AuthnRequestXML string `json:"authn_request_xml,omitempty"`
	// The url the response should send to
	CallbackURL string `json:"callback_url,omitempty"`
}

func NewSAMLSession(authnRequest *saml.AuthnRequest, callbackURL string) *SAMLSession {
	id := rand.StringWithAlphabet(32, base32.Alphabet, rand.SecureRand)

	return &SAMLSession{
		ID:              id,
		AuthnRequestXML: string(authnRequest.ToXMLBytes()),
		CallbackURL:     callbackURL,
	}
}

func (s *SAMLSession) AuthnRequest() *saml.AuthnRequest {
	r, err := saml.ParseAuthnRequest([]byte(s.AuthnRequestXML))
	if err != nil {
		// We should ensure only valid request stored in the session
		// So it is a panic if we got something invalid here
		panic(err)
	}
	return r
}
