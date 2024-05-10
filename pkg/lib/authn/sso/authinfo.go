package sso

import (
	"github.com/authgear/authgear-server/pkg/lib/authn/stdattrs"
)

type UserProfile struct {
	ProviderRawProfile map[string]interface{}
	// ProviderUserID is not necessarily equal to sub.
	// If there exists a more unique identifier than sub, that identifier is chosen instead.
	ProviderUserID     string
	StandardAttributes stdattrs.T
}
