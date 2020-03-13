package policy

import (
	"github.com/skygeario/skygear-server/pkg/core/auth"
	"github.com/skygeario/skygear-server/pkg/core/auth/authinfo"
)

type ContextGetter interface {
	AuthInfo() *authinfo.AuthInfo
	Session() *auth.Session
}

type MemoryContextGetter struct {
	mAuthInfo *authinfo.AuthInfo
	mSession  *auth.Session
}

func (g MemoryContextGetter) AuthInfo() (*authinfo.AuthInfo, error) {
	return g.mAuthInfo, nil
}

func (g MemoryContextGetter) Session() (*auth.Session, error) {
	return g.mSession, nil
}
