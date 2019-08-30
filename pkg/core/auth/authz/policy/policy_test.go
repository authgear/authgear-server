package policy

import (
	"github.com/skygeario/skygear-server/pkg/core/auth/authinfo"
	"github.com/skygeario/skygear-server/pkg/core/auth/session"
	"github.com/skygeario/skygear-server/pkg/core/model"
)

type ContextGetter interface {
	AccessKeyType() model.KeyType
	AuthInfo() *authinfo.AuthInfo
	Session() *session.Session
}

type MemoryContextGetter struct {
	mAccessKeyType model.KeyType
	mAuthInfo      *authinfo.AuthInfo
	mSession       *session.Session
}

func (g MemoryContextGetter) AccessKeyType() model.KeyType {
	return g.mAccessKeyType
}

func (g MemoryContextGetter) AuthInfo() *authinfo.AuthInfo {
	return g.mAuthInfo
}

func (g MemoryContextGetter) Session() *session.Session {
	return g.mSession
}
