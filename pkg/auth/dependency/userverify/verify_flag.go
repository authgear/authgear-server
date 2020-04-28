package userverify

import (
	"github.com/skygeario/skygear-server/pkg/auth/dependency/identity/loginid"
	"github.com/skygeario/skygear-server/pkg/core/config"
)

func IsUserVerified(
	verifyInfo map[string]bool,
	identities []*loginid.Identity,
	criteria config.UserVerificationCriteria,
	verifyConfigs []config.UserVerificationKeyConfiguration,
) (verified bool) {
	verified = false
	if len(verifyConfigs) == 0 {
		return
	}

	switch criteria {
	case config.UserVerificationCriteriaAll:
		// Login IDs to verify exist and all are verified
		loginIDToVerify := 0
		for _, principal := range identities {
			for _, c := range verifyConfigs {
				if principal.LoginIDKey != c.Key {
					continue
				}
				loginIDToVerify++
				if !verifyInfo[principal.LoginID] {
					verified = false
					return
				}
			}
		}
		verified = loginIDToVerify > 0

	case config.UserVerificationCriteriaAny:
		// Login IDs to verify exist and some are verified
		for _, principal := range identities {
			for _, c := range verifyConfigs {
				if principal.LoginIDKey != c.Key {
					continue
				}
				if verifyInfo[principal.LoginID] {
					verified = true
					return
				}
			}
		}
		verified = false

	default:
		panic("userverify: unexpected verify criteria: " + criteria)
	}
	return
}
