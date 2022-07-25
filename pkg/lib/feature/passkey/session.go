package passkey

import (
	"github.com/duo-labs/webauthn/protocol"

	"github.com/authgear/authgear-server/pkg/api/model"
)

// Session is an object to associate a challenge with generated options.
// It is persisted in Redis.
type Session struct {
	Challenge       protocol.URLEncodedBase64      `json:"challenge"`
	CreationOptions *model.WebAuthnCreationOptions `json:"creation_options,omitempty"`
	RequestOptions  *model.WebAuthnRequestOptions  `json:"request_options,omitempty"`
}
