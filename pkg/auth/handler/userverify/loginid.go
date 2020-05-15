package userverify

import (
	"github.com/skygeario/skygear-server/pkg/auth/dependency/userverify"
	"github.com/skygeario/skygear-server/pkg/core/auth/metadata"
)

type LoginIDProvider interface {
	userverify.LoginIDProvider
	IsLoginIDKeyType(loginIDKey string, loginIDKeyType metadata.StandardKey) bool
}
