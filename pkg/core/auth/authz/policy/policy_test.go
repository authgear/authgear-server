package policy

import (
	"github.com/skygeario/skygear-server/pkg/core/auth"
	"github.com/skygeario/skygear-server/pkg/core/auth/authinfo"
	"github.com/skygeario/skygear-server/pkg/core/model"
)

type ContextGetter interface {
	AccessKey() model.AccessKey
	AuthInfo() *authinfo.AuthInfo
	Session() *auth.Session
}

type MemoryContextGetter struct {
	mAccessKey model.AccessKey
	mAuthInfo  *authinfo.AuthInfo
	mSession   *auth.Session
}

func (g MemoryContextGetter) AccessKey() model.AccessKey {
	return g.mAccessKey
}

func (g MemoryContextGetter) AuthInfo() (*authinfo.AuthInfo, error) {
	return g.mAuthInfo, nil
}

func (g MemoryContextGetter) Session() (*auth.Session, error) {
	return g.mSession, nil
}
