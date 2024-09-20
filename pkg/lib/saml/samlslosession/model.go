package samlslosession

import (
	"fmt"

	"github.com/authgear/authgear-server/pkg/util/base32"
	"github.com/authgear/authgear-server/pkg/util/rand"
	"github.com/authgear/authgear-server/pkg/util/setutil"
)

type SAMLSLOSession struct {
	ID    string
	Entry *SAMLSLOSessionEntry
}
type SAMLSLOSessionEntry struct {
	PendingLogoutServiceProviderIDs setutil.Set[string]
	LogoutResponseXML               string
}

func NewSAMLSLOSession(entry *SAMLSLOSessionEntry) *SAMLSLOSession {
	id := rand.StringWithAlphabet(32, base32.Alphabet, rand.SecureRand)

	return &SAMLSLOSession{
		ID:    fmt.Sprintf("samlslosession_%s", id),
		Entry: entry,
	}
}
