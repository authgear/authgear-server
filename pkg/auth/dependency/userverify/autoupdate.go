package userverify

import (
	"fmt"

	"github.com/skygeario/skygear-server/pkg/core/auth/authinfo"
	"github.com/skygeario/skygear-server/pkg/core/config"
)

type AutoUpdateUserVerifyFunc func(*authinfo.AuthInfo)

func CreateAutoUpdateUserVerifyfunc(tConfig config.TenantConfiguration) AutoUpdateUserVerifyFunc {
	if !tConfig.UserVerify.AutoUpdate {
		return nil
	}

	switch tConfig.UserVerify.Criteria {
	case "all":
		return func(authInfo *authinfo.AuthInfo) {
			allVerified := true
			for _, key := range tConfig.UserVerify.Keys {
				if !authInfo.VerifyInfo[key] {
					allVerified = false
					break
				}
			}

			authInfo.Verified = allVerified
		}
	case "any":
		return func(authInfo *authinfo.AuthInfo) {
			for _, key := range tConfig.UserVerify.Keys {
				if authInfo.VerifyInfo[key] {
					authInfo.Verified = true
					return
				}
			}

			authInfo.Verified = false
		}
	default:
		panic(fmt.Errorf("unexpected verify criteria `%s`", tConfig.UserVerify.Criteria))
	}
}
