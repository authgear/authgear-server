package user

import (
	"time"

	"github.com/skygeario/skygear-server/pkg/auth/dependency/identity"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/userprofile"
	"github.com/skygeario/skygear-server/pkg/auth/model"
	"github.com/skygeario/skygear-server/pkg/core/auth/authinfo"
	"github.com/skygeario/skygear-server/pkg/core/authn"
)

func newUser(
	now time.Time,
	authInfo *authinfo.AuthInfo,
	userProfile *userprofile.UserProfile,
	identities []*identity.Info,
) *model.User {
	isAnonymous := false
	for _, i := range identities {
		if i.Type == authn.IdentityTypeAnonymous {
			isAnonymous = true
			break
		}
	}

	return &model.User{
		ID:               authInfo.ID,
		CreatedAt:        userProfile.CreatedAt,
		LastLoginAt:      authInfo.LastLoginAt,
		Verified:         authInfo.IsVerified(),
		ManuallyVerified: authInfo.ManuallyVerified,
		Disabled:         authInfo.IsDisabled(now),
		IsAnonymous:      isAnonymous,
		VerifyInfo:       authInfo.VerifyInfo,
		Metadata:         userProfile.Data,
	}
}
