package dependency

import (
	"github.com/skygeario/skygear-server/pkg/core/audit"
	"github.com/skygeario/skygear-server/pkg/server/skyerr"
)

type PasswordChecker interface {
	ValidatePassword(payload audit.ValidatePasswordPayload) skyerr.Error
}
