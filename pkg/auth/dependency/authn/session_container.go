package authn

import "github.com/skygeario/skygear-server/pkg/auth/dependency/session"

type SessionContainer interface {
	SessionAttrs() *session.Attrs
}
