package context

import (
	"github.com/skygeario/skygear-server/pkg/core/auth/authinfo"
	"github.com/skygeario/skygear-server/pkg/core/auth/authtoken"
	"github.com/skygeario/skygear-server/pkg/core/auth/role"
	"github.com/skygeario/skygear-server/pkg/core/model"
)

type AuthContext struct {
	AccessKeyType model.KeyType
	AuthInfo      *authinfo.AuthInfo
	Roles         []role.Role
	Token         *authtoken.Token
}
