package userverify

import (
	"testing"

	"github.com/skygeario/skygear-server/pkg/auth/dependency/provider/password"
	"github.com/skygeario/skygear-server/pkg/core/auth/authinfo"
	"github.com/skygeario/skygear-server/pkg/core/config"
	. "github.com/smartystreets/goconvey/convey"
)

func TestIsUserVerified(t *testing.T) {
	Convey("Test IsUserVerified", t, func() {
		verifyConfigs := map[string]config.UserVerificationKeyConfiguration{
			"email": config.UserVerificationKeyConfiguration{},
			"phone": config.UserVerificationKeyConfiguration{},
		}

		makeAuthInfo := func(verifiedLoginIDs []string) (authInfo *authinfo.AuthInfo) {
			authInfo = &authinfo.AuthInfo{
				ID:         "user-id",
				VerifyInfo: map[string]bool{},
			}
			for _, loginID := range verifiedLoginIDs {
				authInfo.VerifyInfo[loginID] = true
			}
			return
		}

		makePrincipals := func(loginIDs map[string]string) []*password.Principal {
			principals := []*password.Principal{}
			for key, value := range loginIDs {
				principals = append(principals, &password.Principal{
					LoginIDKey: key,
					LoginID:    value,
				})
			}
			return principals
		}

		Convey("With criteria = all", func() {
			var isVerified bool
			criteria := config.UserVerificationCriteriaAll

			isVerified = IsUserVerified(
				makeAuthInfo([]string{
					"test+1@example.com",
				}),
				makePrincipals(map[string]string{
					"email": "test+1@example.com",
				}),
				criteria, verifyConfigs,
			)
			So(isVerified, ShouldEqual, true)

			isVerified = IsUserVerified(
				makeAuthInfo([]string{
					"test+1@example.com",
				}),
				makePrincipals(map[string]string{
					"email": "test+1@example.com",
					"phone": "+85299999999",
				}),
				criteria, verifyConfigs,
			)
			So(isVerified, ShouldEqual, false)

			isVerified = IsUserVerified(
				makeAuthInfo([]string{}),
				makePrincipals(map[string]string{
					"phone": "+85299999999",
				}),
				criteria, verifyConfigs,
			)
			So(isVerified, ShouldEqual, false)

			isVerified = IsUserVerified(
				makeAuthInfo([]string{}),
				makePrincipals(map[string]string{
					"username": "test",
				}),
				criteria, verifyConfigs,
			)
			So(isVerified, ShouldEqual, false)
		})

		Convey("With criteria = any", func() {
			var isVerified bool
			criteria := config.UserVerificationCriteriaAny

			isVerified = IsUserVerified(
				makeAuthInfo([]string{
					"test+1@example.com",
				}),
				makePrincipals(map[string]string{
					"email": "test+1@example.com",
				}),
				criteria, verifyConfigs,
			)
			So(isVerified, ShouldEqual, true)

			isVerified = IsUserVerified(
				makeAuthInfo([]string{
					"test+1@example.com",
				}),
				makePrincipals(map[string]string{
					"email": "test+1@example.com",
					"phone": "+85299999999",
				}),
				criteria, verifyConfigs,
			)
			So(isVerified, ShouldEqual, true)

			isVerified = IsUserVerified(
				makeAuthInfo([]string{}),
				makePrincipals(map[string]string{
					"phone": "+85299999999",
				}),
				criteria, verifyConfigs,
			)
			So(isVerified, ShouldEqual, false)

			isVerified = IsUserVerified(
				makeAuthInfo([]string{}),
				makePrincipals(map[string]string{
					"username": "test",
				}),
				criteria, verifyConfigs,
			)
			So(isVerified, ShouldEqual, false)
		})

		Convey("With no key to verify", func() {
			var isVerified bool

			isVerified = IsUserVerified(
				makeAuthInfo([]string{}),
				makePrincipals(map[string]string{
					"email": "test+1@example.com",
				}),
				config.UserVerificationCriteriaAny,
				map[string]config.UserVerificationKeyConfiguration{},
			)
			So(isVerified, ShouldEqual, true)

			isVerified = IsUserVerified(
				makeAuthInfo([]string{}),
				makePrincipals(map[string]string{
					"email": "test+1@example.com",
				}),
				config.UserVerificationCriteriaAll,
				map[string]config.UserVerificationKeyConfiguration{},
			)
			So(isVerified, ShouldEqual, true)
		})
	})
}
