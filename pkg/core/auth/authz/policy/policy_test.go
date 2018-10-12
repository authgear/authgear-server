package policy

import (
	"github.com/skygeario/skygear-server/pkg/core/auth/authinfo"
	"github.com/skygeario/skygear-server/pkg/core/auth/authtoken"
	"github.com/skygeario/skygear-server/pkg/core/auth/role"
	"github.com/skygeario/skygear-server/pkg/core/model"
)

type ContextGetter interface {
	AccessKeyType() model.KeyType
	AuthInfo() *authinfo.AuthInfo
	Roles() []role.Role
	Token() *authtoken.Token
}

type MemoryContextGetter struct {
	mAccessKeyType model.KeyType
	mAuthInfo      *authinfo.AuthInfo
	mRoles         []role.Role
	mToken         *authtoken.Token
}

func (g MemoryContextGetter) AccessKeyType() model.KeyType {
	return g.mAccessKeyType
}

func (g MemoryContextGetter) AuthInfo() *authinfo.AuthInfo {
	return g.mAuthInfo
}

func (g MemoryContextGetter) Roles() []role.Role {
	return g.mRoles
}

func (g MemoryContextGetter) Token() *authtoken.Token {
	return g.mToken
}
