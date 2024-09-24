package samlslosession

import (
	"fmt"

	"github.com/authgear/authgear-server/pkg/lib/saml/samlprotocol"
	"github.com/authgear/authgear-server/pkg/util/base32"
	"github.com/authgear/authgear-server/pkg/util/rand"
)

type SAMLSLOSession struct {
	ID    string               `json:"id,omitempty"`
	Entry *SAMLSLOSessionEntry `json:"entry,omitempty"`
}
type SAMLSLOSessionEntry struct {
	PendingLogoutServiceProviderIDs []string                 `json:"pending_logout_service_provider_ids,omitempty"`
	LogoutRequestXML                string                   `json:"logout_request_xml,omitempty"`
	ResponseBinding                 samlprotocol.SAMLBinding `json:"response_binding,omitempty"`
	CallbackURL                     string                   `json:"callback_url,omitempty"`
	RelayState                      string                   `json:"relay_state,omitempty"`
	SID                             string                   `json:"sid,omitempty"`
	UserID                          string                   `json:"user_id,omitempty"`
	IsPartialLogout                 bool                     `json:"is_partial_logout,omitempty"`
	PostLogoutRedirectURI           string                   `json:"post_logout_redirect_uri,omitempty"`
}

func NewSAMLSLOSession(entry *SAMLSLOSessionEntry) *SAMLSLOSession {
	id := rand.StringWithAlphabet(32, base32.Alphabet, rand.SecureRand)

	return &SAMLSLOSession{
		ID:    fmt.Sprintf("samlslosession_%s", id),
		Entry: entry,
	}
}

func (s *SAMLSLOSessionEntry) LogoutRequest() (*samlprotocol.LogoutRequest, bool) {
	if s.LogoutRequestXML == "" {
		return nil, false
	}
	r, err := samlprotocol.ParseLogoutRequest([]byte(s.LogoutRequestXML))
	if err != nil {
		panic(err)
	}
	return r, true
}
