package handler

import (
	"github.com/skygeario/skygear-server/pkg/server/skydb"

	"github.com/skygeario/skygear-server/pkg/server/authtoken"
)

type AuthenticationContext struct {
	Token    *authtoken.Token
	AuthInfo *skydb.AuthInfo
}
