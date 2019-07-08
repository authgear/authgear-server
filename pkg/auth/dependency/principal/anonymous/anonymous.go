package anonymous

import (
	"github.com/skygeario/skygear-server/pkg/auth/dependency/principal"
)

const providerAnonymous string = "anonymous"

type Provider interface {
	principal.Provider
	CreatePrincipal(principal Principal) error
}
