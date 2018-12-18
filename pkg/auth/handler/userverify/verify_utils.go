package userverify

import (
	"fmt"

	"github.com/sirupsen/logrus"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/userprofile"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/userverify"
	"github.com/skygeario/skygear-server/pkg/server/skyerr"
)

type getAndValidateCodeRequest struct {
	VerifyCodeStore userverify.Store
	Logger          *logrus.Entry
}

func (g *getAndValidateCodeRequest) execute(
	codeStr string,
	userProfile userprofile.UserProfile,
) (code userverify.VerifyCode, err error) {
	if err = g.VerifyCodeStore.GetVerifyCodeByCode(codeStr, &code); err != nil {
		g.Logger.WithFields(map[string]interface{}{
			"code":  codeStr,
			"error": err,
		}).Error("failed to get verify code")
		err = g.invalidCodeError(codeStr, userProfile.RecordID)
		return
	}
	if code.Consumed {
		g.Logger.WithField("code", codeStr).Error("code has been consumed")
		err = g.invalidCodeError(codeStr, userProfile.RecordID)
		return
	}

	if userProfile.Data[code.RecordKey] != code.RecordValue {
		err = skyerr.NewError(
			skyerr.InvalidArgument,
			"the user data has since been modified, a new verification is required",
		)
		return
	}

	if code.ExpireAt() != nil && timeNow().After(*code.ExpireAt()) {
		err = skyerr.NewError(skyerr.InvalidArgument, "the code has expired")
		return
	}

	return
}

func (g *getAndValidateCodeRequest) invalidCodeError(code string, userID string) error {
	msg := fmt.Sprintf("the code `%s` is not valid for user `%s`", code, userID)
	return skyerr.NewInvalidArgument(msg, []string{"code"})
}
