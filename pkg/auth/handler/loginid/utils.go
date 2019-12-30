package loginid

import (
	"fmt"
	"strings"

	"github.com/skygeario/skygear-server/pkg/auth/dependency/principal/password"
	"github.com/skygeario/skygear-server/pkg/core/errors"
	"github.com/skygeario/skygear-server/pkg/core/validation"
)

func extractLoginIDs(principals []*password.Principal) []password.LoginID {
	loginIDs := make([]password.LoginID, len(principals))
	for i, p := range principals {
		loginIDs[i] = password.LoginID{Key: p.LoginIDKey, Value: p.LoginID}
	}
	return loginIDs
}

func validateLoginIDs(provider password.Provider, loginIDs []password.LoginID, newLoginIDIndex int) error {
	removePointerPrefix := ""
	if newLoginIDIndex >= 0 {
		removePointerPrefix = fmt.Sprintf("/%d/", newLoginIDIndex)
	}

	err := provider.ValidateLoginIDs(loginIDs)
	if err != nil {
		if causes := validation.ErrorCauses(err); len(causes) > 0 {
			for i, cause := range causes {
				if removePointerPrefix != "" && strings.HasPrefix(cause.Pointer, removePointerPrefix) {
					cause.Pointer = "/" + cause.Pointer[len(removePointerPrefix):]
				} else {
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
