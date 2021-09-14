package sso

import (
	"github.com/authgear/authgear-server/pkg/lib/authn/stdattrs"
)

type AuthInfo struct {
	ProviderRawProfile map[string]interface{}
	StandardAttributes stdattrs.T
}
