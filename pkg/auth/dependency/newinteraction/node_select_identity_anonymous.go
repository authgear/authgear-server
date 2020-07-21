package newinteraction

import (
	"github.com/lestrrat-go/jwx/jwk"

	"github.com/authgear/authgear-server/pkg/auth/dependency/identity"
	"github.com/authgear/authgear-server/pkg/auth/dependency/identity/anonymous"
)

type InputSelectIdentityAnonymous interface {
	GetAnonymousRequestToken() string
}

type EdgeSelectIdentityAnonymous struct {
}

type NodeSelectIdentityAnonymous struct {
	// FIXME: use key set instead of single key for anonymous identities
	Identity *identity.Info          `json:"identity"`
	KeySet   *jwk.Set                `json:"key_set"`
	Action   anonymous.RequestAction `json:"action"`
}
