package forgotpwd

import (
	"crypto/subtle"
	"time"

	"github.com/skygeario/skygear-server/pkg/auth/dependency/audit"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/forgotpwdemail"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/principal/password"
	"github.com/skygeario/skygear-server/pkg/core/auth/authinfo"
	"github.com/skygeario/skygear-server/pkg/core/skyerr"
)

type passwordReseter struct {
	CodeGenerator        *forgotpwdemail.CodeGenerator
	PasswordChecker      *audit.PasswordChecker
	AuthInfoStore        authinfo.Store
	PasswordAuthProvider password.Provider
}

func (r passwordReseter) resetPassword(userID string, now time.Time, expiry time.Time, code string, newPass string) error {
	// check code expiration
	if now.After(expiry) {
		return NewPasswordResetFailed(ExpiredCode, "reset code has expired")
	}

	authInfo := authinfo.AuthInfo{}
	if err := r.AuthInfoStore.GetAuth(userID, &authInfo); err != nil {
		if skyerr.IsKind(err, authinfo.UserNotFound) {
			return NewPasswordResetFailed(InvalidCode, "invalid reset code")
		}
		return err
	}

	// Get password auth principals
	principals, err := r.PasswordAuthProvider.GetPrincipalsByUserID(authInfo.ID)
	if err != nil {
		return err
	}

	if len(principals) == 0 {
		return NewPasswordResetFailed(InvalidCode, "invalid reset code")
	}

	// Get user email from loginIDs
	hashedPassword := principals[0].HashedPassword
	expectedCode := r.CodeGenerator.Generate(authInfo, hashedPassword, expiry)
	if subtle.ConstantTimeCompare([]byte(code), []byte(expectedCode)) == 0 {
		return NewPasswordResetFailed(InvalidCode, "invalid reset code")
	}

	resetPwdCtx := password.ResetPasswordRequestContext{
		PasswordChecker:      r.PasswordChecker,
		PasswordAuthProvider: r.PasswordAuthProvider,
	}

	if err := resetPwdCtx.ExecuteWithPrincipals(newPass, principals); err != nil {
		return err
	}

	return nil
}
