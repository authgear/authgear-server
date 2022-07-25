package webauthn

import (
	"github.com/duo-labs/webauthn/protocol"
)

// Session is an object to associate a challenge with generated options.
// It is persisted in Redis.
type Session struct {
	Challenge       protocol.URLEncodedBase64 `json:"challenge"`
	CreationOptions *CreationOptions          `json:"creation_options,omitempty"`
	RequestOptions  *RequestOptions           `json:"request_options,omitempty"`
}
