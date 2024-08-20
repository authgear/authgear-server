package samlsession

import (
	"fmt"

	"github.com/authgear/authgear-server/pkg/lib/saml/samlprotocol"
	"github.com/authgear/authgear-server/pkg/util/base32"
	"github.com/authgear/authgear-server/pkg/util/rand"
)

type SAMLSessionEntry struct {
	ServiceProviderID string `json:"service_provider_id,omitempty"`
	AuthnRequestXML   string `json:"authn_request_xml,omitempty"`
	// The url the response should send to
	CallbackURL string `json:"callback_url,omitempty"`
}

type SAMLSession struct {
	ID     string            `json:"id,omitempty"`
	Entry  *SAMLSessionEntry `json:"entry,omitempty"`
	UIInfo *SAMLUIInfo       `json:"ui_info,omitempty"`
}

func NewSAMLSession(entry *SAMLSessionEntry, uiInfo *SAMLUIInfo) *SAMLSession {
	id := rand.StringWithAlphabet(32, base32.Alphabet, rand.SecureRand)

	return &SAMLSession{
		ID:     fmt.Sprintf("samlsession_%s", id),
		Entry:  entry,
		UIInfo: uiInfo,
	}
}

func (s *SAMLSessionEntry) AuthnRequest() *samlprotocol.AuthnRequest {
	r, err := samlprotocol.ParseAuthnRequest([]byte(s.AuthnRequestXML))
	if err != nil {
		// We should ensure only valid request stored in the session
		// So it is a panic if we got something invalid here
		panic(err)
	}
	return r
}
