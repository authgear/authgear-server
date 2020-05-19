package interaction

import (
	"github.com/skygeario/skygear-server/pkg/auth/dependency/identity"
	"github.com/skygeario/skygear-server/pkg/core/authn"
)

type Result struct {
	*authn.Attrs
	Identity identity.Info
}
