package user

import (
	"time"

	"github.com/skygeario/skygear-server/pkg/auth/dependency/userprofile"
	"github.com/skygeario/skygear-server/pkg/auth/model"
	"github.com/skygeario/skygear-server/pkg/core/auth/authinfo"
)

func newUser(now time.Time, authInfo *authinfo.AuthInfo, userProfile *userprofile.UserProfile) *model.User {
	return &model.User{
		ID:               authInfo.ID,
		CreatedAt:        userProfile.CreatedAt,
		LastLoginAt:      authInfo.LastLoginAt,
		Verified:         authInfo.IsVerified(),
		ManuallyVerified: authInfo.ManuallyVerified,
		Disabled:         authInfo.IsDisabled(now),
		VerifyInfo:       authInfo.VerifyInfo,
		Metadata:         userProfile.Data,
	}
}
