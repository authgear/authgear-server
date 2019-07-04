package userverify

import (
	"fmt"

	"github.com/sirupsen/logrus"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/provider/password"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/userverify"
	"github.com/skygeario/skygear-server/pkg/core/skyerr"
)

type getAndValidateCodeRequest struct {
	VerifyCodeStore      userverify.Store
	PasswordAuthProvider password.Provider
	Logger               *logrus.Entry
}

func (g *getAndValidateCodeRequest) execute(
	userID string,
	code string,
) (verifyCode userverify.VerifyCode, err error) {
	if verifyCode, err = g.VerifyCodeStore.GetVerifyCodeByCode(userID, code); err != nil {
		g.Logger.WithFields(map[string]interface{}{
			"user_id": userID,
			"code":    code,
			"error":   err,
		}).Error("failed to get verify code")
		err = g.invalidCodeError(code, userID)
		return
	}
	if verifyCode.Consumed {
		g.Logger.WithField("code", code).Error("code has been consumed")
		err = g.invalidCodeError(code, userID)
		return
	}

	principals, err := g.PasswordAuthProvider.GetPrincipalsByLoginID(verifyCode.LoginIDKey, verifyCode.LoginID)
	if err == nil {
		// filter principals belonging to the user
		userPrincipals := []*password.Principal{}
		for _, principal := range principals {
			if principal.UserID == userID {
				userPrincipals = append(userPrincipals, principal)
			}
		}
		principals = userPrincipals
	}

	if err != nil || len(principals) == 0 {
		err = skyerr.NewError(
			skyerr.InvalidArgument,
			"the login ID to verify does not belong to the user",
		)
		return
	}

	// TODO: code expiry
	/*
		if code.ExpireAt() != nil && timeNow().After(*code.ExpireAt()) {
			err = skyerr.NewError(skyerr.InvalidArgument, "the code has expired")
			return
		}
	*/

	return
}

func (g *getAndValidateCodeRequest) invalidCodeError(code string, userID string) error {
	msg := fmt.Sprintf("the code `%s` is not valid for user `%s`", code, userID)
	return skyerr.NewInvalidArgument(msg, []string{"code"})
}
