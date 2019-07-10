package userverify

import (
	"strings"
	"testing"

	"github.com/skygeario/skygear-server/pkg/auth/dependency/provider/password"
	"github.com/skygeario/skygear-server/pkg/core/auth/authinfo"
	"github.com/skygeario/skygear-server/pkg/core/config"
	. "github.com/smartystreets/goconvey/convey"
)

func TestIsUserVerified(t *testing.T) {
	Convey("IsUserVerified", t, func() {
		type verifyRequest struct {
			LoginIDs          map[string]string
			VerifiedLoginIDs  []string
			VerifyLoginIDKeys []string
		}
		type verifyResult struct {
			All bool
			Any bool
		}

		isUserVerified := func(request verifyRequest) verifyResult {
			verifyConfigs := map[string]config.UserVerificationKeyConfiguration{}
			for _, loginIDKey := range request.VerifyLoginIDKeys {
				verifyConfigs[loginIDKey] = config.UserVerificationKeyConfiguration{}
			}

			authInfo := authinfo.AuthInfo{
				ID:         "user-id",
				VerifyInfo: map[string]bool{},
			}
			for _, loginID := range request.VerifiedLoginIDs {
				authInfo.VerifyInfo[loginID] = true
			}

			principals := []*password.Principal{}
			for key, value := range request.LoginIDs {
				for _, loginID := range strings.Split(value, " ") {
					principals = append(principals, &password.Principal{
						LoginIDKey: key,
						LoginID:    loginID,
					})
				}
			}

			return verifyResult{
				All: IsUserVerified(&authInfo, principals, config.UserVerificationCriteriaAll, verifyConfigs),
				Any: IsUserVerified(&authInfo, principals, config.UserVerificationCriteriaAny, verifyConfigs),
			}
		}

		Convey("When no keys to verify", func() {
			Convey("should check empty login IDs", func() {
				So(isUserVerified(verifyRequest{
					LoginIDs:          map[string]string{},
					VerifiedLoginIDs:  []string{},
					VerifyLoginIDKeys: []string{},
				}), ShouldResemble, verifyResult{All: false, Any: false})
			})
			Convey("should check single email login ID", func() {
				So(isUserVerified(verifyRequest{
					LoginIDs:          map[string]string{"email": "a@example.com"},
					VerifiedLoginIDs:  []string{},
					VerifyLoginIDKeys: []string{},
				}), ShouldResemble, verifyResult{All: false, Any: false})
			})
			Convey("should check multiple email login ID", func() {
				So(isUserVerified(verifyRequest{
					LoginIDs:          map[string]string{"email": "a@example.com b@example.com"},
					VerifiedLoginIDs:  []string{},
					VerifyLoginIDKeys: []string{},
				}), ShouldResemble, verifyResult{All: false, Any: false})
			})
		})

		Convey("When need to verify email", func() {
			Convey("should check empty login IDs", func() {
				So(isUserVerified(verifyRequest{
					LoginIDs:          map[string]string{},
					VerifiedLoginIDs:  []string{},
					VerifyLoginIDKeys: []string{"email"},
				}), ShouldResemble, verifyResult{All: false, Any: false})

				So(isUserVerified(verifyRequest{
					LoginIDs:          map[string]string{"username": "test"},
					VerifiedLoginIDs:  []string{},
					VerifyLoginIDKeys: []string{"email"},
				}), ShouldResemble, verifyResult{All: false, Any: false})
			})
			Convey("should check single email login ID", func() {
				So(isUserVerified(verifyRequest{
					LoginIDs:          map[string]string{"email": "a@example.com"},
					VerifiedLoginIDs:  []string{},
					VerifyLoginIDKeys: []string{"email"},
				}), ShouldResemble, verifyResult{All: false, Any: false})

				So(isUserVerified(verifyRequest{
					LoginIDs:          map[string]string{"email": "a@example.com", "username": "test"},
					VerifiedLoginIDs:  []string{},
					VerifyLoginIDKeys: []string{"email"},
				}), ShouldResemble, verifyResult{All: false, Any: false})

				So(isUserVerified(verifyRequest{
					LoginIDs:          map[string]string{"email": "a@example.com", "username": "test"},
					VerifiedLoginIDs:  []string{"a@example.com"},
					VerifyLoginIDKeys: []string{"email"},
				}), ShouldResemble, verifyResult{All: true, Any: true})
			})
			Convey("should check multiple email login IDs", func() {
				So(isUserVerified(verifyRequest{
					LoginIDs:          map[string]string{"email": "a@example.com b@example.com"},
					VerifiedLoginIDs:  []string{},
					VerifyLoginIDKeys: []string{"email"},
				}), ShouldResemble, verifyResult{All: false, Any: false})

				So(isUserVerified(verifyRequest{
					LoginIDs:          map[string]string{"email": "a@example.com b@example.com"},
					VerifiedLoginIDs:  []string{"a@example.com"},
					VerifyLoginIDKeys: []string{"email"},
				}), ShouldResemble, verifyResult{All: false, Any: true})

				So(isUserVerified(verifyRequest{
					LoginIDs:          map[string]string{"email": "a@example.com b@example.com"},
					VerifiedLoginIDs:  []string{"a@example.com", "b@example.com"},
					VerifyLoginIDKeys: []string{"email"},
				}), ShouldResemble, verifyResult{All: true, Any: true})
			})
		})

		Convey("When need to verify email & phone", func() {
			Convey("should check empty login IDs", func() {
				So(isUserVerified(verifyRequest{
					LoginIDs:          map[string]string{},
					VerifiedLoginIDs:  []string{},
					VerifyLoginIDKeys: []string{"email", "phone"},
				}), ShouldResemble, verifyResult{All: false, Any: false})
			})
			Convey("should check email/phone login ID", func() {
				So(isUserVerified(verifyRequest{
					LoginIDs:          map[string]string{"email": "a@example.com"},
					VerifiedLoginIDs:  []string{},
					VerifyLoginIDKeys: []string{"email", "phone"},
				}), ShouldResemble, verifyResult{All: false, Any: false})

				So(isUserVerified(verifyRequest{
					LoginIDs:          map[string]string{"email": "a@example.com"},
					VerifiedLoginIDs:  []string{"a@example.com"},
					VerifyLoginIDKeys: []string{"email", "phone"},
				}), ShouldResemble, verifyResult{All: true, Any: true})

				So(isUserVerified(verifyRequest{
					LoginIDs:          map[string]string{"phone": "+85299999999"},
					VerifiedLoginIDs:  []string{"+85299999999"},
					VerifyLoginIDKeys: []string{"email", "phone"},
				}), ShouldResemble, verifyResult{All: true, Any: true})
			})
			Convey("should check email & phone login ID", func() {
				So(isUserVerified(verifyRequest{
					LoginIDs:          map[string]string{"email": "a@example.com", "phone": "+85299999999"},
					VerifiedLoginIDs:  []string{},
					VerifyLoginIDKeys: []string{"email", "phone"},
				}), ShouldResemble, verifyResult{All: false, Any: false})
				So(isUserVerified(verifyRequest{
					LoginIDs:          map[string]string{"email": "a@example.com", "phone": "+85299999999"},
					VerifiedLoginIDs:  []string{"a@example.com"},
					VerifyLoginIDKeys: []string{"email", "phone"},
				}), ShouldResemble, verifyResult{All: false, Any: true})
				So(isUserVerified(verifyRequest{
					LoginIDs:          map[string]string{"email": "a@example.com", "phone": "+85299999999"},
					VerifiedLoginIDs:  []string{"+85299999999"},
					VerifyLoginIDKeys: []string{"email", "phone"},
				}), ShouldResemble, verifyResult{All: false, Any: true})
				So(isUserVerified(verifyRequest{
					LoginIDs:          map[string]string{"email": "a@example.com", "phone": "+85299999999"},
					VerifiedLoginIDs:  []string{"a@example.com", "+85299999999"},
					VerifyLoginIDKeys: []string{"email", "phone"},
				}), ShouldResemble, verifyResult{All: true, Any: true})
			})
			Convey("should check multiple login IDs", func() {
				So(isUserVerified(verifyRequest{
					LoginIDs:          map[string]string{"email": "a@example.com b@example.com"},
					VerifiedLoginIDs:  []string{},
					VerifyLoginIDKeys: []string{"email", "phone"},
				}), ShouldResemble, verifyResult{All: false, Any: false})

				So(isUserVerified(verifyRequest{
					LoginIDs:          map[string]string{"email": "a@example.com b@example.com", "phone": "+85299999999"},
					VerifiedLoginIDs:  []string{},
					VerifyLoginIDKeys: []string{"email", "phone"},
				}), ShouldResemble, verifyResult{All: false, Any: false})

				So(isUserVerified(verifyRequest{
					LoginIDs:          map[string]string{"email": "a@example.com b@example.com"},
					VerifiedLoginIDs:  []string{"a@example.com"},
					VerifyLoginIDKeys: []string{"email", "phone"},
				}), ShouldResemble, verifyResult{All: false, Any: true})

				So(isUserVerified(verifyRequest{
					LoginIDs:          map[string]string{"email": "a@example.com b@example.com", "phone": "+85299999999"},
					VerifiedLoginIDs:  []string{"a@example.com"},
					VerifyLoginIDKeys: []string{"email", "phone"},
				}), ShouldResemble, verifyResult{All: false, Any: true})

				So(isUserVerified(verifyRequest{
					LoginIDs:          map[string]string{"email": "a@example.com b@example.com"},
					VerifiedLoginIDs:  []string{"a@example.com", "b@example.com"},
					VerifyLoginIDKeys: []string{"email", "phone"},
				}), ShouldResemble, verifyResult{All: true, Any: true})

				So(isUserVerified(verifyRequest{
					LoginIDs:          map[string]string{"email": "a@example.com b@example.com", "phone": "+85299999999"},
					VerifiedLoginIDs:  []string{"a@example.com", "b@example.com"},
					VerifyLoginIDKeys: []string{"email", "phone"},
				}), ShouldResemble, verifyResult{All: false, Any: true})

				So(isUserVerified(verifyRequest{
					LoginIDs:          map[string]string{"email": "a@example.com b@example.com", "phone": "+85299999999"},
					VerifiedLoginIDs:  []string{"a@example.com", "b@example.com", "+85299999999"},
					VerifyLoginIDKeys: []string{"email", "phone"},
				}), ShouldResemble, verifyResult{All: true, Any: true})
			})
		})
	})
}
