package provider

import (
	"github.com/skygeario/skygear-server/pkg/server/audit"
	"github.com/skygeario/skygear-server/pkg/server/skyerr"
)

type PasswordChecker interface {
	ValidatePassword(payload audit.ValidatePasswordPayload) skyerr.Error
}
