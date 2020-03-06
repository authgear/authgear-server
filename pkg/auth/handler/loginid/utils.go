package loginid

import (
	"fmt"
	"regexp"
	"strconv"

	"github.com/skygeario/skygear-server/pkg/auth/dependency/loginid"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/principal/password"
	"github.com/skygeario/skygear-server/pkg/core/errors"
	"github.com/skygeario/skygear-server/pkg/core/validation"
)

func extractLoginIDs(principals []*password.Principal) []loginid.LoginID {
	loginIDs := make([]loginid.LoginID, len(principals))
	for i, p := range principals {
		loginIDs[i] = loginid.LoginID{Key: p.LoginIDKey, Value: p.LoginID}
	}
	return loginIDs
}

var loginIDPointerPrefixRegex = regexp.MustCompile(`^/(\d+)/`)

func validateLoginIDs(provider password.Provider, loginIDs []loginid.LoginID, newLoginIDBeginIndex int) error {
	err := provider.ValidateLoginIDs(loginIDs)
	if err != nil {
		if causes := validation.ErrorCauses(err); len(causes) > 0 {
			for i, cause := range causes {
				isNewLoginID := false

				matches := loginIDPointerPrefixRegex.FindStringSubmatch(cause.Pointer)
				if len(matches) > 0 && newLoginIDBeginIndex >= 0 {
					index, err := strconv.Atoi(matches[1])
					if err == nil && index >= newLoginIDBeginIndex {
						cause.Pointer = fmt.Sprintf("/%d/%s", index-newLoginIDBeginIndex, cause.Pointer[len(matches[0]):])
						isNewLoginID = true
					}
				}

				if !isNewLoginID {
					cause.Pointer = ""
				}
				causes[i] = cause
			}
			err = validation.NewValidationFailed("invalid login ID", causes)
		} else {
			err = errors.HandledWithMessage(err, "invalid login ID")
		}
		return err
	}
	return nil
}
