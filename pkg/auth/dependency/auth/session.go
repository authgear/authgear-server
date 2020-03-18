package auth

import (
	"github.com/skygeario/skygear-server/pkg/auth/model"
	"github.com/skygeario/skygear-server/pkg/core/authn"
)

const (
	SessionTypeIdentityProvider authn.SessionType = "idp"
)

type Session interface {
	authn.Session
	ToAPIModel() *model.Session
}
