package password

import (
	"github.com/skygeario/skygear-server/pkg/auth/dependency/audit"
	"github.com/skygeario/skygear-server/pkg/core/skyerr"
	"github.com/skygeario/skygear-server/pkg/server/skydb"
)

type ResetPasswordRequestContext struct {
	PasswordChecker      *audit.PasswordChecker
	PasswordAuthProvider Provider
}

func (r *ResetPasswordRequestContext) ExecuteWithPrincipals(newPassword string, principals []*Principal) (err error) {
	if err = r.PasswordChecker.ValidatePassword(audit.ValidatePasswordPayload{
		PlainPassword: newPassword,
	}); err != nil {
		return
	}

	for _, p := range principals {
		p.PlainPassword = newPassword
		err = r.PasswordAuthProvider.UpdatePrincipal(*p)
		if err != nil {
			return
		}
	}

	return
}

func (r *ResetPasswordRequestContext) ExecuteWithUserID(newPassword string, userID string) (err error) {
	principals, err := r.PasswordAuthProvider.GetPrincipalsByUserID(userID)
	if err != nil {
		if err == skydb.ErrUserNotFound {
			err = skyerr.NewError(skyerr.ResourceNotFound, "user not found")
			return
		}

		return
	}

	err = r.ExecuteWithPrincipals(newPassword, principals)
	return
}
