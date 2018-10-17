package handler

import (
	"time"

	"github.com/skygeario/skygear-server/pkg/core/auth/authinfo"
	"github.com/skygeario/skygear-server/pkg/server/skyerr"
	"github.com/skygeario/skygear-server/pkg/server/uuid"
)

var (
	uuidNew = uuid.New
	timeNow = func() time.Time { return time.Now().UTC() }
)

// checkUserIsNotDisabled is used by login handlers to check if the user is
// not disabled.
func checkUserIsNotDisabled(authInfo *authinfo.AuthInfo) error {
	if authInfo.IsDisabled() {
		info := map[string]interface{}{}
		if authInfo.DisabledExpiry != nil {
			info["expiry"] = authInfo.DisabledExpiry.Format(time.RFC3339)
		}
		if authInfo.DisabledMessage != "" {
			info["message"] = authInfo.DisabledMessage
		}
		return skyerr.NewErrorWithInfo(skyerr.UserDisabled, "user is disabled", info)
	}

	authInfo.RefreshDisabledStatus()
	return nil
}
