package policy

import (
	"github.com/skygeario/skygear-server/pkg/core/auth/authinfo"
	"github.com/skygeario/skygear-server/pkg/core/auth/session"
	"github.com/skygeario/skygear-server/pkg/core/model"
)

type ContextGetter interface {
	AccessKey() model.AccessKey
	AuthInfo() *authinfo.AuthInfo
	Session() *session.Session
}

type MemoryContextGetter struct {
	mAccessKey model.AccessKey
	mAuthInfo  *authinfo.AuthInfo
	mSession   *session.Session
}

func (g MemoryContextGetter) AccessKey() model.AccessKey {
	return g.mAccessKey
}

func (g MemoryContextGetter) AuthInfo() *authinfo.AuthInfo {
	return g.mAuthInfo
}

func (g MemoryContextGetter) Session() *session.Session {
	return g.mSession
}
